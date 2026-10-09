package repver

import (
	"testing"
)

func TestGetCommand_Found(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
		},
	}

	cmd, err := config.GetCommand("beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if cmd.Name != "beta" {
		t.Errorf("expected command name 'beta', got %q", cmd.Name)
	}
}

func TestGetCommand_NotFound(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{
			{Name: "alpha"},
		},
	}

	cmd, err := config.GetCommand("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing command")
	}
	if cmd != nil {
		t.Fatal("expected nil command for missing command")
	}
}

func TestGetCommand_EmptyCommands(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{},
	}

	cmd, err := config.GetCommand("any")
	if err == nil {
		t.Fatal("expected error for empty commands")
	}
	if cmd != nil {
		t.Fatal("expected nil command for empty commands")
	}
}

func TestConfigGetParameterNames(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{
			{
				Name: "cmd1",
				Targets: []RepverTarget{
					{Path: "a.txt", Pattern: `^(?P<version>.*)$`},
				},
			},
			{
				Name: "cmd2",
				Targets: []RepverTarget{
					{Path: "b.txt", Pattern: `^(?P<name>.*)$`},
				},
			},
		},
	}

	names, err := config.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	if !nameSet["version"] {
		t.Error("expected 'version' in parameter names")
	}
	if !nameSet["name"] {
		t.Error("expected 'name' in parameter names")
	}
	if len(names) != 2 {
		t.Errorf("expected 2 unique parameter names, got %d", len(names))
	}
}

func TestConfigGetParameterNames_Dedup(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{
			{
				Name: "cmd1",
				Targets: []RepverTarget{
					{Path: "a.txt", Pattern: `^(?P<version>.*)$`},
				},
			},
			{
				Name: "cmd2",
				Targets: []RepverTarget{
					{Path: "b.txt", Pattern: `^(?P<version>.*)$`},
				},
			},
		},
	}

	names, err := config.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 {
		t.Errorf("expected 1 unique parameter name after dedup, got %d", len(names))
	}
}

