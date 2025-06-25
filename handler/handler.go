package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/jeffresc/maxmind-geoip-authz/geoip"
)

// LookupCountryFn is used by Authz to determine the requester's country. It
// points to geoip.LookupCountry by default but can be overridden in tests.
var LookupCountryFn = geoip.LookupCountry

// Authz returns an HTTP handler that authorizes requests using the provided
// configuration.
func Authz(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r)
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			deny(w, "Invalid IP")
			return
		}

		if cfg.Debug {
			log.Printf("Request from IP: %s", ip)
		}

		if geoip.IsPrivateIP(parsedIP) && cfg.BlockPrivateIPs {
			deny(w, "Private IP blocked")
			return
		}

		countryCode := LookupCountryFn(parsedIP)
		if cfg.Debug {
			log.Printf("Resolved Country: %s", countryCode)
		}

		inList := stringInSlice(countryCode, cfg.Countries)
		if (cfg.Mode == "allowlist" && !inList) || (cfg.Mode == "blocklist" && inList) {
			deny(w, "Country policy blocked")
			return
		}

		allow(w)
	}
}

func extractClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		if idx := strings.IndexByte(forwarded, ','); idx != -1 {
			forwarded = forwarded[:idx]
		}
		return strings.TrimSpace(forwarded)
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
