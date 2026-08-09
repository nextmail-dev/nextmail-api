package geoip

import "testing"

func TestOpen_missingFile(t *testing.T) {
	_, err := Open("does-not-exist.mmdb")
	if err == nil {
		t.Fatal("expected an error when the database file is missing")
	}
}
