package main

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var geoDB *geoip2.Reader
var dbLock sync.RWMutex

func OpenGeoDB(path string) (*geoip2.Reader, error) {
    return geoip2.Open(path)
}

func IsPrivateIP(ip net.IP) bool {
	// https://datatracker.ietf.org/doc/html/rfc1918
    privateBlocks := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "127.0.0.0/8",
    }
    for _, block := range privateBlocks {
        _, cidr, _ := net.ParseCIDR(block)
        if cidr.Contains(ip) {
            return true
        }
    }
    return false
}

func LookupCountry(ip net.IP) string {
    dbLock.RLock()
    defer dbLock.RUnlock()
    record, err := geoDB.Country(ip)
    if err != nil || record.Country.IsoCode == "" {
        return "UNKNOWN"
    }
    return record.Country.IsoCode
}
