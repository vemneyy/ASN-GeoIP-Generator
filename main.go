package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"google.golang.org/protobuf/proto"

	"github.com/vemneyy/ASN-GeoIP-Generator/proto/geoip"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Minute,
}

const asnMMDBURL = "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"
const nroStatsURL = "https://ftp.ripe.net/pub/stats/ripencc/nro-stats/latest/nro-delegated-stats"

type ASNRecord struct {
	AutonomousSystemNumber uint `maxminddb:"autonomous_system_number"`
}

func main() {
	if err := os.MkdirAll("output", 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Downloading ASN mmdb data from P3TERX...")
	asnData, err := downloadAndParseMMDB()
	if err != nil {
		fmt.Printf("Error processing ASN data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ASN groups extracted: %d\n", len(asnData))

	fmt.Println("\nDownloading ASN-to-country mapping from NRO...")
	asnCountry, err := downloadASNCountryMap()
	if err != nil {
		fmt.Printf("  WARNING: could not build ASN-to-country map: %v\n", err)
		fmt.Println("  Country tags (AS-XX) are SKIPPED for this build.")
	} else {
		fmt.Printf("  ASN-to-country entries: %d\n", len(asnCountry))
		countryData := buildCountryData(asnData, asnCountry)
		fmt.Printf("  Country groups built: %d\n", len(countryData))
		for code, cidrs := range countryData {
			asnData[code] = cidrs
		}
	}

	fmt.Println("\nApplying custom aliases from aliases.json...")
	aliases, err := loadAliases("aliases.json")
	if err != nil {
		fmt.Printf("  Error loading aliases.json: %v\n", err)
	} else if len(aliases) > 0 {
		applyAliases(asnData, aliases)
	} else {
		fmt.Println("  No aliases.json found or file is empty, skipping.")
	}

	for _, cidrs := range asnData {
		sortCIDRs(cidrs)
	}

	fmt.Println("\nGenerating asn.dat...")
	if err := generateDatFile(asnData, "output/asn.dat"); err != nil {
		fmt.Printf("Error generating asn.dat: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated asn.dat (ASN numbers + country groups + aliases)")

	fmt.Println("\nGenerating asn-text.zip...")
	if err := generateTextFilesZip(asnData, "output/asn-text.zip"); err != nil {
		fmt.Printf("Error generating asn-text.zip: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated asn-text.zip (text files with IP networks)")

	fmt.Println("\nDone!")
}

func sortCIDRs(cidrs []*geoip.CIDR) {
	sort.Slice(cidrs, func(i, j int) bool {
		lenI := len(cidrs[i].Ip)
		lenJ := len(cidrs[j].Ip)
		
		if lenI != lenJ {
			return lenI < lenJ
		}
		
		cmp := bytes.Compare(cidrs[i].Ip, cidrs[j].Ip)
		if cmp != 0 {
			return cmp < 0
		}
		
		return cidrs[i].Prefix < cidrs[j].Prefix
	})
}

func loadAliases(filename string) (map[string][]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var aliases map[string][]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, fmt.Errorf("invalid json format: %w", err)
	}
	return aliases, nil
}

func applyAliases(asnData map[string][]*geoip.CIDR, aliases map[string][]string) {
	for aliasName, asns := range aliases {
		var combined []*geoip.CIDR

		for _, asn := range asns {
			asnKey := strings.ToUpper(strings.TrimSpace(asn))
			if !strings.HasPrefix(asnKey, "AS") {
				asnKey = "AS" + asnKey
			}

			if cidrs, ok := asnData[asnKey]; ok {
				combined = append(combined, cidrs...)
			} else {
				fmt.Printf("    WARNING: ASN %s not found for alias '%s'\n", asnKey, aliasName)
			}
		}

		if len(combined) > 0 {
			asnData[aliasName] = combined
			fmt.Printf("  Created alias: '%s' (merged %d ASNs, %d CIDR blocks)\n", aliasName, len(asns), len(combined))
		}
	}
}

func downloadAndParseMMDB() (map[string][]*geoip.CIDR, error) {
	fmt.Printf("  Downloading %s...\n", asnMMDBURL)

	resp, err := httpClient.Get(asnMMDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", asnMMDBURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: status %d", asnMMDBURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	db, err := maxminddb.FromBytes(body)
	if err != nil {
		return nil, fmt.Errorf("failed to open mmdb: %w", err)
	}
	defer db.Close()

	asnData := make(map[string][]*geoip.CIDR)

	networks := db.Networks(maxminddb.SkipAliasedNetworks)
	var record ASNRecord

	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			continue
		}

		if record.AutonomousSystemNumber == 0 {
			continue
		}

		ones, _ := subnet.Mask.Size()

		ipBytes := subnet.IP
		if v4 := subnet.IP.To4(); v4 != nil {
			ipBytes = v4
		}

		cidr := &geoip.CIDR{
			Ip:     ipBytes,
			Prefix: uint32(ones),
		}

		asnKey := fmt.Sprintf("AS%d", record.AutonomousSystemNumber)
		asnData[asnKey] = append(asnData[asnKey], cidr)
	}

	if err := networks.Err(); err != nil {
		return nil, fmt.Errorf("error during mmdb iteration: %w", err)
	}

	return asnData, nil
}

func downloadASNCountryMap() (map[uint32]string, error) {
	fmt.Printf("  Downloading %s...\n", nroStatsURL)

	resp, err := httpClient.Get(nroStatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", nroStatsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: status %d", nroStatsURL, resp.StatusCode)
	}

	result := make(map[uint32]string)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}

		if fields[2] != "asn" {
			continue
		}

		status := fields[6]
		if status != "allocated" && status != "assigned" {
			continue
		}

		cc := strings.ToUpper(strings.TrimSpace(fields[1]))
		if !isCountryCode(cc) {
			continue
		}

		start, err := strconv.ParseUint(fields[3], 10, 32)
		if err != nil {
			continue
		}

		count, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil || count == 0 {
			continue
		}

		for i := uint64(0); i < count; i++ {
			asn := start + i
			if asn > math.MaxUint32 {
				break
			}
			result[uint32(asn)] = cc
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading NRO stats: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("parsed 0 ASN-to-country records (unexpected format?)")
	}

	return result, nil
}

