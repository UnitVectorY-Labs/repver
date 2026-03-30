package repver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutionPlanNoOp(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("version: 1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "1.2.3"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected execution plan")
	}
	if plan.Modified {
		t.Fatal("expected no changes to be planned")
	}

	modified, err := target.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan returned error: %v", err)
	}
	if modified {
		t.Fatal("expected ExecutePlan to be a no-op")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "version: 1.2.3\n" {
		t.Fatalf("file was modified unexpectedly: %q", string(content))
	}
}

func TestCommandPlanComposesTargets(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "workflow.yml")
			otherPath := filepath.Join(dir, "other.txt")
			original := "go-version: '1.0'\ngo-version-input: '1.0'\n"
			for name, content := range map[string]string{path: original, otherPath: "version: 1.0\n"} {
				if err := os.WriteFile(name, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			first := RepverTarget{Path: path, Pattern: `^go-version: '(?P<version>.*)'$`}
			second := RepverTarget{Path: path, Pattern: `^go-version-input: '(?P<version>.*)'$`}
			if reverse {
				first, second = second, first
			}
			command := RepverCommand{Targets: []RepverTarget{
				first,
				{Path: otherPath, Pattern: `^version: (?P<version>.*)$`},
				second,
				second, // A trailing no-op must retain all earlier changes.
			}}
			plans, err := command.Plan(map[string]string{"version": "2.0"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(plans) != 2 {
				t.Fatalf("expected one plan per path, got %d", len(plans))
			}
			if plans[0].Path != path || !plans[0].Modified || plans[0].ModifiedContent != "go-version: '2.0'\ngo-version-input: '2.0'\n" || len(plans[0].Changes) != 2 {
				t.Fatalf("unexpected composed plan: %+v", plans[0])
			}
			if plans[1].Path != otherPath || !plans[1].Modified || plans[1].ModifiedContent != "version: 2.0\n" {
				t.Fatalf("unexpected other-file plan: %+v", plans[1])
			}
			content, err := os.ReadFile(path)
			if err != nil || string(content) != original {
				t.Fatalf("planning changed the file: %q, %v", content, err)
			}
		})
	}
}

func TestCommandPlanUsesPreviousOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "version.txt")
	if err := os.WriteFile(path, []byte("version: old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	command := RepverCommand{Targets: []RepverTarget{
		{Path: path, Pattern: `^version: (?P<version>old)$`},
		{Path: path, Pattern: `^version: (?P<version>new)$`, Transform: "{{final}}"},
	}}
	for _, final := range []string{"final", "old"} {
		plans, err := command.Plan(map[string]string{"version": "new"}, map[string]string{"final": final})
		if err != nil {
			t.Fatal(err)
		}
		if len(plans) != 1 || plans[0].ModifiedContent != "version: "+final+"\n" || plans[0].Modified != (final != "old") {
			t.Fatalf("unexpected plan for final value %q: %+v", final, plans)
		}
	}
}

func TestCommandPlanFailureDoesNotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "version.txt")
	original := "version: old\n"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	for _, target := range []RepverTarget{
		{Path: path + ".missing", Pattern: `(?P<version>.*)`},
		{Path: path, Pattern: `[`},
	} {
		command := RepverCommand{Targets: []RepverTarget{
			{Path: path, Pattern: `^version: (?P<version>.*)$`},
			target,
		}}
		if _, err := command.Plan(map[string]string{"version": "new"}, nil); err == nil {
			t.Fatal("expected planning error")
		}
		content, err := os.ReadFile(path)
		if err != nil || string(content) != original {
			t.Fatalf("failed planning changed the file: %q, %v", content, err)
		}
	}
}

func TestPlanWithModification(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected execution plan")
	}
	if !plan.Modified {
		t.Fatal("expected changes to be planned")
	}
	if len(plan.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(plan.Changes))
	}
	if plan.Changes[0].OldLine != "version: 1.0.0" {
		t.Errorf("expected old line 'version: 1.0.0', got %q", plan.Changes[0].OldLine)
	}
	if plan.Changes[0].NewLine != "version: 2.0.0" {
		t.Errorf("expected new line 'version: 2.0.0', got %q", plan.Changes[0].NewLine)
	}
	if plan.Changes[0].LineNumber != 1 {
		t.Errorf("expected line number 1, got %d", plan.Changes[0].LineNumber)
	}
}

func TestPlanWithMultipleMatches(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "multi.txt")
	content := "version: 1.0.0\nother line\nversion: 1.0.0\n"
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !plan.Modified {
		t.Fatal("expected changes to be planned")
	}
	if len(plan.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(plan.Changes))
	}
	if plan.Changes[0].LineNumber != 1 {
		t.Errorf("expected first change on line 1, got %d", plan.Changes[0].LineNumber)
	}
	if plan.Changes[1].LineNumber != 3 {
		t.Errorf("expected second change on line 3, got %d", plan.Changes[1].LineNumber)
	}
}

