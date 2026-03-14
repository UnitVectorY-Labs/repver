package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runCommand(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, output)
	}
	return string(output)
}

func TestNoOpRunDoesNotChangeGitState(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	runCommand(t, tmpDir, "git", "init", "-b", "main")
	runCommand(t, tmpDir, "git", "config", "user.name", "Repver Test")
	runCommand(t, tmpDir, "git", "config", "user.email", "repver@example.com")

	repverContent := `commands:
  - name: "goversion"
    targets:
    - path: "version.txt"
      pattern: "^version: (?P<version>.*)$"
    git:
      create_branch: true
      branch_name: "release-{{version}}"
      commit: true
      commit_message: "Update version to {{version}}"
      return_to_original_branch: true
      delete_branch: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "version.txt"), []byte("version: 1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runCommand(t, tmpDir, "git", "add", ".")
	runCommand(t, tmpDir, "git", "commit", "-m", "Initial commit")

	beforeBranch := strings.TrimSpace(runCommand(t, tmpDir, "git", "rev-parse", "--abbrev-ref", "HEAD"))

	cmd := exec.Command(binary, "--command=goversion", "--param-version=1.2.3")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("repver returned error: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "No updates needed; target files already match the requested values.") {
		t.Fatalf("expected no-op message, got:\n%s", output)
	}

	afterBranch := strings.TrimSpace(runCommand(t, tmpDir, "git", "rev-parse", "--abbrev-ref", "HEAD"))
	if afterBranch != beforeBranch {
		t.Fatalf("expected branch to remain %q, got %q", beforeBranch, afterBranch)
	}

	status := strings.TrimSpace(runCommand(t, tmpDir, "git", "status", "--porcelain"))
	if status != "" {
		t.Fatalf("expected clean git status, got:\n%s", status)
	}

	branches := runCommand(t, tmpDir, "git", "branch", "--list", "release-1.2.3")
	if strings.TrimSpace(branches) != "" {
		t.Fatalf("expected no release branch to be created, got:\n%s", branches)
	}
}

func TestNoOpDryRunDoesNotPrintActionPreview(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	runCommand(t, tmpDir, "git", "init", "-b", "main")
	runCommand(t, tmpDir, "git", "config", "user.name", "Repver Test")
	runCommand(t, tmpDir, "git", "config", "user.email", "repver@example.com")

	repverContent := `commands:
  - name: "goversion"
    targets:
    - path: "version.txt"
      pattern: "^version: (?P<version>.*)$"
    git:
      create_branch: true
      branch_name: "release-{{version}}"
      commit: true
      commit_message: "Update version to {{version}}"
      return_to_original_branch: true
      delete_branch: true
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".repver"), []byte(repverContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "version.txt"), []byte("version: 1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runCommand(t, tmpDir, "git", "add", ".")
	runCommand(t, tmpDir, "git", "commit", "-m", "Initial commit")

	cmd := exec.Command(binary, "--command=goversion", "--param-version=1.2.3", "--dry-run")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("repver returned error: %v\n%s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "No updates needed; target files already match the requested values.") {
		t.Fatalf("expected no-op message, got:\n%s", got)
	}
	if strings.Contains(got, "[DRYRUN]") {
		t.Fatalf("expected no dry-run action preview for a no-op, got:\n%s", got)
	}
	if strings.Contains(got, "DRY RUN MODE ENABLED") {
		t.Fatalf("expected no dry-run banner for a no-op, got:\n%s", got)
	}
}
