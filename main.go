package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	cfg "github.com/jeffresc/maxmind-geoip-authz/config"
	"github.com/jeffresc/maxmind-geoip-authz/geoip"
	"github.com/jeffresc/maxmind-geoip-authz/handler"
)

// function variables so tests can stub behavior
var (
	openGeoDBFn    = geoip.Open
	listenAndServe = http.ListenAndServe
)

type Viper struct {
	configFile string
	raw        []byte
}

func New() *Viper { return &Viper{} }

func (v *Viper) SetConfigFile(path string) { v.configFile = path }

func (v *Viper) ReadInConfig() error {
	data, err := os.ReadFile(v.configFile)
	if err != nil {
		return err
	}
	v.raw = data
	return nil
}

func (v *Viper) Unmarshal(out interface{}) error { return yaml.Unmarshal(v.raw, out) }

func loadConfig(path string) cfg.Config {
	v := New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	var c cfg.Config
	if err := v.Unmarshal(&c); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	if c.Mode != "allowlist" && c.Mode != "blocklist" {
		log.Fatalf("Invalid mode: %s", c.Mode)
	}
	return c
}

// run initializes resources and starts the HTTP server. It is separated from
// main so tests can exercise the startup logic without exiting the process.
func serve() error {
  config := loadConfig("config.yaml")

	var err error
	geoip.DB, err = openGeoDBFn(config.GeoIPDBPath)
	if err != nil {
		return fmt.Errorf("Failed to open GeoIP DB: %v", err)
	}

	if config.Debug {
		log.Printf("Starting server on %s", config.ListenAddr)
	}

	http.HandleFunc("/authz", handler.Authz(config))
	return listenAndServe(config.ListenAddr, nil)
}

func main() {
	if err := serve(); err != nil {
		log.Fatal(err)
	}
}
