package geoip

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// DB is the currently opened GeoIP database.
var DB *geoip2.Reader

// DBLock guards access to DB.
var DBLock sync.RWMutex

// CountryLookup points to the function used to look up a country record in the
// database. It can be overridden in tests.
var CountryLookup = func(ip net.IP) (*geoip2.Country, error) {
	return DB.Country(ip)
}

// Open opens the GeoIP database at the provided path.
func Open(path string) (*geoip2.Reader, error) {
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

// IsPrivateIP returns true if the IP is in a private network range.
func IsPrivateIP(ip net.IP) bool {
	for _, cidr := range privateNets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// LookupCountry returns the ISO country code for the given IP.
func LookupCountry(ip net.IP) string {
	DBLock.RLock()
	defer DBLock.RUnlock()
	record, err := CountryLookup(ip)
	if err != nil || record.Country.IsoCode == "" {
		return "UNKNOWN"
	}
	return record.Country.IsoCode
}
