# ASN GeoIP Generator

Этот проект каждые 6 часов автоматически генерирует файл `asn.dat`, содержащий **только номера ASN (Autonomous Systems)** для использования в различных Proxy/VPN приложениях.

## Источник данных

- [Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip) - GeoLite2 ASN CSV файлы (на основе MaxMind GeoLite2)

## Выходной формат

- `asn.dat` - V2Ray/Xray-совместимый файл, содержащий только ASN номера
- `asn-text.zip` - ZIP-архив с текстовыми файлами (as1.txt, as10001.txt, ...) содержащими IP сети на каждой строке

## Использование AS номеров

Возможность обращаться к IP-адресам любой автономной системы по её AS номеру:

```
asn.dat:AS13238
asn.dat:AS15169
asn.dat:AS13335
```

**Пример конфигурации V2Ray/Xray:**

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

### Популярные AS номера

| AS номер | Организация |
|----------|-------------|
| AS13238 | Yandex LLC |
| AS44534 | Yandex.Cloud LLC |
| AS200350 | Yandex.Cloud LLC |
| AS15169 | Google LLC |
| AS13335 | Cloudflare, Inc. |
| AS32934 | Meta Platforms, Inc. (Facebook) |
| AS16509 | Amazon.com, Inc. (AWS) |
| AS62041 | Telegram Messenger Inc |
| AS59930 | Telegram Messenger LLP |

## Скачать

По данному адресу всегда доступна последняя версия файлов:

- **asn.dat**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn.dat](https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn.dat)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn.dat](https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn.dat)

- **asn-text.zip**
    - [https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn-text.zip](https://raw.githubusercontent.com/vemneyy/russia-blocked-geoip/release/asn-text.zip)
    - [https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn-text.zip](https://cdn.jsdelivr.net/gh/vemneyy/russia-blocked-geoip@release/asn-text.zip)

## Сборка

```bash
go build -o asn-generator ./
./asn-generator
```

## Credits

- [Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip) - источник ASN данных
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) - оригинальная база данных ASN
