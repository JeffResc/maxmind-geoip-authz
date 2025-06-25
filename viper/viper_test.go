package viper

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Foo string `yaml:"foo"`
	Num int    `yaml:"num"`
}

func TestSetConfigFile(t *testing.T) {
	v := New()
	v.SetConfigFile("/tmp/test.yaml")
	if v.configFile != "/tmp/test.yaml" {
		t.Fatalf("configFile not set: %s", v.configFile)
	}
}

func TestReadInConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := []byte("foo: bar\nnum: 42")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	v := New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig error: %v", err)
	}
	if string(v.raw) != string(content) {
		t.Fatalf("raw content mismatch: %s", string(v.raw))
	}
}

func TestReadInConfigError(t *testing.T) {
	v := New()
	v.SetConfigFile("/nonexistent/file")
	if err := v.ReadInConfig(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestUnmarshal(t *testing.T) {
	v := &Viper{raw: []byte("foo: bar\nnum: 7")}
	var cfg testConfig
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if cfg.Foo != "bar" || cfg.Num != 7 {
		t.Fatalf("unexpected result: %#v", cfg)
	}
}
