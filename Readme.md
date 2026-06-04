# ASN GeoIP Generator

This project automatically generates an `asn.dat` file every 72 hours containing **Autonomous System Numbers (ASN)**, plus **per-country ASN groups**, for use in various proxy and VPN applications.

## Data Source

- [P3TERX/GeoLite.mmdb](https://github.com/P3TERX/GeoLite.mmdb) - GeoLite2 ASN MMDB database (based on MaxMind GeoLite2) — provides ASN → IP networks.
- [NRO delegated statistics](https://ftp.ripe.net/pub/stats/ripencc/nro-stats/latest/nro-delegated-stats) - the combined RIR (ARIN/RIPE/APNIC/LACNIC/AFRINIC) registry that maps every allocated/assigned ASN to the country it is registered/delegated to. This is required because the GeoLite2-ASN database itself carries **no** country information.

## Output Formats

- `asn.dat` - A V2Ray/Xray-compatible database file containing both per-ASN data and per-country ASN groups.
- `asn-text.zip` - A ZIP archive containing plain text files with IP subnets listed on each line:
    - per-ASN files: `as1.txt`, `as13335.txt`, …
    - per-country files: `as-ru.txt`, `as-us.txt`, …

## Using AS Numbers

Route traffic to IP addresses belonging to specific autonomous systems by their AS numbers:

```
asn.dat:AS13238
asn.dat:AS15169
asn.dat:AS13335
```

## Using Country ASN Groups

Each country group (`AS-XX`, where `XX` is the ISO 3166-1 alpha-2 code) contains the IP networks of **every** ASN registered/delegated to that country, combined into a single tag:

```
asn.dat:AS-RU
asn.dat:AS-US
asn.dat:AS-DE
```

> Note: a handful of non-country region codes used by the registries may also appear — `AS-EU` (RIPE multi-country) and `AS-AP` (APNIC region). Codes marked unknown (`ZZ`) are skipped.

**V2Ray/Xray Configuration Example:**

```json
{
  "routing": {
    "rules": [
      {
        "type": "field",
        "outboundTag": "direct",
        "ip": [
          "ext:asn.dat:AS13238",
          "ext:asn.dat:AS44534",
          "ext:asn.dat:AS-RU"
        ]
      }
    ]
  }
}
```

### Popular AS Numbers

| AS Number | Organization |
|-----------|--------------|
| AS13238 | Yandex LLC |
| AS44534 | Yandex.Cloud LLC |
| AS200350 | Yandex.Cloud LLC |
| AS15169 | Google LLC |
| AS13335 | Cloudflare, Inc. |
| AS32934 | Meta Platforms, Inc. (Facebook) |
| AS16509 | Amazon.com, Inc. (AWS) |
| AS62041 | Telegram Messenger Inc |
| AS59930 | Telegram Messenger LLP |

### Country Group Examples

| Tag | Meaning |
|-----|---------|
| AS-RU | All ASNs registered to Russia |
| AS-US | All ASNs registered to the United States |
| AS-DE | All ASNs registered to Germany |
| AS-CN | All ASNs registered to China |

## Download

The latest version of the files is always available at the following links:

- **asn.dat**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn.dat](https://raw.githubusercontent.com/vemneyy/ASN-GeoIP-Generator/release/asn.dat)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn.dat](https://cdn.jsdelivr.net/gh/vemneyy/ASN-GeoIP-Generator@release/asn.dat)

- **asn-text.zip**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn-text.zip](https://raw.githubusercontent.com/vemneyy/ASN-GeoIP-Generator/release/asn-text.zip)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn-text.zip](https://cdn.jsdelivr.net/gh/vemneyy/ASN-GeoIP-Generator@release/asn-text.zip)

## Build

To compile and run the generator locally:

```bash
go build -o asn-generator ./
./asn-generator
```

## Credits

- [P3TERX/GeoLite.mmdb](https://github.com/P3TERX/GeoLite.mmdb) - ASN database source
- [Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip) - Database format reference
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) - Original ASN database provider
- [NRO / RIR delegated statistics](https://www.nro.net/about/rirs/statistics/) - ASN → country mapping
