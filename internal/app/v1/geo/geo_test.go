package geo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errInvalid stands in for the error geoip.Lookup returns on a bad IP.
var errInvalid = errors.New("invalid ip")

type fakeLookup struct {
	code string
	err  error
	last string
}

func (f *fakeLookup) CountryCode(ip string) (string, error) {
	f.last = ip
	return f.code, f.err
}

func newHandler(code string, err error) (http.Handler, *fakeLookup) {
	f := &fakeLookup{code: code, err: err}
	mux := http.NewServeMux()
	Register(mux, f)
	return mux, f
}

func doRequest(h http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandleGeo_success(t *testing.T) {
	h, f := newHandler("US", nil)
	rec := doRequest(h, "/geo?ip=8.8.8.8")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if f.last != "8.8.8.8" {
		t.Fatalf("lookup got %q, want 8.8.8.8", f.last)
	}
	var body struct {
		IP          string `json:"ip"`
		Type        string `json:"type"`
		CountryCode string `json:"country_code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.CountryCode != "US" {
		t.Fatalf("country_code = %q, want US", body.CountryCode)
	}
	if body.IP != "8.8.8.8" {
		t.Fatalf("ip = %q, want 8.8.8.8", body.IP)
	}
	if body.Type != "ipv4" {
		t.Fatalf("type = %q, want ipv4", body.Type)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}
}

func TestHandleGeo_ipv6(t *testing.T) {
	h, f := newHandler("US", nil)
	rec := doRequest(h, "/geo?ip=2001:4860:4860::8888")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if f.last != "2001:4860:4860::8888" {
		t.Fatalf("lookup got %q, want 2001:4860:4860::8888", f.last)
	}
	var body struct {
		IP   string `json:"ip"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.IP != "2001:4860:4860::8888" {
		t.Fatalf("ip = %q, want 2001:4860:4860::8888", body.IP)
	}
	if body.Type != "ipv6" {
		t.Fatalf("type = %q, want ipv6", body.Type)
	}
}

func TestHandleGeo_notFound(t *testing.T) {
	// Lookup succeeds but yields no country (e.g. private IP) -> 200 with empty code.
	h, _ := newHandler("", nil)
	rec := doRequest(h, "/geo?ip=192.168.1.1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		IP          string `json:"ip"`
		Type        string `json:"type"`
		CountryCode string `json:"country_code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.CountryCode != "" {
		t.Fatalf("country_code = %q, want empty", body.CountryCode)
	}
	if body.IP != "192.168.1.1" {
		t.Fatalf("ip = %q, want 192.168.1.1", body.IP)
	}
	if body.Type != "ipv4" {
		t.Fatalf("type = %q, want ipv4", body.Type)
	}
}

func TestHandleGeo_invalidIP(t *testing.T) {
	// Lookup returns an error (malformed IP) -> 400.
	h, _ := newHandler("", errInvalid)
	rec := doRequest(h, "/geo?ip=not-an-ip")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error != "invalid ip" {
		t.Fatalf("error = %q, want invalid ip", body.Error)
	}
}

func TestHandleGeo_methodNotAllowed(t *testing.T) {
	h, _ := newHandler("US", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/geo", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{
			name:   "query param wins",
			target: "/geo?ip=1.1.1.1",
			xff:    "2.2.2.2",
			xri:    "3.3.3.3",
			want:   "1.1.1.1",
		},
		{
			name:       "x-forwarded-for first entry",
			target:     "/geo",
			xff:        "203.0.113.5, 70.41.13.18",
			remoteAddr: "127.0.0.1:1234",
			want:       "203.0.113.5",
		},
		{
			name:       "x-forwarded-for single",
			target:     "/geo",
			xff:        "203.0.113.5",
			remoteAddr: "127.0.0.1:1234",
			want:       "203.0.113.5",
		},
		{
			name:       "x-real-ip",
			target:     "/geo",
			xri:        "198.51.100.7",
			remoteAddr: "127.0.0.1:1234",
			want:       "198.51.100.7",
		},
		{
			name:       "remote addr fallback",
			target:     "/geo",
			remoteAddr: "192.0.2.44:5678",
			want:       "192.0.2.44",
		},
		{
			name:       "remote addr without port",
			target:     "/geo",
			remoteAddr: "192.0.2.44",
			want:       "192.0.2.44",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}
			if got := clientIP(req); got != tt.want {
				t.Fatalf("clientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
