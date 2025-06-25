package main

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var geoDB *geoip2.Reader
var dbLock sync.RWMutex

// geoDBCountryLookup points to the function used to look up a country record in
// the GeoIP database. It calls geoDB.Country by default but can be overridden
// in tests.
var geoDBCountryLookup = func(ip net.IP) (*geoip2.Country, error) {
	return geoDB.Country(ip)
}

func openGeoDB(path string) (*geoip2.Reader, error) {
	return geoip2.Open(path)
}

var privateNets []*net.IPNet

func init() {
	blocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}
	for _, block := range blocks {
		_, cidr, err := net.ParseCIDR(block)
		if err == nil {
			privateNets = append(privateNets, cidr)
		}
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, cidr := range privateNets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func lookupCountry(ip net.IP) string {
	dbLock.RLock()
	defer dbLock.RUnlock()
	record, err := geoDBCountryLookup(ip)
	if err != nil || record.Country.IsoCode == "" {
		return "UNKNOWN"
	}
	return record.Country.IsoCode
}
