package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmdShowsHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"geoip"}

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := rootCmd.Execute()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(string(out), "serve") {
		t.Fatalf("help output missing commands: %s", out)
	}
}

func TestRootCmdUnknownCommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"geoip", "bogus"}

	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown command: bogus") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServeCommandRuns(t *testing.T) {
	called := false
	origRunE := serveCmd.RunE
	serveCmd.RunE = func(cmd *cobra.Command, args []string) error { called = true; return nil }
	defer func() { serveCmd.RunE = origRunE }()

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"geoip", "serve"}

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatalf("serve command not executed")
	}
}

func TestUpdateDatabaseCommandRuns(t *testing.T) {
	called := false
	origRun := updateDBCmd.Run
	updateDBCmd.Run = func(cmd *cobra.Command, args []string) { called = true }
	defer func() { updateDBCmd.Run = origRun }()

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"geoip", "update", "database"}

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatalf("update database command not executed")
	}
}

func TestHelpFlagShowsHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"geoip", "-h"}

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := rootCmd.Execute()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(string(out), "serve") {
		t.Fatalf("help output missing commands: %s", out)
	}
}
