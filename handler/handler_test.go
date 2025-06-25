package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	cfg "github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/jeffresc/maxmind-geoip-authz/geoip"
)

// testLookup maps IP string to country code used by fake lookup
var testLookup map[string]string

func fakeLookup(ip net.IP) string {
	if c, ok := testLookup[ip.String()]; ok {
		return c
	}
	return "UNKNOWN"
}

// helper to run a single request and return status and body map
func runRequest(c cfg.Config, req *http.Request) (int, map[string]string) {
	rr := httptest.NewRecorder()
	Authz(c)(rr, req)
	var body map[string]string
	json.Unmarshal(rr.Body.Bytes(), &body)
	return rr.Code, body
}

func TestInvalidIP(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "not_an_ip")
	status, body := runRequest(config, req)

	if status != http.StatusForbidden || body["reason"] != "Invalid IP" {
		t.Fatalf("expected invalid IP denial, got %v %#v", status, body)
	}
}

func TestPrivateIPBlocked(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", AllowPrivateIPs: false}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	status, body := runRequest(config, req)

	if status != http.StatusForbidden || body["reason"] != "Private IP blocked" {
		t.Fatalf("expected private IP denial, got %v %#v", status, body)
	}
}

func TestPrivateIPAllowed(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", AllowPrivateIPs: true}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed, got %v %#v", status, body)
	}
}

func TestAllowlist(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "allowlist", Countries: []string{"US"}}
	testLookup = map[string]string{"1.1.1.1": "US"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed, got %v %#v", status, body)
	}
}

func TestAllowlistDenied(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "allowlist", Countries: []string{"US"}}
	testLookup = map[string]string{"2.2.2.2": "FR"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "2.2.2.2")
	status, body := runRequest(config, req)

	if status != http.StatusForbidden || body["reason"] != "Country policy blocked" {
		t.Fatalf("expected allowlist denial, got %v %#v", status, body)
	}
}

func TestBlocklist(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", Countries: []string{"FR"}}
	testLookup = map[string]string{"2.2.2.2": "FR"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "2.2.2.2")
	status, body := runRequest(config, req)

	if status != http.StatusForbidden || body["reason"] != "Country policy blocked" {
		t.Fatalf("expected blocklist denial, got %v %#v", status, body)
	}
}

func TestBlocklistAllowed(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", Countries: []string{"FR"}}
	testLookup = map[string]string{"1.1.1.1": "US"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed, got %v %#v", status, body)
	}
}

func TestUnknownCountryAllowed(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", Countries: []string{"US"}, UnknownAction: "allow"}
	testLookup = map[string]string{}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed for unknown country, got %v %#v", status, body)
	}
}

func TestUnknownCountryDenied(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist", Countries: []string{"US"}, UnknownAction: "deny"}
	testLookup = map[string]string{}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	status, body := runRequest(config, req)

	if status != http.StatusForbidden || body["reason"] != "Unknown country" {
		t.Fatalf("expected denial for unknown country, got %v %#v", status, body)
	}
}

func TestMultipleForwardedFor(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist"}
	testLookup = map[string]string{"5.5.5.5": "US"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.Header.Set("X-Forwarded-For", "5.5.5.5, 6.6.6.6")
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed using first forwarded IP, got %v %#v", status, body)
	}
}

func TestRemoteAddrUsed(t *testing.T) {
	LookupCountryFn = fakeLookup
	defer func() { LookupCountryFn = geoip.LookupCountry }()

	config := cfg.Config{Mode: "blocklist"}
	testLookup = map[string]string{"7.7.7.7": "US"}

	req := httptest.NewRequest("GET", "/authz", nil)
	req.RemoteAddr = "7.7.7.7:1234"
	status, body := runRequest(config, req)

	if status != http.StatusOK || body["status"] != "allowed" {
		t.Fatalf("expected allowed using remote addr, got %v %#v", status, body)
	}
}
