package main

import (
	"fmt"
	"log"
	"net/http"
)

// function variables so tests can stub behavior
var (
	downloadGeoIPDBIfUpdatedFn = downloadGeoIPDBIfUpdated
	openGeoDBFn                = openGeoDB
	listenAndServe             = http.ListenAndServe
)

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func serve() error {
	config = loadConfig("config.yaml")

	var err error
	geoDB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", authzHandler)
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	execute()
}
