package config

// Config holds application settings loaded from YAML.

type Config struct {
	Mode            string   `yaml:"mode"`
	Countries       []string `yaml:"countries"`
	BlockPrivateIPs bool     `yaml:"block_private_ips"`
	UnknownAction   string   `yaml:"unknown_action"`
	GeoIPDBPath     string   `yaml:"geoip_db_path"`
	ListenAddr      string   `yaml:"listen_addr"`
	Debug           bool     `yaml:"debug"`
}
