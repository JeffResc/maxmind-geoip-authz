package viper

import (
	"os"

	"gopkg.in/yaml.v2"
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

func (v *Viper) Unmarshal(out interface{}) error {
	return yaml.Unmarshal(v.raw, out)
}
