package main

import (
	"log"
	"net/http"
)

func main() {
    config = LoadConfig("config.yaml")
    accountID, licenseKey = LoadMaxMindCredentials(
        config.MaxMindAccountIDFile,
        config.MaxMindLicenseKeyFile,
    )

    DownloadGeoIPDBIfUpdated()

    var err error
    geoDB, err = OpenGeoDB(config.GeoIPDBPath)
    if err != nil {
        log.Fatalf("Failed to open GeoIP DB: %v", err)
    }

    go PeriodicUpdater()

    if config.Debug {
        log.Printf("Starting server on %s", config.ListenAddr)
    }

    http.HandleFunc("/authz", AuthzHandler)
    log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}
