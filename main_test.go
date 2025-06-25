package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/oschwald/geoip2-golang"
)

func TestRunStartsServer(t *testing.T) {
	// create temp directory for config and credentials
	dir := t.TempDir()

	// create dummy credentials
	accFile := filepath.Join(dir, "acc")
	licFile := filepath.Join(dir, "lic")
	os.WriteFile(accFile, []byte("id"), 0o600)
	os.WriteFile(licFile, []byte("key"), 0o600)

	// create config
	cfg := `mode: "blocklist"
geoip_db_path: "db.mmdb"
listen_addr: ":1234"
maxmind_account_id_file: "` + accFile + `"
maxmind_license_key_file: "` + licFile + `"
maxmind_edition_id: "GeoLite2-Country"
update_check_interval_hours: 1
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600)

	// override working directory so run loads our config
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	calledDownload := false
	downloadGeoIPDBIfUpdated = func() { calledDownload = true }
	defer func() { downloadGeoIPDBIfUpdated = DownloadGeoIPDBIfUpdated }()

	var openPath string
	openGeoDBFn = func(path string) (*geoip2.Reader, error) {
		openPath = path
		return nil, nil
	}
	defer func() { openGeoDBFn = OpenGeoDB }()

	served := false
	listenAndServe = func(addr string, h http.Handler) error {
		served = true
		if addr != ":1234" {
			t.Fatalf("expected addr :1234, got %s", addr)
		}
		return nil
	}
	defer func() { listenAndServe = http.ListenAndServe }()

	periodicUpdaterFn = func() {}
	defer func() { periodicUpdaterFn = PeriodicUpdater }()

	if err := run(); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !calledDownload {
		t.Errorf("download function not called")
	}
	if openPath != "db.mmdb" {
		t.Errorf("openGeoDB path = %s", openPath)
	}
	if !served {
		t.Errorf("listenAndServe not called")
	}
}

func TestRunOpenGeoDBError(t *testing.T) {
	dir := t.TempDir()

	accFile := filepath.Join(dir, "acc")
	licFile := filepath.Join(dir, "lic")
	os.WriteFile(accFile, []byte("id"), 0o600)
	os.WriteFile(licFile, []byte("key"), 0o600)

	cfg := `mode: "blocklist"
geoip_db_path: "db.mmdb"
listen_addr: ":0"
maxmind_account_id_file: "` + accFile + `"
maxmind_license_key_file: "` + licFile + `"
maxmind_edition_id: "GeoLite2-Country"
update_check_interval_hours: 1
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600)

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	openGeoDBFn = func(path string) (*geoip2.Reader, error) { return nil, fmt.Errorf("bad") }
	defer func() { openGeoDBFn = OpenGeoDB }()

	periodicUpdaterFn = func() {}
	defer func() { periodicUpdaterFn = PeriodicUpdater }()

	listenAndServe = func(addr string, h http.Handler) error { return nil }
	defer func() { listenAndServe = http.ListenAndServe }()

	downloadGeoIPDBIfUpdated = func() {}
	defer func() { downloadGeoIPDBIfUpdated = DownloadGeoIPDBIfUpdated }()

	if err := run(); err == nil {
		t.Fatalf("expected error from run when openGeoDB fails")
	}
}
