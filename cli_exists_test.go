package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary builds the repver binary for testing
func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "repver")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = filepath.Dir(os.Args[0])
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try from the module root
		cmd = exec.Command("go", "build", "-o", binary, ".")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to build binary: %v\n%s", err, output)
		}
	}
	return binary
}

func TestExistsMode_NoRepverFile(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	cmd := exec.Command(binary, "--command=test", "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code when .repver is missing")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}

func TestExistsMode_ValidRepverCommandExists(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create a valid .repver file
	repverContent := `commands:
  - name: "testcmd"
    targets:
    - path: "test.txt"
      pattern: "^version: (?P<version>.*)$"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the target file
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "--command=testcmd", "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit code 0, got error: %v", err)
	}
}

func TestExistsMode_ValidRepverCommandMissing(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create a valid .repver file
	repverContent := `commands:
  - name: "testcmd"
    targets:
    - path: "test.txt"
      pattern: "^version: (?P<version>.*)$"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the target file
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "--command=nonexistent", "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code when command is missing")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}

func TestExistsMode_InvalidRepver(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create an invalid .repver file (invalid YAML)
	repverContent := `this is not valid yaml: {{{`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "--command=test", "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code when .repver is invalid")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}

func TestExistsMode_DoesNotRequireParams(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create a valid .repver file with a command that requires params
	repverContent := `commands:
  - name: "paramcmd"
    targets:
    - path: "test.txt"
      pattern: "^version: (?P<version>.*)$"
    - path: "other.txt"
      pattern: "^name: (?P<name>.*)$"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the target files
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("name: example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Invoke --exists without any --param-* flags - should still succeed
	cmd := exec.Command(binary, "--command=paramcmd", "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit code 0 (exists mode ignores params), got error: %v", err)
	}
}

func TestExistsMode_NoCommandProvided(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create a valid .repver file
	repverContent := `commands:
  - name: "testcmd"
    targets:
    - path: "test.txt"
      pattern: "^version: (?P<version>.*)$"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the target file
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "--exists")
	cmd.Dir = tmpDir
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code when --command is not provided with --exists")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}
