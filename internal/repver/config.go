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

type RepverCommand struct {
	// Name of the command
	Name string `yaml:"name"`
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
