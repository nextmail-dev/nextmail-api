# nextmail-api

Backend utility API for the nextmail app.

## Features

- `GET /api/v1/geo` - resolve a requester's IP to a country code, returning the IP and its type (ipv4/ipv6).
- `GET /health` - service health probe (also available at `GET /api/v1/health`).

## Project layout

The project follows a layered, versioned layout designed so that adding a new
API is a one-package, one-line change.

```text
cmd/server/main.go              entry point: config -> deps -> server
internal/
  config/                       env-based configuration
  platform/                     reusable infrastructure
    geoip/                      MaxMind GeoLite2 wrapper
    web/                        shared JSON response helpers
  middleware/                   cross-cutting HTTP middleware (logger, recover)
  server/                       HTTP server assembly + root router (versioning)
  app/                          the API surface, isolated per version
    v1/
      v1.go                     registers all v1 modules
      geo/                      GET /geo
      health/                   GET /health
scripts/download-geoip.sh       fetch + install the GeoIP database
```

### Adding a new endpoint

1. Create a package under `internal/app/v1/<feature>/` exposing a `Register`
   function that mounts its routes on a `*http.ServeMux`, taking only the
   dependencies it needs:

   ```go
   package feature

   func Register(mux *http.ServeMux, deps Deps) {
       mux.HandleFunc("GET /feature", handle)
   }
   ```

2. Add one line in [`internal/app/v1/v1.go`](internal/app/v1/v1.go) to wire it up.

That's it - no core changes. New versions (v2, ...) live under `internal/app/v2/`
and mount in [`internal/server/router.go`](internal/server/router.go).

### API versioning

Each version is an isolated sub-router mounted under its prefix in
`server/router.go`:

```go
v1Mux := http.NewServeMux()
v1.Register(v1Mux, v1.Deps{GeoIP: deps.GeoIP})
mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Mux))
```

A future v2 mounts the same way without touching v1.

## Prerequisites

- Go 1.22+
- A MaxMind GeoLite2-Country database (`.mmdb`). GeoLite2 is free but requires
  a MaxMind account.

### Get the GeoIP database

1. Sign up at <https://www.maxmind.com/en/geolite2/signup> and generate a
   license key.
2. Download and install the database:

   ```bash
   export GEOIP_ACCOUNT_ID=<your account id>
   export GEOIP_LICENSE_KEY=<your license key>
   bash scripts/download-geoip.sh
   ```

   This places the file at `data/GeoLite2-Country.mmdb`.

   Alternatively, download it manually from the MaxMind portal and put the
   `GeoLite2-Country.mmdb` file in `data/`.

## Configuration

| Env var         | Default                        | Description                       |
| --------------- | ------------------------------ | --------------------------------- |
| `PORT`          | `8080`                         | HTTP port to listen on            |
| `APP_ENV`       | `dev`                          | Deployment environment            |
| `GEOIP_DB_PATH` | `./data/GeoLite2-Country.mmdb` | Path to the GeoLite2-Country file |

## Run

```bash
go run ./cmd/server
```

## Usage

```bash
# Country for the requester's own IP
curl http://localhost:8080/api/v1/geo
# -> {"ip":"203.0.113.5","type":"ipv4","country_code":"US"}

# Look up an arbitrary IP (for testing)
curl 'http://localhost:8080/api/v1/geo?ip=8.8.8.8'
# -> {"ip":"8.8.8.8","type":"ipv4","country_code":"US"}

# IPv6 addresses are reported as type "ipv6"
curl 'http://localhost:8080/api/v1/geo?ip=2001:4860:4860::8888'
# -> {"ip":"2001:4860:4860::8888","type":"ipv6","country_code":"US"}

# Private / unknown ranges return an empty code
curl 'http://localhost:8080/api/v1/geo?ip=192.168.1.1'
# -> {"ip":"192.168.1.1","type":"ipv4","country_code":""}

# A malformed IP is rejected
curl 'http://localhost:8080/api/v1/geo?ip=not-an-ip'
# -> {"error":"invalid ip"}
```

### Proxy / forwarded headers

When deployed behind a reverse proxy, `X-Forwarded-For` (first entry) and
`X-Real-IP` are honored to recover the client IP. **Only trust these headers
when the server is behind a proxy you control** - otherwise a client can spoof
its apparent country by setting them directly.
