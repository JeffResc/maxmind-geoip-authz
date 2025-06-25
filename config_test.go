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
block_private_ips: true
geoip_db_path: /path/to/db
listen_addr: :8080
debug: true
maxmind_account_id_file: acc
maxmind_license_key_file: lic
maxmind_edition_id: GeoLite2-Country
update_check_interval_hours: 12
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

	c := LoadConfig(tmp.Name())
	if c.Mode != "allowlist" || !c.BlockPrivateIPs || c.ListenAddr != ":8080" {
		t.Fatalf("unexpected config: %#v", c)
	}
	if len(c.Countries) != 2 || c.Countries[0] != "US" || c.Countries[1] != "CA" {
		t.Fatalf("countries not parsed: %#v", c.Countries)
	}
	if c.GeoIPDBPath != "/path/to/db" || c.MaxMindEditionID != "GeoLite2-Country" {
		t.Fatalf("paths not parsed: %#v", c)
	}
	if c.UpdateCheckIntervalHours != 12 {
		t.Fatalf("interval not parsed: %#v", c.UpdateCheckIntervalHours)
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
		LoadConfig(tmp.Name())
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

func TestLoadMaxMindCredentialsValid(t *testing.T) {
	acc, err := ioutil.TempFile("", "acc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(acc.Name())
	lic, err := ioutil.TempFile("", "lic")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lic.Name())
	acc.WriteString(" 123 \n")
	lic.WriteString(" abc \n")
	acc.Close()
	lic.Close()

	a, l := LoadMaxMindCredentials(acc.Name(), lic.Name())
	if a != "123" || l != "abc" {
		t.Fatalf("unexpected credentials: %q %q", a, l)
	}
}

func TestLoadMaxMindCredentialsMissing(t *testing.T) {
	if os.Getenv("TEST_FATAL_CRED") == "1" {
		LoadMaxMindCredentials("/nonexistent/account", "/nonexistent/license")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadMaxMindCredentialsMissing")
	cmd.Env = append(os.Environ(), "TEST_FATAL_CRED=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected failure")
	}
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("expected exit error, got %v", err)
	}
}
