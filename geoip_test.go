package main

import (
	"errors"
	"net"
	"testing"

	"github.com/oschwald/geoip2-golang"
)

func TestOpenGeoDBFailure(t *testing.T) {
	if _, err := openGeoDB("/nonexistent/path.mmdb"); err == nil {
		t.Fatal("expected error for missing DB")
	}
}

func TestIsPrivateIP(t *testing.T) {
	cases := []struct {
		ip     string
		expect bool
	}{
		{"10.0.0.1", true},
		{"172.16.5.4", true},
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"8.8.8.8", false},
	}
	for _, c := range cases {
		if isPrivateIP(net.ParseIP(c.ip)) != c.expect {
			t.Fatalf("IsPrivateIP(%s) expected %v", c.ip, c.expect)
		}
	}
}

func TestLookupCountry(t *testing.T) {
	orig := geoDBCountryLookup
	defer func() { geoDBCountryLookup = orig }()

	geoDBCountryLookup = func(ip net.IP) (*geoip2.Country, error) {
		rec := &geoip2.Country{}
		rec.Country.IsoCode = "US"
		return rec, nil
	}

	if c := lookupCountry(net.ParseIP("1.2.3.4")); c != "US" {
		t.Fatalf("expected US, got %s", c)
	}
}

func TestLookupCountryUnknown(t *testing.T) {
	orig := geoDBCountryLookup
	defer func() { geoDBCountryLookup = orig }()

	geoDBCountryLookup = func(ip net.IP) (*geoip2.Country, error) {
		return &geoip2.Country{}, nil
	}

	if c := lookupCountry(net.ParseIP("1.2.3.4")); c != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN, got %s", c)
	}
}

func TestLookupCountryError(t *testing.T) {
	orig := geoDBCountryLookup
	defer func() { geoDBCountryLookup = orig }()

	geoDBCountryLookup = func(ip net.IP) (*geoip2.Country, error) {
		return nil, errors.New("bad")
	}

	if c := lookupCountry(net.ParseIP("1.2.3.4")); c != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN on error, got %s", c)
	}
}
