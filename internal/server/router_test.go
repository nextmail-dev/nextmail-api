package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeGeo implements geo.Lookuper for router tests.
type fakeGeo struct {
	code string
	err  error
}

func (f fakeGeo) CountryCode(ip string) (string, error) {
	return f.code, f.err
}

func TestRouter_v1Geo_success(t *testing.T) {
	router := newRouter(Deps{GeoIP: fakeGeo{code: "US"}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/geo?ip=8.8.8.8", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		IP          string `json:"ip"`
		Type        string `json:"type"`
		CountryCode string `json:"country_code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.CountryCode != "US" {
		t.Fatalf("country_code = %q, want US", body.CountryCode)
	}
	if body.IP != "8.8.8.8" {
		t.Fatalf("ip = %q, want 8.8.8.8", body.IP)
	}
	if body.Type != "ipv4" {
		t.Fatalf("type = %q, want ipv4", body.Type)
	}
}

func TestRouter_v1Geo_notFound(t *testing.T) {
	router := newRouter(Deps{GeoIP: fakeGeo{code: ""}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/geo?ip=192.168.1.1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		CountryCode string `json:"country_code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.CountryCode != "" {
		t.Fatalf("country_code = %q, want empty", body.CountryCode)
	}
}

func TestRouter_v1Geo_methodNotAllowed(t *testing.T) {
	router := newRouter(Deps{GeoIP: fakeGeo{code: "US"}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/geo", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestRouter_v1Geo_unknownPath(t *testing.T) {
	router := newRouter(Deps{GeoIP: fakeGeo{code: "US"}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRouter_health(t *testing.T) {
	router := newRouter(Deps{})

	for _, path := range []string{"/health", "/api/v1/health"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", path, rec.Code)
		}
		var body struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body.Status != "ok" {
			t.Fatalf("%s: status body = %q, want ok", path, body.Status)
		}
	}
}
