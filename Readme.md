# ASN GeoIP Generator

This project automatically generates an `asn.dat` file every 6 hours containing only **Autonomous System Numbers (ASN)** for use in various proxy and VPN applications.

## Data Source

- [P3TERX/GeoLite.mmdb](https://github.com/P3TERX/GeoLite.mmdb) - GeoLite2 ASN MMDB database (based on MaxMind GeoLite2)

## Output Formats

- `asn.dat` - A V2Ray/Xray-compatible database file containing only ASN data.
- `asn-text.zip` - A ZIP archive containing plain text files (`as1.txt`, `as10001.txt`, etc.) with IP subnets listed on each line.

## Using AS Numbers

This allows routing traffic to IP addresses belonging to specific autonomous systems by using their AS numbers:

```
asn.dat:AS13238
asn.dat:AS15169
asn.dat:AS13335
```

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
          "ext:asn.dat:AS44534"
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

## Download

The latest version of the files is always available at the following links:

- **asn.dat**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn.dat](https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn.dat)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn.dat](https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn.dat)

- **asn-text.zip**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn-text.zip](https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn-text.zip)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn-text.zip](https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn-text.zip)

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
