package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// newTicker creates a ticker that produces events at the given duration.
// It is overridden in tests to allow deterministic control of ticker events.
var newTicker = func(d time.Duration) <-chan time.Time { return time.Tick(d) }

func periodicUpdater() {
	ticker := newTicker(time.Duration(config.UpdateCheckIntervalHours) * time.Hour)
	for range ticker {
		downloadGeoIPDBIfUpdatedFn()
	}
}

func downloadGeoIPDBIfUpdated() {
	url := fmt.Sprintf(
		"https://download.maxmind.com/app/geoip_download?edition_id=%s&license_key=%s&suffix=zip",
		config.MaxMindEditionID, licenseKey,
	)

	tmpFile := "/tmp/geoip.zip"
	req, _ := http.NewRequest("GET", url, nil)

	if fi, err := os.Stat(config.GeoIPDBPath); err == nil {
		req.Header.Set("If-Modified-Since", fi.ModTime().UTC().Format(http.TimeFormat))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusNotModified {
		if config.Debug {
			log.Printf("GeoIP DB is up to date")
		}
		resp.Body.Close()
		return
	}
	defer resp.Body.Close()

	out, _ := os.Create(tmpFile)
	io.Copy(out, resp.Body)
	out.Close()

	extractAndSwapDB(tmpFile)
}

func extractAndSwapDB(zipPath string) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Printf("Failed to unzip DB: %v", err)
		return
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".mmdb") {
			rc, _ := f.Open()
			defer rc.Close()
			tmpDB := config.GeoIPDBPath + ".tmp"
			outFile, _ := os.Create(tmpDB)
			io.Copy(outFile, rc)
			outFile.Close()

			dbLock.Lock()
			if geoDB != nil {
				geoDB.Close()
			}
			geoDB, _ = geoip2.Open(tmpDB)
			os.Rename(tmpDB, config.GeoIPDBPath)
			dbLock.Unlock()

			if config.Debug {
				log.Printf("GeoIP DB updated successfully")
			}
			return
		}
	}
	log.Printf("MMDB file not found in archive")
}
