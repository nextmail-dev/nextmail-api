// Package config holds runtime configuration loaded from the environment.
package config

import "os"

// Config is the application configuration.
type Config struct {
	// Port is the HTTP port the server listens on (default "8080").
	Port string
	// Env is the deployment environment: "dev" (default) or "prod".
	Env string
	// GeoIPDBPath is the path to the MaxMind GeoLite2-Country .mmdb file
	// (default "./data/GeoLite2-Country.mmdb").
	GeoIPDBPath string
}

// Load reads configuration from environment variables, applying defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	dbPath := os.Getenv("GEOIP_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/GeoLite2-Country.mmdb"
	}

	return Config{
		Port:        port,
		Env:         env,
		GeoIPDBPath: dbPath,
	}
}
