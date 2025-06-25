package main

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Mode                  string   `yaml:"mode"`
	Countries             []string `yaml:"countries"`
	BlockPrivateIPs       bool     `yaml:"block_private_ips"`
	GeoIPDBPath           string   `yaml:"geoip_db_path"`
	ListenAddr            string   `yaml:"listen_addr"`
	Debug                 bool     `yaml:"debug"`
	MaxMindAccountIDFile  string   `yaml:"maxmind_account_id_file"`
	MaxMindLicenseKeyFile string   `yaml:"maxmind_license_key_file"`
	MaxMindEditionID      string   `yaml:"maxmind_edition_id"`
}

var config Config
var accountID, licenseKey string

func LoadConfig(path string) Config {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	if c.Mode != "allowlist" && c.Mode != "blocklist" {
		log.Fatalf("Invalid mode: %s", c.Mode)
	}
	return c
}

func LoadMaxMindCredentials(accountPath, licensePath string) (string, string) {
	accData, err := os.ReadFile(accountPath)
	if err != nil {
		log.Fatalf("Failed to read MaxMind Account ID: %v", err)
	}
	licData, err := os.ReadFile(licensePath)
	if err != nil {
		log.Fatalf("Failed to read MaxMind License Key: %v", err)
	}
	accountID := strings.TrimSpace(string(accData))
	licenseKey := strings.TrimSpace(string(licData))
	if accountID == "" || licenseKey == "" {
		log.Fatalf("MaxMind credentials incomplete")
	}
	return accountID, licenseKey
}
