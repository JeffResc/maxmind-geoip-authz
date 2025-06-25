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
	periodicUpdaterFn          = periodicUpdater
	listenAndServe             = http.ListenAndServe
)

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func run() error {
	config = loadConfig("config.yaml")
	accountID, licenseKey = loadMaxMindCredentials(
		config.MaxMindAccountIDFile,
		config.MaxMindLicenseKeyFile,
	)

	downloadGeoIPDBIfUpdatedFn()

	var err error
	geoDB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	go periodicUpdaterFn()

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", authzHandler)
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
