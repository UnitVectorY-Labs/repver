package repver

import (
	"bytes"
	"os"
	"testing"
)

func TestDebuglnEnabled(t *testing.T) {
	// Save original state
	origDebug := Debug
	defer func() { Debug = origDebug }()

	Debug = true

	// Capture stderr
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	Debugln("test message %d", 42)

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected debug output when Debug is true")
	}
	if !bytes.Contains([]byte(output), []byte("[DEBUG]")) {
		t.Errorf("expected [DEBUG] prefix, got %q", output)
	}
	if !bytes.Contains([]byte(output), []byte("test message 42")) {
		t.Errorf("expected 'test message 42' in output, got %q", output)
	}
}

func TestDebuglnDisabled(t *testing.T) {
	// Save original state
	origDebug := Debug
	defer func() { Debug = origDebug }()

	Debug = false

	// Capture stderr
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	Debugln("this should not appear %d", 99)

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected no output when Debug is false, got %q", output)
	}
}
