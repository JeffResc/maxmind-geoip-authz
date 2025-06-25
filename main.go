package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	cfg "github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/jeffresc/maxmind-geoip-authz/handler"
)

// function variables so tests can stub behavior
var (
	openGeoDBFn    = geoip.Open
	listenAndServe = http.ListenAndServe
)

func loadConfig(path string) cfg.Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	var c cfg.Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	if c.Mode != "allowlist" && c.Mode != "blocklist" {
		log.Fatalf("Invalid mode: %s", c.Mode)
	}
	if c.PrivateIPAction == "" {
		c.PrivateIPAction = "deny"
	}
	if c.PrivateIPAction != "allow" && c.PrivateIPAction != "deny" {
		log.Fatalf("Invalid private_ip_action: %s", c.PrivateIPAction)
	}
	if c.UnknownAction == "" {
		if c.Mode == "allowlist" {
			c.UnknownAction = "deny"
		} else {
			c.UnknownAction = "allow"
		}
	}
	if c.UnknownAction != "allow" && c.UnknownAction != "deny" {
		log.Fatalf("Invalid unknown_action: %s", c.UnknownAction)
	}
	return c
}

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func serve() error {
	config := loadConfig("config.yaml")

	var err error
	geoip.DB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", handler.Authz(config))
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	if err := serve(); err != nil {
		log.Fatal(err)
	}
}
