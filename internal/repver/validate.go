package repver

import (
	"fmt"
	"os"
	"regexp"
)

// Validate validates the RepverConfig structure
func (c *RepverConfig) Validate() error {
	// Check if the commands are valid
	for _, command := range c.Commands {
		if err := command.Validate(); err != nil {
			return err
		}
	}

	// verify the commands are unique
	seen := make(map[string]bool)
	for _, command := range c.Commands {
		if seen[command.Name] {
			return fmt.Errorf("duplicate command name found: %s", command.Name)
		}
		seen[command.Name] = true
	}

	return nil
}

// Validate validates the RepverCommand structure
func (c *RepverCommand) Validate() error {
	// Check if the name is empty
	if c.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	// Check if the targets are valid
	for _, target := range c.Targets {
		if err := target.Validate(); err != nil {
			return err
		}
	}

	// Check if the git options are valid if any are specified
	if c.GitOptions.GitOptionsSpecified() {
		if err := c.GitOptions.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate  validates the RepverGit structure
func (g *RepverGit) Validate() error {

	if g.DeleteBranch && !g.CreateBranch {
		return fmt.Errorf("delete_branch can only be set if create_branch is set")
	}

	if g.CreateBranch && g.BranchName == "" {
		return fmt.Errorf("branch_name must be set if create_branch is set")
	}

	if g.Commit && g.CommitMessage == "" {
		return fmt.Errorf("commit_message must be set if commit is set")
	}

	if g.Push && g.Remote == "" {
		return fmt.Errorf("remote must be set if push is set")
	}

	if g.ReturnToOriginalBranch && !g.CreateBranch {
		return fmt.Errorf("return_to_original_branch can only be set if create_branch is set")
	}

	return nil
}

// Validate validates the RepverTarget structure
func (t *RepverTarget) Validate() error {
	// Check if the path is empty
	if t.Path == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	// Open the root with os.OpenRoot
	root, err := os.OpenRoot(".")
	if err != nil {
		return fmt.Errorf("failed to open root: %s", err)
	}
	defer root.Close()

	// Check if the path is within the root
	if err := checkFileWithinRoot(root, t.Path); err != nil {
		return fmt.Errorf("target path is not within the root: %s", err)
	}

	// Check if the pattern is empty
	if t.Pattern == "" {
		return fmt.Errorf("target pattern cannot be empty")
	}

	// First, check if the user is using (?<name>...) syntax instead of Go's (?P<name>...) syntax
	incorrectSyntaxRegex := regexp.MustCompile(`\(\?<([^>]+)>`)
	if incorrectSyntaxRegex.MatchString(t.Pattern) {
		// Convert the incorrect syntax to Go's regex syntax for the error message
		correctedPattern := incorrectSyntaxRegex.ReplaceAllString(t.Pattern, `(?P<$1>`)
		return fmt.Errorf("Go regex requires (?P<name>...) syntax for named capture groups, not (?<name>...). Try: %s", correctedPattern)
	}

	// Validate that pattern is a valid regex
	_, err = regexp.Compile(t.Pattern)
	if err != nil {
		return fmt.Errorf("target pattern is not a valid regex: %s", err)
	}

	err = validateNamedGroups(t.Pattern)
	if err != nil {
		return fmt.Errorf("error validating named groups: %s", err)
	}

	return nil
}

func checkFileWithinRoot(root *os.Root, path string) error {
	// Get file info using Stat within the confined root
	info, err := root.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist")
		}
		return fmt.Errorf("could not stat file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	// Try to open the file to test readability
	f, err := root.Open(path)
	if err != nil {
		return fmt.Errorf("file is not readable: %w", err)
	}
	defer f.Close()

	return nil
}

// ValidateNamedGroups takes a regex pattern as input and returns an error
// if any capturing group is not a named group. Non-capturing groups (e.g., (?:...))
// are ignored since they are not considered capturing.
func validateNamedGroups(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex: %v", err)
	}

	// SubexpNames returns a slice where the first element is always the whole match,
	// and subsequent elements correspond to the capturing groups. An empty string
	// means the group is not named.
	names := re.SubexpNames()
	for i, name := range names {
		if i == 0 { // skip the full match entry
			continue
		}
		if name == "" {
			return fmt.Errorf("unnamed capturing group at index %d", i)
		}
	}
	return nil
}