func TestConfigGetParameterNames_InvalidPattern(t *testing.T) {
	config := &RepverConfig{
		Commands: []RepverCommand{
			{
				Name: "cmd1",
				Targets: []RepverTarget{
					{Path: "a.txt", Pattern: `^(?P<version>[invalid$`},
				},
			},
		},
	}

	_, err := config.GetParameterNames()
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestCommandGetParameterNames(t *testing.T) {
	cmd := &RepverCommand{
		Name: "test",
		Targets: []RepverTarget{
			{Path: "a.txt", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)$`},
		},
	}

	names, err := cmd.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	if !nameSet["major"] {
		t.Error("expected 'major' in parameter names")
	}
	if !nameSet["minor"] {
		t.Error("expected 'minor' in parameter names")
	}
}

func TestCommandGetParameterNames_Dedup(t *testing.T) {
	cmd := &RepverCommand{
		Name: "test",
		Targets: []RepverTarget{
			{Path: "a.txt", Pattern: `^(?P<version>.*)$`},
			{Path: "b.txt", Pattern: `^(?P<version>.*)$`},
		},
	}

	names, err := cmd.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 {
		t.Errorf("expected 1 unique parameter name after dedup, got %d", len(names))
	}
}

func TestTargetGetParameterNames(t *testing.T) {
	target := &RepverTarget{
		Path:    "file.txt",
		Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`,
	}

	names, err := target.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 3 {
		t.Fatalf("expected 3 parameter names, got %d", len(names))
	}

	expected := map[string]bool{"major": false, "minor": false, "patch": false}
	for _, n := range names {
		expected[n] = true
	}
	for k, found := range expected {
		if !found {
			t.Errorf("expected parameter name %q not found", k)
		}
	}
}

func TestTargetGetParameterNames_NoGroups(t *testing.T) {
	target := &RepverTarget{
		Path:    "file.txt",
		Pattern: `^abc$`,
	}

	names, err := target.GetParameterNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 parameter names, got %d", len(names))
	}
}

func TestTargetGetParameterNames_InvalidPattern(t *testing.T) {
	target := &RepverTarget{
		Path:    "file.txt",
		Pattern: `^(?P<v>[invalid$`,
	}

	_, err := target.GetParameterNames()
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestBuildBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		vals       map[string]string
		expected   string
	}{
		{
			"single placeholder",
			"release-{{version}}",
			map[string]string{"version": "1.2.3"},
			"release-1.2.3",
		},
		{
			"multiple placeholders",
			"{{project}}-release-{{version}}",
			map[string]string{"project": "myapp", "version": "2.0.0"},
			"myapp-release-2.0.0",
		},
		{
			"no placeholders",
			"static-branch",
			map[string]string{"version": "1.0"},
			"static-branch",
		},
		{
			"empty vals",
			"release-{{version}}",
			map[string]string{},
			"release-{{version}}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := &RepverGit{BranchName: tc.branchName}
			result := git.BuildBranchName(tc.vals)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestBuildCommitMessage(t *testing.T) {
	tests := []struct {
		name          string
		commitMessage string
		vals          map[string]string
		expected      string
	}{
		{
			"single placeholder",
			"Update to {{version}}",
			map[string]string{"version": "3.0.0"},
			"Update to 3.0.0",
		},
		{
			"multiple placeholders",
			"Update {{name}} to {{version}}",
			map[string]string{"name": "app", "version": "1.0"},
			"Update app to 1.0",
		},
		{
			"no placeholders",
			"Static commit message",
			map[string]string{"version": "1.0"},
			"Static commit message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := &RepverGit{CommitMessage: tc.commitMessage}
			result := git.BuildCommitMessage(tc.vals)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGitOptionsSpecified(t *testing.T) {
	tests := []struct {
		name     string
		git      RepverGit
		expected bool
	}{
		{"no options", RepverGit{}, false},
		{"create branch", RepverGit{CreateBranch: true}, true},
		{"delete branch", RepverGit{DeleteBranch: true}, true},
		{"commit", RepverGit{Commit: true}, true},
		{"push", RepverGit{Push: true}, true},
		{"return to original", RepverGit{ReturnToOriginalBranch: true}, true},
		{"multiple options", RepverGit{CreateBranch: true, Commit: true, Push: true}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.git.GitOptionsSpecified()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetParam(t *testing.T) {
	cmd := &RepverCommand{
		Name: "test",
		Params: []RepverParam{
			{Name: "version", Pattern: `^.*$`},
			{Name: "name", Pattern: `^[a-z]+$`},
		},
	}

	t.Run("found", func(t *testing.T) {
		param := cmd.GetParam("version")
		if param == nil {
			t.Fatal("expected non-nil param")
		}
		if param.Name != "version" {
			t.Errorf("expected param name 'version', got %q", param.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		param := cmd.GetParam("nonexistent")
		if param != nil {
			t.Fatal("expected nil param for nonexistent name")
		}
	})

	t.Run("empty params", func(t *testing.T) {
		emptyCmd := &RepverCommand{Name: "empty"}
		param := emptyCmd.GetParam("any")
		if param != nil {
			t.Fatal("expected nil param from empty params list")
		}
	})
}

func TestExtractNamedGroups(t *testing.T) {
	tests := []struct {
		name           string
		param          RepverParam
		value          string
		expectedGroups map[string]string
		shouldError    bool
	}{
		{
			"semver extraction",
			RepverParam{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`},
			"1.2.3",
			map[string]string{"major": "1", "minor": "2", "patch": "3"},
			false,
		},
		{
			"no match",
			RepverParam{Name: "version", Pattern: `^(?P<major>\d+)$`},
			"abc",
			nil,
			true,
		},
		{
			"invalid pattern",
			RepverParam{Name: "bad", Pattern: `^(?P<v>[invalid$`},
			"test",
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups, err := tc.param.ExtractNamedGroups(tc.value)
			if tc.shouldError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for k, v := range tc.expectedGroups {
				if groups[k] != v {
					t.Errorf("expected group %q = %q, got %q", k, v, groups[k])
				}
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		name        string
		param       RepverParam
		value       string
		shouldError bool
	}{
		{
			"valid match",
			RepverParam{Name: "v", Pattern: `^[0-9]+$`},
			"123",
			false,
		},
		{
			"no match",
			RepverParam{Name: "v", Pattern: `^[0-9]+$`},
			"abc",
			true,
		},
		{
			"invalid pattern",
			RepverParam{Name: "v", Pattern: `^[invalid$`},
			"test",
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.ValidateValue(tc.value)
			if tc.shouldError && err == nil {
				t.Error("expected error")
			} else if !tc.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetTransformParamNames(t *testing.T) {
	tests := []struct {
		name      string
		transform string
		expected  []string
	}{
		{
			"single placeholder",
			"{{major}}",
			[]string{"major"},
		},
		{
			"multiple placeholders",
			"{{major}}.{{minor}}.{{patch}}",
			[]string{"major", "minor", "patch"},
		},
		{
			"duplicate placeholders",
			"{{major}}-{{major}}",
			[]string{"major"},
		},
		{
			"no placeholders",
			"static",
			nil,
		},
		{
			"empty transform",
			"",
			nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := &RepverTarget{
				Path:      "file.txt",
				Pattern:   `^.*$`,
				Transform: tc.transform,
			}

			names := target.GetTransformParamNames()
			if tc.expected == nil {
				if names != nil {
					t.Errorf("expected nil, got %v", names)
				}
				return
			}
			if len(names) != len(tc.expected) {
				t.Fatalf("expected %d names, got %d: %v", len(tc.expected), len(names), names)
			}
			for i, expected := range tc.expected {
				if names[i] != expected {
					t.Errorf("expected names[%d] = %q, got %q", i, expected, names[i])
				}
			}
		})
	}
}