func TestPlanWithTransform(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("go 1.20\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:      targetPath,
		Pattern:   `^go (?P<gover>.*)$`,
		Transform: "{{major}}.{{minor}}",
	}

	extractedGroups := map[string]string{
		"major": "1",
		"minor": "26",
	}

	plan, err := target.Plan(map[string]string{"gover": "1.26"}, extractedGroups)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !plan.Modified {
		t.Fatal("expected changes to be planned")
	}
	if len(plan.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(plan.Changes))
	}
	if plan.Changes[0].NewLine != "go 1.26" {
		t.Errorf("expected new line 'go 1.26', got %q", plan.Changes[0].NewLine)
	}
}

func TestPlanNoMatchingLines(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "noop.txt")
	if err := os.WriteFile(targetPath, []byte("no matching lines here\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "1.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan.Modified {
		t.Fatal("expected no modification when no lines match")
	}
}

func TestPlanInvalidPattern(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(targetPath, []byte("test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^(?P<v>[invalid$`,
	}

	_, err := target.Plan(map[string]string{"v": "test"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestPlanFileNotFound(t *testing.T) {
	target := RepverTarget{
		Path:    "/nonexistent/path/file.txt",
		Pattern: `^(?P<v>.*)$`,
	}

	_, err := target.Plan(map[string]string{"v": "test"}, nil)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExecutePlanNil(t *testing.T) {
	target := RepverTarget{
		Path:    "file.txt",
		Pattern: `^.*$`,
	}

	_, err := target.ExecutePlan(nil)
	if err == nil {
		t.Fatal("expected error for nil plan")
	}
}

func TestExecutePlanWriteFile(t *testing.T) {
	// Save and restore DryRun state
	origDryRun := DryRun
	defer func() { DryRun = origDryRun }()
	DryRun = false

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	modified, err := target.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan returned error: %v", err)
	}
	if !modified {
		t.Fatal("expected file to be modified")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "version: 2.0.0") {
		t.Fatalf("expected file to contain 'version: 2.0.0', got %q", string(content))
	}
}

func TestExecutePlanDryRun(t *testing.T) {
	// Save and restore DryRun state
	origDryRun := DryRun
	defer func() { DryRun = origDryRun }()
	DryRun = true

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	originalContent := "version: 1.0.0\n"
	if err := os.WriteFile(targetPath, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	modified, err := target.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan returned error: %v", err)
	}
	if !modified {
		t.Fatal("expected dry run to report modification")
	}

	// Verify file was NOT actually modified
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != originalContent {
		t.Fatalf("expected file to remain unchanged in dry run, got %q", string(content))
	}
}

func TestExecuteConvenienceMethod(t *testing.T) {
	// Save and restore DryRun state
	origDryRun := DryRun
	defer func() { DryRun = origDryRun }()
	DryRun = false

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	modified, err := target.Execute(map[string]string{"version": "3.0.0"}, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !modified {
		t.Fatal("expected file to be modified")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "version: 3.0.0") {
		t.Fatalf("expected file to contain 'version: 3.0.0', got %q", string(content))
	}
}

func TestExecuteNoOp(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	modified, err := target.Execute(map[string]string{"version": "1.0.0"}, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if modified {
		t.Fatal("expected no modification when value is the same")
	}
}

func TestPlanPreservesFinalNewline(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	// File with trailing newline
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !strings.HasSuffix(plan.ModifiedContent, "\n") {
		t.Error("expected modified content to end with newline")
	}
}

func TestPlanFileWithoutTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "version.txt")
	// File without trailing newline
	if err := os.WriteFile(targetPath, []byte("version: 1.0.0"), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if strings.HasSuffix(plan.ModifiedContent, "\n") {
		t.Error("expected modified content to NOT end with newline when original didn't")
	}
}

func TestPlanMultipleLinesUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "config.txt")
	content := "line one\nversion: 1.0.0\nline three\n"
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	target := RepverTarget{
		Path:    targetPath,
		Pattern: `^version: (?P<version>.*)$`,
	}

	plan, err := target.Plan(map[string]string{"version": "2.0.0"}, nil)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !plan.Modified {
		t.Fatal("expected changes")
	}

	// Verify non-matching lines are preserved
	if !strings.Contains(plan.ModifiedContent, "line one") {
		t.Error("expected 'line one' to be preserved")
	}
	if !strings.Contains(plan.ModifiedContent, "line three") {
		t.Error("expected 'line three' to be preserved")
	}
	if !strings.Contains(plan.ModifiedContent, "version: 2.0.0") {
		t.Error("expected 'version: 2.0.0' in modified content")
	}
}
