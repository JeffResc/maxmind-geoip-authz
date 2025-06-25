package main

import (
	"fmt"
	"log"
	"net/http"
)

// function variables so tests can stub behavior
var (
	downloadGeoIPDBIfUpdated = DownloadGeoIPDBIfUpdated
	openGeoDBFn              = OpenGeoDB
	periodicUpdaterFn        = PeriodicUpdater
	listenAndServe           = http.ListenAndServe
)

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func run() error {
	config = LoadConfig("config.yaml")
	accountID, licenseKey = LoadMaxMindCredentials(
		config.MaxMindAccountIDFile,
		config.MaxMindLicenseKeyFile,
	)

	downloadGeoIPDBIfUpdated()

	var err error
	geoDB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	go periodicUpdaterFn()

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", AuthzHandler)
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
