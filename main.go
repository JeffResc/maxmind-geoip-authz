package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// function variables so tests can stub behavior
var (
	downloadGeoIPDBIfUpdated = DownloadGeoIPDBIfUpdated
	openGeoDBFn              = OpenGeoDB
	listenAndServe           = http.ListenAndServe
)

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func serve() error {
	config = LoadConfig("config.yaml")
	accountID, licenseKey = LoadMaxMindCredentials(
		config.MaxMindAccountIDFile,
		config.MaxMindLicenseKeyFile,
	)

	var err error
	geoDB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", AuthzHandler)
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected subcommand: serve or update")
	}
	switch os.Args[1] {
	case "serve":
		if err := serve(); err != nil {
			log.Fatal(err)
		}
	case "update":
		if len(os.Args) < 3 || os.Args[2] != "database" {
			log.Fatal("usage: update database")
		}
		config = LoadConfig("config.yaml")
		accountID, licenseKey = LoadMaxMindCredentials(
			config.MaxMindAccountIDFile,
			config.MaxMindLicenseKeyFile,
		)
		downloadGeoIPDBIfUpdated()
	default:
		log.Fatalf("unknown subcommand %s", os.Args[1])
	}
}
