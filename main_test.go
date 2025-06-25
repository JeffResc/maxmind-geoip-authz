package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/oschwald/geoip2-golang"
)

func TestServeStartsServer(t *testing.T) {
	// create temp directory for config
	dir := t.TempDir()

	// create config without MaxMind credentials
	cfg := `mode: "blocklist"
geoip_db_path: "db.mmdb"
listen_addr: ":1234"
maxmind_edition_id: "GeoLite2-Country"
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600)

	// override working directory so run loads our config
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	calledDownload := false
	downloadGeoIPDBIfUpdatedFn = func() { calledDownload = true }
	defer func() { downloadGeoIPDBIfUpdatedFn = downloadGeoIPDBIfUpdated }()

	var openPath string
	openGeoDBFn = func(path string) (*geoip2.Reader, error) {
		openPath = path
		return nil, nil
	}
	defer func() { openGeoDBFn = geoip.Open }()

	served := false
	listenAndServe = func(addr string, h http.Handler) error {
		served = true
		if addr != ":1234" {
			t.Fatalf("expected addr :1234, got %s", addr)
		}
		return nil
	}
	defer func() { listenAndServe = http.ListenAndServe }()

	if err := serve(); err != nil {
		t.Fatalf("serve returned error: %v", err)
	}

	if calledDownload {
		t.Errorf("download function should not be called")
	}
	if openPath != "db.mmdb" {
		t.Errorf("openGeoDB path = %s", openPath)
	}
	if !served {
		t.Errorf("listenAndServe not called")
	}
}

func TestServeOpenGeoDBError(t *testing.T) {
	dir := t.TempDir()

	cfg := `mode: "blocklist"
geoip_db_path: "db.mmdb"
listen_addr: ":0"
maxmind_edition_id: "GeoLite2-Country"
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600)

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	openGeoDBFn = func(path string) (*geoip2.Reader, error) { return nil, fmt.Errorf("bad") }
	defer func() { openGeoDBFn = geoip.Open }()

	listenAndServe = func(addr string, h http.Handler) error { return nil }
	defer func() { listenAndServe = http.ListenAndServe }()

	downloadGeoIPDBIfUpdatedFn = func() {}
	defer func() { downloadGeoIPDBIfUpdatedFn = downloadGeoIPDBIfUpdated }()

	if err := serve(); err == nil {
		t.Fatalf("expected error from run when openGeoDB fails")
	}
}
