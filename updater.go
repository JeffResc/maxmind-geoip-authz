package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/oschwald/geoip2-golang"
)

func downloadGeoIPDBIfUpdated() error {
	url := fmt.Sprintf(
		"https://download.maxmind.com/app/geoip_download?edition_id=%s&license_key=%s&suffix=zip",
		config.MaxMindEditionID, licenseKey,
	)

	tmpFile := "/tmp/geoip.zip"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(config.GeoIPDBPath); err == nil {
		req.Header.Set("If-Modified-Since", fi.ModTime().UTC().Format(http.TimeFormat))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotModified {
		if config.Debug {
			log.Printf("GeoIP DB is up to date")
		}
		resp.Body.Close()
		return nil
	}
	defer resp.Body.Close()

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	return extractAndSwapDB(tmpFile)
}

func extractAndSwapDB(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("Failed to unzip DB: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".mmdb") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			tmpDB := config.GeoIPDBPath + ".tmp"
			outFile, err := os.Create(tmpDB)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}

			geoip.DBLock.Lock()
			if geoip.DB != nil {
				geoip.DB.Close()
			}
			db, err := geoip2.Open(tmpDB)
			if err != nil {
				geoip.DBLock.Unlock()
				return err
			}
			geoip.DB = db
			if err := os.Rename(tmpDB, config.GeoIPDBPath); err != nil {
				geoip.DBLock.Unlock()
				return err
			}
			geoip.DBLock.Unlock()

			if config.Debug {
				log.Printf("GeoIP DB updated successfully")
			}
			return nil
		}
	}
	return fmt.Errorf("MMDB file not found in archive")
}
