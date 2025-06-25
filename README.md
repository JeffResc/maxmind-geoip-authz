# MaxMind GeoIP Authorization

`maxmind-geoip-authz` is a small HTTP service that authorizes requests based on the
requester's IP address. It uses the [MaxMind](https://www.maxmind.com) GeoIP2
country database to determine the country associated with an IP and then
applies either an allowlist or blocklist policy.

The service exposes a single endpoint:

```
GET /authz
```

Requests are allowed or denied based on the configuration file. Responses are
returned as JSON and include either `{"status": "allowed"}` or
`{"status": "denied", "reason": "..."}`.

## Configuration

Configuration is loaded from `config.yaml` in the working directory. An example
file is included in this repository:

```yaml
mode: "blocklist"
countries:
  - "CN"
  - "RU"
block_private_ips: true
unknown_action: allow
geoip_db_path: "/app/GeoLite2-Country.mmdb"
listen_addr: ":8080"
debug: false
```

- **mode** – `allowlist` or `blocklist` to control how the `countries` list is
  interpreted.
- **countries** – list of ISO country codes.
- **block_private_ips** – when set, requests from private IP ranges are denied.
  IPv6 ranges such as `::1`, `fc00::/7`, and `fe80::/10` are also considered
  private.
- **unknown_action** – `allow` or `deny` requests when the country cannot be determined.
- **geoip_db_path** – path to the MaxMind GeoIP2 country database.
- **listen_addr** – address the HTTP server listens on.
- **debug** – enable verbose logging.

## Building and Running

To build the server directly with Go:

```bash
go build -o maxmind-geoip-authz main.go
./maxmind-geoip-authz
```

A `Dockerfile` is also provided which builds a static binary and runs it in a
minimal container:

```bash
docker build -t maxmind-geoip-authz .
docker run -p 8080:8080 maxmind-geoip-authz
```

## Testing

Unit tests can be executed with the standard Go tooling:

```bash
go test ./...
```

