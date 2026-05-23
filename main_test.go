package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"testing"
)

func TestBuildVersionOutputAddsVPrefixAndMetadata(t *testing.T) {
	got := buildVersionOutput("1.2.3")
	want := fmt.Sprintf("repver version v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutputPreservesExistingVPrefix(t *testing.T) {
	got := buildVersionOutput("v1.2.3")
	want := fmt.Sprintf("repver version v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestBuildVersionOutputPreservesNonReleaseVersionText(t *testing.T) {
	got := buildVersionOutput("dev")
	want := fmt.Sprintf("repver version dev (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if got != want {
		t.Fatalf("unexpected version output: got %q, want %q", got, want)
	}
}

func TestVersionFlagPrintsStandardizedOutput(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--version")
	cmd.Dir = t.TempDir()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version returned error: %v\n%s", err, output)
	}

	want := buildVersionOutput(Version) + "\n"
	if string(output) != want {
		t.Fatalf("unexpected --version output: got %q, want %q", output, want)
	}
}