func isCountryCode(s string) bool {
	if len(s) != 2 || s == "ZZ" {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func buildCountryData(asnData map[string][]*geoip.CIDR, asnCountry map[uint32]string) map[string][]*geoip.CIDR {
	countryData := make(map[string][]*geoip.CIDR)

	keys := make([]string, 0, len(asnData))
	for k := range asnData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, asnKey := range keys {
		numStr := strings.TrimPrefix(asnKey, "AS")
		num, err := strconv.ParseUint(numStr, 10, 32)
		if err != nil {
			continue
		}

		cc, ok := asnCountry[uint32(num)]
		if !ok {
			continue
		}

		countryKey := "AS-" + cc
		countryData[countryKey] = append(countryData[countryKey], asnData[asnKey]...)
	}

	return countryData
}

func generateDatFile(data map[string][]*geoip.CIDR, outputPath string) error {
	list := &geoip.GeoIPList{}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		cidrs := data[name]
		entry := &geoip.GeoIP{
			CountryCode: name,
			Cidr:        cidrs,
		}
		list.Entry = append(list.Entry, entry)
	}

	out, err := proto.Marshal(list)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, out, 0644)
}

func generateTextFilesZip(data map[string][]*geoip.CIDR, outputPath string) error {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		cidrs := data[name]

		filename := strings.ToLower(name) + ".txt"

		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		for _, cidr := range cidrs {
			ip := net.IP(cidr.Ip)
			line := fmt.Sprintf("%s/%d\n", ip.String(), cidr.Prefix)
			if _, err := writer.Write([]byte(line)); err != nil {
				return err
			}
		}
	}

	return nil
}
