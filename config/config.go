package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Mode            string   `yaml:"mode"`
	Countries       []string `yaml:"countries"`
	BlockPrivateIPs bool     `yaml:"block_private_ips"`
	GeoIPDBPath     string   `yaml:"geoip_db_path"`
	ListenAddr      string   `yaml:"listen_addr"`
	Debug           bool     `yaml:"debug"`
}

// Load reads configuration from the given YAML file and returns a Config.
func Load(path string) Config {
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
