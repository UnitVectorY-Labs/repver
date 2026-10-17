package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSamePathTargetsCLI(t *testing.T) {
	binary := buildBinary(t)
	for _, tc := range []struct {
		name    string
		reverse bool
		dryRun  bool
		noOp    bool
		dirty   bool
	}{
		{name: "commit both changes"},
		{name: "reverse target order", reverse: true},
		{name: "dry run", dryRun: true},
		{name: "no-op", noOp: true},
		{name: "dirty no-op", noOp: true, dirty: true},
		{name: "dirty workspace rejected", dirty: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Every command runs in a disposable repository with no remotes.
			dir := t.TempDir()
			runCommand(t, dir, "git", "init", "-b", "main")
			runCommand(t, dir, "git", "config", "user.name", "Repver Test")
			runCommand(t, dir, "git", "config", "user.email", "repver@example.com")
			first, second := "go-version", "go-version-input"
			if tc.reverse {
				first, second = second, first
			}
			config := fmt.Sprintf(`commands:
  - name: goversion
    targets:
      - path: workflow.yml
        pattern: "^%s: '(?P<version>.*)'$"
      - path: workflow.yml
        pattern: "^%s: '(?P<version>.*)'$"
      - path: workflow.yml
        pattern: "^%s: '(?P<version>.*)'$"
    git:
      create_branch: true
      branch_name: release-{{version}}
      commit: true
      commit_message: Update version to {{version}}
`, first, second, second)
			want := "go-version: '2.0'\ngo-version-input: '2.0'\n"
			original := "go-version: '1.0'\ngo-version-input: '1.0'\n"
			if tc.noOp {
				original = want
			}
			for name, content := range map[string]string{".repver": config, "workflow.yml": original} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			runCommand(t, dir, "git", "add", ".")
			runCommand(t, dir, "git", "commit", "-m", "Initial commit")
			if tc.dirty {
				if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("local work"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			beforeRefs := runCommand(t, dir, "git", "show-ref")
			beforeStatus := runCommand(t, dir, "git", "status", "--porcelain")
			args := []string{"--command=goversion", "--param-version=2.0", "--no-color"}
			if tc.dryRun {
				args = append(args, "--dry-run")
			}
			cmd := exec.Command(binary, args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			rejected := tc.dirty && !tc.noOp
			if (err != nil) != rejected {
				t.Fatalf("unexpected result: %v\n%s", err, output)
			}
			unchanged := tc.dryRun || tc.noOp || rejected
			if unchanged {
				want = original
				if refs := runCommand(t, dir, "git", "show-ref"); refs != beforeRefs {
					t.Fatalf("git refs changed: %s", refs)
				}
				if status := runCommand(t, dir, "git", "status", "--porcelain"); status != beforeStatus {
					t.Fatalf("git status changed: %s", status)
				}
				if branch := strings.TrimSpace(runCommand(t, dir, "git", "branch", "--show-current")); branch != "main" {
					t.Fatalf("branch changed to %q", branch)
				}
			} else {
				if committed := runCommand(t, dir, "git", "show", "HEAD:workflow.yml"); committed != want {
					t.Fatalf("commit lost target changes: %q", committed)
				}
				if status := runCommand(t, dir, "git", "status", "--porcelain"); status != "" {
					t.Fatalf("expected clean workspace: %s", status)
				}
			}
			content, err := os.ReadFile(filepath.Join(dir, "workflow.yml"))
			if err != nil || string(content) != want {
				t.Fatalf("unexpected file: %q, %v; want %q", content, err, want)
			}
			if tc.noOp && !strings.Contains(string(output), "No updates needed") {
				t.Fatalf("missing no-op message: %s", output)
			}
			if tc.dryRun {
				for _, expected := range []string{"+ go-version: '2.0'", "+ go-version-input: '2.0'", "File: workflow.yml", "  - workflow.yml\n"} {
					if strings.Count(string(output), expected) != 1 {
						t.Fatalf("expected %q exactly once in preview:\n%s", expected, output)
					}
				}
			}
		})
	}
}
