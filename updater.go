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

func PeriodicUpdater() {
	ticker := newTicker(time.Duration(config.UpdateCheckIntervalHours) * time.Hour)
	for range ticker {
		downloadGeoIPDBIfUpdated()
	}
}

func DownloadGeoIPDBIfUpdated() {
	url := fmt.Sprintf(
		"https://download.maxmind.com/app/geoip_download?edition_id=%s&license_key=%s&suffix=zip",
		config.MaxMindEditionID, licenseKey,
	)

	tmpFile := "/tmp/geoip.zip"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	if fi, err := os.Stat(config.GeoIPDBPath); err == nil {
		req.Header.Set("If-Modified-Since", fi.ModTime().UTC().Format(http.TimeFormat))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to download GeoIP DB: %v", err)
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

	out, err := os.Create(tmpFile)
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("Failed to save GeoIP DB: %v", err)
		out.Close()
		return
	}
	if err := out.Close(); err != nil {
		log.Printf("Failed to close temp file: %v", err)
		return
	}

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
			rc, err := f.Open()
			if err != nil {
				log.Printf("Failed to open file in archive: %v", err)
				return
			}
			defer rc.Close()

			tmpDB := config.GeoIPDBPath + ".tmp"
			outFile, err := os.Create(tmpDB)
			if err != nil {
				log.Printf("Failed to create temp DB: %v", err)
				return
			}
			if _, err := io.Copy(outFile, rc); err != nil {
				log.Printf("Failed to write temp DB: %v", err)
				outFile.Close()
				return
			}
			if err := outFile.Close(); err != nil {
				log.Printf("Failed to close temp DB: %v", err)
				return
			}

			dbLock.Lock()
			if geoDB != nil {
				geoDB.Close()
			}
			newDB, err := geoip2.Open(tmpDB)
			if err != nil {
				log.Printf("Failed to open new DB: %v", err)
			} else {
				geoDB = newDB
			}
			if err := os.Rename(tmpDB, config.GeoIPDBPath); err != nil {
				dbLock.Unlock()
				log.Printf("Failed to move DB into place: %v", err)
				return
			}
			dbLock.Unlock()

			if config.Debug {
				log.Printf("GeoIP DB updated successfully")
			}
			return
		}
	}
	log.Printf("MMDB file not found in archive")
}
