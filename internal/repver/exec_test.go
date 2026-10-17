package repver

import (
	"os"
	"path/filepath"
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
