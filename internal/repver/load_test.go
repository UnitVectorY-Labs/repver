package repver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValidYAML(t *testing.T) {
	yamlContent := `commands:
  - name: "goversion"
    targets:
    - path: "version.txt"
      pattern: "^version: (?P<version>.*)$"
`
	config, err := Parse(yamlContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(config.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(config.Commands))
	}
	if config.Commands[0].Name != "goversion" {
		t.Errorf("expected command name 'goversion', got %q", config.Commands[0].Name)
	}
	if len(config.Commands[0].Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(config.Commands[0].Targets))
	}
	if config.Commands[0].Targets[0].Path != "version.txt" {
		t.Errorf("expected target path 'version.txt', got %q", config.Commands[0].Targets[0].Path)
	}
}

func TestParseMultipleCommands(t *testing.T) {
	yamlContent := `commands:
  - name: "cmd1"
    targets:
    - path: "a.txt"
      pattern: "^(?P<v>.*)$"
  - name: "cmd2"
    targets:
    - path: "b.txt"
      pattern: "^(?P<v>.*)$"
`
	config, err := Parse(yamlContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(config.Commands))
	}
	if config.Commands[0].Name != "cmd1" {
		t.Errorf("expected first command name 'cmd1', got %q", config.Commands[0].Name)
	}
	if config.Commands[1].Name != "cmd2" {
		t.Errorf("expected second command name 'cmd2', got %q", config.Commands[1].Name)
	}
}

func TestParseWithParams(t *testing.T) {
	yamlContent := `commands:
  - name: "test"
    params:
    - name: "version"
      pattern: "^(?P<major>\\d+)\\.(?P<minor>\\d+)$"
    targets:
    - path: "file.txt"
      pattern: "^v(?P<ver>.*)$"
      transform: "{{major}}.{{minor}}"
`
	config, err := Parse(yamlContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := config.Commands[0]
	if len(cmd.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(cmd.Params))
	}
	if cmd.Params[0].Name != "version" {
		t.Errorf("expected param name 'version', got %q", cmd.Params[0].Name)
	}
	if cmd.Targets[0].Transform != "{{major}}.{{minor}}" {
		t.Errorf("expected transform '{{major}}.{{minor}}', got %q", cmd.Targets[0].Transform)
	}
}

func TestParseWithGitOptions(t *testing.T) {
	yamlContent := `commands:
  - name: "release"
    targets:
    - path: "ver.txt"
      pattern: "^(?P<v>.*)$"
    git:
      create_branch: true
      branch_name: "release-{{v}}"
      commit: true
      commit_message: "Release {{v}}"
      push: true
      remote: "origin"
      pull_request: "GITHUB_CLI"
      return_to_original_branch: true
      delete_branch: true
`
	config, err := Parse(yamlContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	git := config.Commands[0].GitOptions
	if !git.CreateBranch {
		t.Error("expected CreateBranch to be true")
	}
	if git.BranchName != "release-{{v}}" {
		t.Errorf("expected BranchName 'release-{{v}}', got %q", git.BranchName)
	}
	if !git.Commit {
		t.Error("expected Commit to be true")
	}
	if git.CommitMessage != "Release {{v}}" {
		t.Errorf("expected CommitMessage 'Release {{v}}', got %q", git.CommitMessage)
	}
	if !git.Push {
		t.Error("expected Push to be true")
	}
	if git.Remote != "origin" {
		t.Errorf("expected Remote 'origin', got %q", git.Remote)
	}
	if git.PullRequest != "GITHUB_CLI" {
		t.Errorf("expected PullRequest 'GITHUB_CLI', got %q", git.PullRequest)
	}
	if !git.ReturnToOriginalBranch {
		t.Error("expected ReturnToOriginalBranch to be true")
	}
	if !git.DeleteBranch {
		t.Error("expected DeleteBranch to be true")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	yamlContent := `this is not valid yaml: {{{`
	config, err := Parse(yamlContent)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if config != nil {
		t.Fatal("expected nil config for invalid YAML")
	}
}

func TestParseEmptyYAML(t *testing.T) {
	config, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config for empty YAML")
	}
	if len(config.Commands) != 0 {
		t.Errorf("expected 0 commands for empty YAML, got %d", len(config.Commands))
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, ".repver")
	yamlContent := `commands:
  - name: "loadtest"
    targets:
    - path: "test.txt"
      pattern: "^(?P<v>.*)$"
`
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := Load(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(config.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(config.Commands))
	}
	if config.Commands[0].Name != "loadtest" {
		t.Errorf("expected command name 'loadtest', got %q", config.Commands[0].Name)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	config, err := Load("/nonexistent/path/.repver")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if config != nil {
		t.Fatal("expected nil config for non-existent file")
	}
}
