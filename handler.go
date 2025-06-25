package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

func AuthzHandler(w http.ResponseWriter, r *http.Request) {
    ip := ExtractClientIP(r)
    parsedIP := net.ParseIP(ip)
    if parsedIP == nil {
        deny(w, "Invalid IP")
        return
    }

    if config.Debug {
        log.Printf("Request from IP: %s", ip)
    }

    if IsPrivateIP(parsedIP) && config.BlockPrivateIPs {
        deny(w, "Private IP blocked")
        return
    }

    countryCode := LookupCountry(parsedIP)
    if config.Debug {
        log.Printf("Resolved Country: %s", countryCode)
    }

    inList := StringInSlice(countryCode, config.Countries)
    if (config.Mode == "allowlist" && !inList) || (config.Mode == "blocklist" && inList) {
        deny(w, "Country policy blocked")
        return
    }

    allow(w)
}

func ExtractClientIP(r *http.Request) string {
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

func StringInSlice(val string, list []string) bool {
    for _, item := range list {
        if strings.EqualFold(val, item) {
            return true
        }
    }
    return false
}
