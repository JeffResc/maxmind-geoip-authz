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

	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/oschwald/geoip2-golang"
)

func downloadGeoIPDBIfUpdated() {
	url := fmt.Sprintf(
		"https://download.maxmind.com/geoip/databases/%s/download?suffix=zip",
		config.MaxMindEditionID,
	)

	// Determine modification time of local DB
	var localMod time.Time
	if fi, err := os.Stat(config.GeoIPDBPath); err == nil {
		localMod = fi.ModTime()
	}

	// Check remote modification time using HEAD
	req, _ := http.NewRequest("HEAD", url, nil)
	req.SetBasicAuth(accountID, licenseKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if config.Debug {
			log.Printf("Unexpected status checking GeoIP DB: %s", resp.Status)
		}
		return
	}

	var remoteMod time.Time
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			remoteMod = t
		}
	}
	if !remoteMod.IsZero() && !remoteMod.After(localMod) {
		if config.Debug {
			log.Printf("GeoIP DB is up to date")
		}
		return
	}

	// Download new database
	tmpFile := "/tmp/geoip.zip"
	req, _ = http.NewRequest("GET", url, nil)
	req.SetBasicAuth(accountID, licenseKey)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if config.Debug {
			log.Printf("Download failed: %s", resp.Status)
		}
		return
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		if config.Debug {
			log.Printf("Failed to create temp file: %v", err)
		}
		return
	}
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

			geoip.DBLock.Lock()
			if geoip.DB != nil {
				geoip.DB.Close()
			}
			geoip.DB, _ = geoip2.Open(tmpDB)
			os.Rename(tmpDB, config.GeoIPDBPath)
			geoip.DBLock.Unlock()

			if config.Debug {
				log.Printf("GeoIP DB updated successfully")
			}
			return
		}
	}
	log.Printf("MMDB file not found in archive")
}
