package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
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

type ASNRecord struct {
	AutonomousSystemNumber uint `maxminddb:"autonomous_system_number"`
}

func main() {
	// Create output directory
	if err := os.MkdirAll("output", 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Download and process ASN data from P3TERX (MMDB)
	fmt.Println("Downloading ASN mmdb data from P3TERX...")
	asnData, err := downloadAndParseMMDB()
	if err != nil {
		fmt.Printf("Error processing ASN data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ASN groups extracted: %d\n", len(asnData))

	// Generate asn.dat (only ASN numbers)
	fmt.Println("\nGenerating asn.dat...")
	if err := generateDatFile(asnData, "output/asn.dat"); err != nil {
		fmt.Printf("Error generating asn.dat: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated asn.dat (ASN numbers only)")

	// Generate text files and zip them
	fmt.Println("\nGenerating asn-text.zip...")
	if err := generateTextFilesZip(asnData, "output/asn-text.zip"); err != nil {
		fmt.Printf("Error generating asn-text.zip: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated asn-text.zip (text files with IP networks)")

	fmt.Println("\nDone!")
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

	// Sort keys for deterministic output
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
