package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitActionOutput(t *testing.T) {
	binary := buildBinary(t)
	for _, scenario := range []string{"success", "push failure", "dry run"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			remote := filepath.Join(t.TempDir(), "remote.git")
			runCommand(t, dir, "git", "init", "-b", "main")
			runCommand(t, dir, "git", "config", "user.name", "Repver Test")
			runCommand(t, dir, "git", "config", "user.email", "repver@example.com")
			runCommand(t, dir, "git", "init", "--bare", remote)
			if scenario == "push failure" {
				remote = filepath.Join(dir, "missing.git")
			}
			runCommand(t, dir, "git", "remote", "add", "origin", remote)
			config := `commands:
  - name: update
    targets:
      - path: version.txt
        pattern: '^version: (?P<version>.*)$'
    git:
      create_branch: true
      branch_name: release-{{version}}
      commit: true
      commit_message: Update version to {{version}}
      push: true
      remote: origin
      pull_request: GITHUB_CLI
      return_to_original_branch: true
      delete_branch: true
`
			for name, content := range map[string]string{".repver": config, "version.txt": "version: 1.0.0\n"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			runCommand(t, dir, "git", "add", ".")
			runCommand(t, dir, "git", "commit", "-m", "Initial")
			binDir := t.TempDir()
			// Stub only GitHub PR creation; branch, commit, push and cleanup use real git.
			if err := os.WriteFile(filepath.Join(binDir, "gh"), []byte("#!/bin/sh\nprintf '%s\\n' 'https://github.com/example/repo/pull/123'\n"), 0755); err != nil {
				t.Fatal(err)
			}
			args := []string{"--command=update", "--param-version=2.0.0", "--no-color"}
			if scenario == "dry run" {
				args = append(args, "--dry-run")
			}
			cmd := exec.Command(binary, args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if !strings.Contains(stdout.String(), "FILE CHANGES:") {
				t.Fatalf("missing diff: %s; stderr: %s; error: %v", &stdout, &stderr, err)
			}
			if strings.Contains(stdout.String(), "Committed ") {
				t.Fatalf("progress leaked to stdout: %s", &stdout)
			}
			if scenario == "dry run" {
				if err != nil || stderr.Len() != 0 {
					t.Fatalf("dry run: %v, %s", err, &stderr)
				}
				if got := strings.TrimSpace(runCommand(t, dir, "git", "branch", "--show-current")); got != "main" {
					t.Fatalf("branch: %s", got)
				}
				return
			}
			revision := "release-2.0.0"
			if scenario == "success" {
				revision = "origin/release-2.0.0"
			}
			sha := strings.TrimSpace(runCommand(t, dir, "git", "rev-parse", "--short", revision))
			expected := []string{"Created branch release-2.0.0", "Committed " + sha + ": Update version to 2.0.0"}
			if scenario == "push failure" {
				if err == nil {
					t.Fatal("expected push failure")
				}
				expected = append(expected, "Failed to push origin/release-2.0.0", "Commit "+sha+" remains on branch release-2.0.0")
				for _, absent := range []string{"Pushed origin/", "Created pull request:", "Switched to main", "Deleted local branch"} {
					if strings.Contains(stderr.String(), absent) {
						t.Fatalf("false success %q: %s", absent, &stderr)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("run: %v, %s", err, &stderr)
				}
				expected = append(expected, "Pushed origin/release-2.0.0", "Created pull request: https://github.com/example/repo/pull/123", "Switched to main", "Deleted local branch release-2.0.0")
				if got := strings.TrimSpace(runCommand(t, dir, "git", "branch", "--show-current")); got != "main" {
					t.Fatalf("branch: %s", got)
				}
				if got := runCommand(t, dir, "git", "branch", "--list", "release-2.0.0"); strings.TrimSpace(got) != "" {
					t.Fatalf("branch not deleted: %s", got)
				}
				runCommand(t, dir, "git", "--git-dir="+remote, "rev-parse", "refs/heads/release-2.0.0")
			}
			remaining := stderr.String()
			for _, line := range expected {
				index := strings.Index(remaining, line)
				if index < 0 {
					t.Fatalf("missing or unordered %q: %s", line, &stderr)
				}
				remaining = remaining[index+len(line):]
			}
		})
	}
}
