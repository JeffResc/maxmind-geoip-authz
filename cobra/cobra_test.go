package cobra

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestAddCommand(t *testing.T) {
	root := &Command{Use: "root"}
	sub1 := &Command{Use: "sub1"}
	sub2 := &Command{Use: "sub2"}

	root.AddCommand(sub1, sub2)

	if len(root.commands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(root.commands))
	}
	if root.commands[0] != sub1 || root.commands[1] != sub2 {
		t.Fatalf("subcommands not added correctly")
	}
}

func TestExecuteRunsRun(t *testing.T) {
	called := false
	root := &Command{Use: "root", Run: func(cmd *Command, args []string) { called = true }}

	orig := os.Args
	os.Args = []string{"prog"}
	defer func() { os.Args = orig }()

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatalf("Run not called")
	}
}

func TestExecuteRunsRunEWithArgs(t *testing.T) {
	var got []string
	errRet := errors.New("fail")
	root := &Command{Use: "root", RunE: func(cmd *Command, args []string) error {
		got = args
		return errRet
	}}

	orig := os.Args
	os.Args = []string{"prog", "a", "b"}
	defer func() { os.Args = orig }()

	if err := root.Execute(); err != errRet {
		t.Fatalf("expected %v, got %v", errRet, err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("args not passed to RunE: %#v", got)
	}
}

func TestExecuteShowsHelp(t *testing.T) {
	root := &Command{Use: "root"}
	root.AddCommand(&Command{Use: "sub"})

	origArgs := os.Args
	os.Args = []string{"prog"}
	defer func() { os.Args = origArgs }()

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := root.Execute()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(string(out), "sub") {
		t.Fatalf("help output missing subcommand: %s", out)
	}
}

func TestExecuteRunsSubcommand(t *testing.T) {
	var got []string
	sub := &Command{Use: "sub", Run: func(cmd *Command, args []string) { got = args }}
	root := &Command{Use: "root"}
	root.AddCommand(sub)

	orig := os.Args
	os.Args = []string{"prog", "sub", "x", "y"}
	defer func() { os.Args = orig }()

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("subcommand args not passed: %#v", got)
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	root := &Command{Use: "root"}
	root.AddCommand(&Command{Use: "sub"})

	orig := os.Args
	os.Args = []string{"prog", "bad"}
	defer func() { os.Args = orig }()

	err := root.Execute()
	if err == nil || err.Error() != "unknown command: bad" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteHelpFlag(t *testing.T) {
	root := &Command{Use: "root"}
	root.AddCommand(&Command{Use: "sub"})

	origArgs := os.Args
	os.Args = []string{"prog", "-h"}
	defer func() { os.Args = origArgs }()

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := root.Execute()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(string(out), "sub") {
		t.Fatalf("help output missing subcommand: %s", out)
	}
}

func TestExecuteRunWithArgs(t *testing.T) {
	var got []string
	root := &Command{Use: "root", Run: func(cmd *Command, args []string) { got = args }}

	orig := os.Args
	os.Args = []string{"prog", "a"}
	defer func() { os.Args = orig }()

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("args not passed to Run: %#v", got)
	}
}
