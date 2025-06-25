package cobra

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureOutput runs fn while capturing anything written to os.Stdout.
func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	orig := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func TestHelpNoCommands(t *testing.T) {
	cmd := &Command{Use: "root"}
	out := captureOutput(func() { cmd.Help() })
	expect := "Usage: root [command]\n"
	if out != expect {
		t.Fatalf("unexpected help output: %q", out)
	}
}

func TestHelpWithCommands(t *testing.T) {
	sub1 := &Command{Use: "sub1", Short: "first"}
	sub2 := &Command{Use: "sub2", Short: "second"}
	cmd := &Command{Use: "root"}
	cmd.AddCommand(sub1, sub2)

	out := captureOutput(func() { cmd.Help() })
	expect := "Usage: root [command]\n\nAvailable Commands:\n  sub1\tfirst\n  sub2\tsecond\n"
	if out != expect {
		t.Fatalf("unexpected help output: %q", out)
	}
}
