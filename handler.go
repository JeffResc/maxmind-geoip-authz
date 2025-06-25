package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

// lookupCountryFn is used by authzHandler to determine the requester's country.
// It points to lookupCountry by default but can be overridden in tests.
var lookupCountryFn = lookupCountry

func authzHandler(w http.ResponseWriter, r *http.Request) {
	ip := extractClientIP(r)
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		deny(w, "Invalid IP")
		return
	}

	if config.Debug {
		log.Printf("Request from IP: %s", ip)
	}

	if isPrivateIP(parsedIP) && config.BlockPrivateIPs {
		deny(w, "Private IP blocked")
		return
	}

	countryCode := lookupCountryFn(parsedIP)
	if config.Debug {
		log.Printf("Resolved Country: %s", countryCode)
	}

	inList := stringInSlice(countryCode, config.Countries)
	if (config.Mode == "allowlist" && !inList) || (config.Mode == "blocklist" && inList) {
		deny(w, "Country policy blocked")
		return
	}

	allow(w)
}

func extractClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func deny(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{"status": "denied", "reason": msg})
}

func allow(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "allowed"})
}

func stringInSlice(val string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(val, item) {
			return true
		}
	}
	return false
}
