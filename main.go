package main

import (
	"fmt"
	"log"
	"net/http"

	cfg "github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/jeffresc/maxmind-geoip-authz/handler"
)

// function variables so tests can stub behavior
var (
	openGeoDBFn    = geoip.Open
	listenAndServe = http.ListenAndServe
	config         cfg.Config
	licenseKey     string
)

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func serve() error {
	config = cfg.Load("config.yaml")

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
	execute()
}
