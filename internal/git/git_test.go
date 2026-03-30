package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a new git repository in a temporary directory and returns
// the directory path. It sets up user config and makes an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	runGit(t, tmpDir, "init", "-b", "main")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Create an initial file and commit
	filePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(filePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	return tmpDir
}

// runGit runs a git command in the given directory and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
	return string(output)
}

func TestIsGitRoot(t *testing.T) {
	repoDir := initTestRepo(t)

	// Save and restore the working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("at git root", func(t *testing.T) {
		if err := os.Chdir(repoDir); err != nil {
			t.Fatal(err)
		}

		isRoot, err := IsGitRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isRoot {
			t.Error("expected to be at git root")
		}
	})

	t.Run("in subdirectory", func(t *testing.T) {
		subDir := filepath.Join(repoDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(subDir); err != nil {
			t.Fatal(err)
		}

		isRoot, err := IsGitRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isRoot {
			t.Error("expected not to be at git root in subdirectory")
		}
	})
}

func TestIsGitRoot_NotARepo(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err = IsGitRoot()
	if err == nil {
		t.Error("expected error when not in a git repo")
	}
}

func TestBranchExists(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	t.Run("existing branch", func(t *testing.T) {
		exists, err := BranchExists("main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected 'main' branch to exist")
		}
	})

	t.Run("non-existing branch", func(t *testing.T) {
		exists, err := BranchExists("nonexistent-branch")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected 'nonexistent-branch' to not exist")
		}
	})
}

func TestGetCurrentBranch(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("expected branch 'main', got %q", branch)
	}
}

func TestCreateAndSwitchBranch(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	_, err = CreateAndSwitchBranch("feature-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error getting branch: %v", err)
	}
	if branch != "feature-test" {
		t.Errorf("expected to be on branch 'feature-test', got %q", branch)
	}

	exists, err := BranchExists("feature-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected 'feature-test' branch to exist")
	}
}

func TestSwitchToBranch(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create another branch to switch to
	runGit(t, repoDir, "checkout", "-b", "other-branch")
	runGit(t, repoDir, "checkout", "main")

	_, err = SwitchToBranch("other-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	if branch != "other-branch" {
		t.Errorf("expected to be on 'other-branch', got %q", branch)
	}
}

func TestSwitchToBranch_NonExistent(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	_, err = SwitchToBranch("does-not-exist")
	if err == nil {
		t.Error("expected error when switching to non-existent branch")
	}
}

func TestCheckGitClean(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	t.Run("clean repo", func(t *testing.T) {
		err := CheckGitClean()
		if err != nil {
			t.Fatalf("expected clean repo, got error: %v", err)
		}
	})

	t.Run("dirty repo", func(t *testing.T) {
		// Create an untracked file to make the repo dirty
		dirtyFile := filepath.Join(repoDir, "dirty.txt")
		if err := os.WriteFile(dirtyFile, []byte("dirty"), 0644); err != nil {
			t.Fatal(err)
		}

		err := CheckGitClean()
		if err == nil {
			t.Error("expected error for dirty repo")
		}

		// Clean up
		os.Remove(dirtyFile)
	})
}

func TestAddAndCommitFiles(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create a new file to commit
	newFile := filepath.Join(repoDir, "newfile.txt")
	if err := os.WriteFile(newFile, []byte("new content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := AddAndCommitFiles([]string{"newfile.txt"}, "Add new file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output")
	}

	// Verify the commit happened
	err = CheckGitClean()
	if err != nil {
		t.Fatalf("repo should be clean after commit, got: %v", err)
	}

	// Verify commit message
	logOutput := runGit(t, repoDir, "log", "--oneline", "-1")
	if !strings.Contains(logOutput, "Add new file") {
		t.Errorf("expected commit message in log, got: %s", logOutput)
	}
}

func TestAddAndCommitFiles_MultipleFiles(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create multiple files
	for _, name := range []string{"file1.txt", "file2.txt"} {
		if err := os.WriteFile(filepath.Join(repoDir, name), []byte(name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	_, err = AddAndCommitFiles([]string{"file1.txt", "file2.txt"}, "Add two files")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = CheckGitClean()
	if err != nil {
		t.Fatalf("repo should be clean after commit, got: %v", err)
	}
}

func TestDeleteLocalBranch(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create a branch, switch back to main, then delete the branch
	runGit(t, repoDir, "checkout", "-b", "to-delete")
	runGit(t, repoDir, "checkout", "main")

	_, err = DeleteLocalBranch("to-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, err := BranchExists("to-delete")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected branch 'to-delete' to no longer exist")
	}
}

func TestDeleteLocalBranch_NonExistent(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	_, err = DeleteLocalBranch("does-not-exist")
	if err == nil {
		t.Error("expected error when deleting non-existent branch")
	}
}

func TestPushChanges(t *testing.T) {
	// Create a bare remote repo and a working repo that pushes to it
	tmpDir := t.TempDir()
	bareDir := filepath.Join(tmpDir, "bare.git")
	workDir := filepath.Join(tmpDir, "work")

	// Create bare repo
	cmd := exec.Command("git", "init", "--bare", bareDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create bare repo: %v\n%s", err, output)
	}

	// Create a working repo (not a clone, to control the branch name)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "init", "-b", "main")
	runGit(t, workDir, "config", "user.name", "Test User")
	runGit(t, workDir, "config", "user.email", "test@example.com")
	runGit(t, workDir, "remote", "add", "origin", bareDir)

	// Create initial commit and push
	initFile := filepath.Join(workDir, "init.txt")
	if err := os.WriteFile(initFile, []byte("initial\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", ".")
	runGit(t, workDir, "commit", "-m", "Initial commit")
	runGit(t, workDir, "push", "-u", "origin", "main")

	// Create a new commit to push via the PushChanges function
	newFile := filepath.Join(workDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new content\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", ".")
	runGit(t, workDir, "commit", "-m", "New commit")

	// Save and restore working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	_, err = PushChanges("origin", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPushChanges_InvalidRemote(t *testing.T) {
	repoDir := initTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	_, err = PushChanges("nonexistent-remote", "main")
	if err == nil {
		t.Error("expected error when pushing to non-existent remote")
	}
}

func TestCreateGitHubPullRequest_NoGH(t *testing.T) {
	// This test verifies error handling when gh CLI is not available or
	// not configured. In CI/test environments, gh is typically not authenticated.
	_, err := CreateGitHubPullRequest()
	if err == nil {
		// If gh is available and authenticated, the call might succeed (unlikely in tests),
		// but we still want to exercise the code path.
		t.Skip("gh CLI is authenticated; skipping error path test as success is valid")
	}
	// Verify we get a meaningful error
	if err != nil {
		if !strings.Contains(err.Error(), "error creating GitHub pull request") {
			t.Errorf("expected error about creating PR, got: %v", err)
		}
	}
}
