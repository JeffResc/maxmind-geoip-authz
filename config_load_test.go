package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	data := []byte(`
mode: allowlist
countries:
  - US
  - CA
private_ip_action: deny
unknown_action: deny
geoip_db_path: /path/to/db
listen_addr: :8080
debug: true
`)
	tmp, err := ioutil.TempFile("", "cfg*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	c := loadConfig(tmp.Name())
	if c.Mode != "allowlist" || c.PrivateIPAction != "deny" || c.ListenAddr != ":8080" {
		t.Fatalf("unexpected config: %#v", c)
	}
	if len(c.Countries) != 2 || c.Countries[0] != "US" || c.Countries[1] != "CA" {
		t.Fatalf("countries not parsed: %#v", c.Countries)
	}
	if c.GeoIPDBPath != "/path/to/db" || c.UnknownAction != "deny" {
		t.Fatalf("paths not parsed: %#v", c)
	}
}

func TestLoadConfigInvalidMode(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		data := []byte("mode: invalid")
		tmp, err := ioutil.TempFile("", "badcfg*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmp.Write(data); err != nil {
			t.Fatal(err)
		}
		tmp.Close()
		loadConfig(tmp.Name())
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadConfigInvalidMode")
	cmd.Env = append(os.Environ(), "TEST_FATAL=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected failure")
	}
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
}
