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

	// Validate the command name format
	if err := validateCommandName(c.Name); err != nil {
		return err
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

// Validate validates the RepverGit structure
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

	if g.PullRequest == "" {
		g.PullRequest = "NO"
	}

	if g.PullRequest != "NO" && g.PullRequest != "GITHUB_CLI" {
		return fmt.Errorf("invalid pull_request value: %s", g.PullRequest)
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

	// Validate the pattern
	if err := validatePattern(t.Pattern); err != nil {
		return fmt.Errorf("target pattern is not valid: %s", err)
	}

	return nil
}

// checkFileWithinRoot checks if the file is within the confined root and is readable.
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

// validateCommandName checks if the command name is valid.
func validateCommandName(name string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9]{1,30}$`)
	if !re.MatchString(name) {
		return fmt.Errorf("command name must be alphanumeric and between 1 and 30 characters")
	}

	return nil
}

// validatePattern checks if the pattern is valid.
func validatePattern(pattern string) error {

	// Check if the pattern is empty
	if pattern == "" {
		return fmt.Errorf("cannot be empty")
	}

	// The regex pattern must start with ^ and end with $
	if !regexp.MustCompile(`^\^.*\$$`).MatchString(pattern) {
		return fmt.Errorf("must start with ^ and end with $ defining a pattern for the entire line")
	}

	// First, check if the user is using (?<name>...) syntax instead of Go's (?P<name>...) syntax
	incorrectSyntaxRegex := regexp.MustCompile(`\(\?<([^>]+)>`)
	if incorrectSyntaxRegex.MatchString(pattern) {
		// Convert the incorrect syntax to Go's regex syntax for the error message
		correctedPattern := incorrectSyntaxRegex.ReplaceAllString(pattern, `(?P<$1>`)
		return fmt.Errorf("Go regex requires (?P<name>...) syntax for named capture groups, not (?<name>...). Try: %s", correctedPattern)
	}

	// Validate that pattern is a valid regex
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("not a valid regex: %s", err)
	}

	err = validateNamedGroups(pattern)
	if err != nil {
		return err
	}

	return nil
}

// validateNamedGroups checks if the named groups in the regex pattern are valid.
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

	// Check for nested named groups by examining the pattern structure
	namedGroupPattern := regexp.MustCompile(`\(\?P<([^>]+)>`)
	matches := namedGroupPattern.FindAllStringIndex(pattern, -1)

	if len(matches) > 1 {
		// Check each named group to see if it contains another named group
		for i, outerMatch := range matches {
			outerStart := outerMatch[0]

			// Find the closing parenthesis for this group
			// This is an approximation - a proper parser would be better for complex patterns
			depth := 1
			outerEnd := outerStart + 1
			for outerEnd < len(pattern) && depth > 0 {
				if pattern[outerEnd] == '(' {
					depth++
				} else if pattern[outerEnd] == ')' {
					depth--
				}
				outerEnd++
			}

			// Check if any other named group starts within this one
			for j, innerMatch := range matches {
				if i != j && innerMatch[0] > outerStart && innerMatch[0] < outerEnd {
					return fmt.Errorf("nested named capture groups are not allowed")
				}
			}
		}
	}

	return nil
}
