// Command server starts the nextmail HTTP API.
package main

import (
	"log"

	"nextmail-api/internal/config"
	"nextmail-api/internal/platform/geoip"
	"nextmail-api/internal/server"
)

func main() {
	cfg := config.Load()

	lookup, err := geoip.Open(cfg.GeoIPDBPath)
	if err != nil {
		log.Fatalf("geoip: %v", err)
	}
	defer lookup.Close()
	log.Printf("loaded geoip database from %s", cfg.GeoIPDBPath)

	srv := server.New(cfg, server.Deps{GeoIP: lookup})
	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
