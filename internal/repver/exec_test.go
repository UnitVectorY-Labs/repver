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
