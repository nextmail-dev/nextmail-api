// Package geoip wraps the MaxMind GeoLite2-Country database for IP-to-country
// lookups.
package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// Lookup answers IP-to-country queries against a GeoLite2-Country database.
type Lookup struct {
	db *geoip2.Reader
}

// Open opens the GeoLite2-Country database at path. The returned Lookup must be
// closed when no longer needed.
func Open(path string) (*Lookup, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open geoip database %q: %w", path, err)
	}
	return &Lookup{db: db}, nil
}

// Close releases the database resources.
func (l *Lookup) Close() error {
	return l.db.Close()
}

// CountryCode returns the ISO 3166-1 alpha-2 country code (e.g. "US") for the
// given IP. It returns "" without error when the IP cannot be geo-located
// (e.g. private or reserved ranges). An error is returned only when ipStr is
// not a valid IP address.
func (l *Lookup) CountryCode(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid ip: %q", ipStr)
	}

	record, err := l.db.Country(ip)
	if err != nil {
		return "", fmt.Errorf("lookup %q: %w", ipStr, err)
	}
	return record.Country.IsoCode, nil
}
