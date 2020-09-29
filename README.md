# PowerDNS Exporter
Export [PowerDNS](https://www.powerdns.com/) service metrics for scraping by Prometheus.

## Supported PowerDNS services
* [Authoritative Server](https://www.powerdns.com/auth.html)

---

## Flags

Name | Description | Default
---- | ---- | ----
listen-address | Address to listen on for incoming connections. | `:9120`
api-url | PowerDNS service HTTP API base URL. | `http://localhost:8081/`
api-key | PowerDNS service HTTP API Key | `-`

The`api-url` flag format:

* PowerDNS 3.x: `http://<HOST>:<API-PORT>/`
* PowerDNS 4.x: `http://<HOST>:<API-PORT>/api/v1`

## Status
**Production (Released)**
## Up && Running
Download the most suitable executable from [releases](https://github.com/konradasb/pdns_exporter/releases) and run it with
```
./pdns_exporter <flags>
```
## Contributing
Pull Requests are welcome and appreciated. For more major changes, create an issue in this repository to discuss the changes before creating a Pull Request.
## License
Distributed under the MIT License. See LICENSE for more information.
## Contact
Konradas Bunikis - konradas.bunikis@zohomail.eu
