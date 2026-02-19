package repver

import (
	"fmt"
	"regexp"
	"strings"
)

type RepverConfig struct {
	// Commands is an array of version modification commands
	Commands []RepverCommand `yaml:"commands"`
}

// RepverParam defines a parameter validation configuration
type RepverParam struct {
	// Name is the name of the parameter (must match command-line --param-<name>)
	Name string `yaml:"name"`
	// Pattern is the regex pattern to validate and extract values from the parameter
	// It can contain named capture groups (e.g., (?P<major>\d+)) for use in transforms
	Pattern string `yaml:"pattern"`
}

type RepverCommand struct {
	// Name of the command
	Name string `yaml:"name"`
	// Params defines optional validation patterns for command-line parameters
	Params []RepverParam `yaml:"params"`
	// Targets is a list of files and patterns to modify
	Targets []RepverTarget `yaml:"targets"`
	// GitOptions configures Git operations to perform
	GitOptions RepverGit `yaml:"git"`
}

type RepverGit struct {
	// CreateBranch indicates whether to create a new branch
	CreateBranch bool `yaml:"create_branch"`
	// DeleteBranch indicates whether to delete the branch after execution
	// Only used when ReturnToOriginalBranch is true
	DeleteBranch bool `yaml:"delete_branch"`
	// BranchName is the name for the new branch
	BranchName string `yaml:"branch_name"`
	// Commit indicates whether to commit changes
	Commit bool `yaml:"commit"`
	// CommitMessage is the message to use for the commit
	CommitMessage string `yaml:"commit_message"`
	// Push indicates whether to push changes
	Push bool `yaml:"push"`
	// Remote is the Git remote to push to
	Remote string `yaml:"remote"`
	// PullRequest specifies whether to open a PR (values: NO, GITHUB_CLI)
	PullRequest string `yaml:"pull_request"`
	// ReturnToOriginalBranch indicates whether to switch back to the original branch
	ReturnToOriginalBranch bool `yaml:"return_to_original_branch"`
}

type RepverTarget struct {
	// Path to the target file
	Path string `yaml:"path"`
	// Pattern is the regex pattern to match content in the target file
	Pattern string `yaml:"pattern"`
	// Transform specifies how to transform parameter values using named groups from params
	// Uses {{name}} syntax to reference named groups from the params pattern
	// If not specified, the raw parameter value is used
	Transform string `yaml:"transform"`
}

// GetCommand returns a command by name; if not found, it returns an error
func (c *RepverConfig) GetCommand(name string) (*RepverCommand, error) {
	for _, command := range c.Commands {
		if command.Name == name {
			return &command, nil
		}
	}
	return nil, fmt.Errorf("command %s not found", name)
}

// GetParameterNames returns a list of all unique parameter names
func (c *RepverConfig) GetParameterNames() ([]string, error) {
	uniqueSet := make(map[string]struct{})
	for _, command := range c.Commands {
		groups, err := command.GetParameterNames()
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			uniqueSet[group] = struct{}{}
		}
	}

	captureGroups := make([]string, 0, len(uniqueSet))
	for group := range uniqueSet {
		captureGroups = append(captureGroups, group)
	}

	return captureGroups, nil
}

// GetParameterNames returns a list of all unique parameter names
func (c *RepverCommand) GetParameterNames() ([]string, error) {
	uniqueSet := make(map[string]struct{})
	for _, target := range c.Targets {
		groups, err := target.GetParameterNames()
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			uniqueSet[group] = struct{}{}
		}
	}

	captureGroups := make([]string, 0, len(uniqueSet))
	for group := range uniqueSet {
		captureGroups = append(captureGroups, group)
	}

	return captureGroups, nil
}

// GetParameterNames returns a list of all unique parameter names
func (t *RepverTarget) GetParameterNames() ([]string, error) {
	re, err := regexp.Compile(t.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex: %s", err)
	}
	names := re.SubexpNames()
	captureGroups := []string{}
	for i, name := range names {
		if i > 0 && name != "" {
			captureGroups = append(captureGroups, name)
		}
	}
	return captureGroups, nil
}

// BuildBranchName builds the branch name by replacing placeholders with values
func (g *RepverGit) BuildBranchName(vals map[string]string) string {
	branchName := g.BranchName
	for key, val := range vals {
		branchName = strings.ReplaceAll(branchName, "{{"+key+"}}", val)
	}

	return branchName
}

// BuildCommitMessage builds the commit message by replacing placeholders with values
func (g *RepverGit) BuildCommitMessage(vals map[string]string) string {
	commitMessage := g.CommitMessage
	for key, val := range vals {
		commitMessage = strings.ReplaceAll(commitMessage, "{{"+key+"}}", val)
	}

	return commitMessage
}

// GitOptionsSpecified checks if any Git options are specified
func (g *RepverGit) GitOptionsSpecified() bool {
	return g.CreateBranch || g.DeleteBranch || g.Commit || g.Push || g.ReturnToOriginalBranch
}

// GetParam returns a param definition by name; if not found, it returns nil
func (c *RepverCommand) GetParam(name string) *RepverParam {
	for i := range c.Params {
		if c.Params[i].Name == name {
			return &c.Params[i]
		}
	}
	return nil
}

// ExtractNamedGroups extracts named groups from a value using the param's pattern
// Returns a map of group names to their captured values
func (p *RepverParam) ExtractNamedGroups(value string) (map[string]string, error) {
	re, err := regexp.Compile(p.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile param pattern: %w", err)
	}

	matches := re.FindStringSubmatch(value)
	if matches == nil {
		return nil, fmt.Errorf("value '%s' does not match pattern '%s'", value, p.Pattern)
	}

	result := make(map[string]string)
	names := re.SubexpNames()
	for i, name := range names {
		if i > 0 && name != "" && i < len(matches) {
			result[name] = matches[i]
		}
	}

	return result, nil
}

// ValidateValue validates that a value matches the param's pattern
func (p *RepverParam) ValidateValue(value string) error {
	re, err := regexp.Compile(p.Pattern)
	if err != nil {
		return fmt.Errorf("failed to compile param pattern: %w", err)
	}

	if !re.MatchString(value) {
		return fmt.Errorf("value '%s' does not match pattern '%s'", value, p.Pattern)
	}

	return nil
}

// ApplyTransform applies the transform template using extracted named groups
// If transform is empty, returns the original value unchanged
func ApplyTransform(transform string, extractedGroups map[string]string) string {
	if transform == "" {
		return ""
	}

	result := transform
	for key, val := range extractedGroups {
		result = strings.ReplaceAll(result, "{{"+key+"}}", val)
	}

	return result
}

// GetTransformParamName extracts the param name that is referenced in a transform template
// Returns the first param name found in the transform, or empty string if none found
func (t *RepverTarget) GetTransformParamNames() []string {
	if t.Transform == "" {
		return nil
	}

	// Find all {{name}} patterns in the transform
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(t.Transform, -1)

	var names []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			names = append(names, match[1])
			seen[match[1]] = true
		}
	}

	return names
}
