package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper RoundTripper to mock HTTP responses
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// sets http.DefaultClient.Transport to mock; returns restore function
func withHTTPMock(fn roundTripFunc) func() {
	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = fn
	return func() { http.DefaultClient.Transport = orig }
}

// testBody is an io.ReadCloser that records if Close was called.
type testBody struct{ closed *bool }

func (b testBody) Read(p []byte) (int, error) { return 0, io.EOF }
func (b testBody) Close() error               { *b.closed = true; return nil }

func TestDownloadGeoIPDBIfUpdated_NoUpdate(t *testing.T) {
	// ensure no leftover file
	os.Remove("/tmp/geoip.zip")

	tmpDir := t.TempDir()
	config = Config{GeoIPDBPath: filepath.Join(tmpDir, "db.mmdb"), MaxMindEditionID: "test"}
	licenseKey = "lic"

	// create existing DB file to trigger If-Modified-Since header
	os.WriteFile(config.GeoIPDBPath, []byte("old"), 0644)

	var headerSeen bool
	closed := false
	restore := withHTTPMock(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("If-Modified-Since") != "" {
			headerSeen = true
		}
		return &http.Response{StatusCode: http.StatusNotModified, Body: testBody{closed: &closed}, Header: make(http.Header)}, nil
	}))
	defer restore()

	DownloadGeoIPDBIfUpdated()

	if !headerSeen {
		t.Fatalf("If-Modified-Since header not set")
	}

	if _, err := os.Stat("/tmp/geoip.zip"); !os.IsNotExist(err) {
		t.Fatalf("zip file should not be created")
	}
	if !closed {
		t.Fatalf("response body not closed")
	}
}

func TestDownloadGeoIPDBIfUpdated_Downloads(t *testing.T) {
	os.Remove("/tmp/geoip.zip")

	tmpDir := t.TempDir()
	config = Config{GeoIPDBPath: filepath.Join(tmpDir, "db.mmdb"), MaxMindEditionID: "test"}
	licenseKey = "lic"

	mmdbContent := []byte("dummydb")
	restore := withHTTPMock(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		buf := new(bytes.Buffer)
		zw := zip.NewWriter(buf)
		w, _ := zw.Create("GeoLite2-Country.mmdb")
		w.Write(mmdbContent)
		zw.Close()
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(buf.Bytes())), Header: make(http.Header)}, nil
	}))
	defer restore()

	DownloadGeoIPDBIfUpdated()

	data, err := os.ReadFile(config.GeoIPDBPath)
	if err != nil {
		t.Fatalf("DB file not written: %v", err)
	}
	if !bytes.Equal(data, mmdbContent) {
		t.Fatalf("DB content mismatch")
	}
}

func TestExtractAndSwapDB_NoMMDB(t *testing.T) {
	tmpDir := t.TempDir()
	config = Config{GeoIPDBPath: filepath.Join(tmpDir, "db.mmdb")}

	zipPath := filepath.Join(tmpDir, "test.zip")
	buf, _ := os.Create(zipPath)
	zw := zip.NewWriter(buf)
	zw.Create("notdb.txt")
	zw.Close()
	buf.Close()

	extractAndSwapDB(zipPath)

	if _, err := os.Stat(config.GeoIPDBPath); !os.IsNotExist(err) {
		t.Fatalf("DB file should not exist")
	}
}

func TestPeriodicUpdater(t *testing.T) {
	calls := 0
	downloadGeoIPDBIfUpdated = func() { calls++ }
	defer func() { downloadGeoIPDBIfUpdated = DownloadGeoIPDBIfUpdated }()

	ch := make(chan time.Time)
	tickerFactory = func(d time.Duration) <-chan time.Time { return ch }
	defer func() { tickerFactory = func(d time.Duration) <-chan time.Time { return time.NewTicker(d).C } }()

	done := make(chan struct{})
	go func() { PeriodicUpdater(); close(done) }()

	ch <- time.Now()
	close(ch)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("PeriodicUpdater did not exit")
	}

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}
