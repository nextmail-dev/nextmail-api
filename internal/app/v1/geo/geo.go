// Package geo implements the GET /geo endpoint, resolving a requester's IP to
// an ISO 3166-1 alpha-2 country code.
package geo

import (
	"net"
	"net/http"
	"strings"

	"nextmail-api/internal/platform/web"
)

// Lookuper resolves an IP address to an ISO 3166-1 alpha-2 country code.
// *geoip.Lookup satisfies it; tests may provide a fake.
type Lookuper interface {
	CountryCode(ip string) (string, error)
}

// Register mounts the geo endpoints on the given mux.
func Register(mux *http.ServeMux, lookup Lookuper) {
	h := &handler{lookup: lookup}
	mux.HandleFunc("GET /geo", h.handleGeo)
}

type handler struct {
	lookup Lookuper
}

func (h *handler) handleGeo(w http.ResponseWriter, r *http.Request) {
	ipStr := clientIP(r)

	code, err := h.lookup.CountryCode(ipStr)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "invalid ip")
		return
	}
	web.WriteJSON(w, http.StatusOK, geoResponse{
		IP:          ipStr,
		Type:        ipType(ipStr),
		CountryCode: code,
	})
}

// geoResponse is the JSON body returned by GET /geo.
type geoResponse struct {
	IP          string `json:"ip"`
	Type        string `json:"type"`
	CountryCode string `json:"country_code"`
}

// ipType reports whether ip is an IPv4 or IPv6 address. It returns "" for a
// value that is not a valid IP literal.
func ipType(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	if parsed.To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

// clientIP determines the requester's IP address. It honors an optional ?ip=
// query parameter (useful for testing), then the standard proxy headers, and
// finally falls back to the direct connection address.
//
// Note: X-Forwarded-For / X-Real-IP are only trustworthy when the server sits
// behind a trusted proxy that sets them.
func clientIP(r *http.Request) string {
	if q := strings.TrimSpace(r.URL.Query().Get("ip")); q != "" {
		return q
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
