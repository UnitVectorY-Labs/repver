package repver

import (
	"fmt"
	"regexp"
	"strings"
)

type RepverConfig struct {
	// Array of commands
	Commands []RepverCommand `yaml:"commands"`
}

type RepverCommand struct {
	// Name of the command
	Name string `yaml:"name"`
	// Array of targets
	Targets []RepverTarget `yaml:"targets"`
	// Git command options
	GitOptions RepverGit `yaml:"git"`
}

type RepverGit struct {
	// Whether to create a new branch
	CreateBranch bool `yaml:"create_branch"`
	// Whether to delete the new branch after the command is executed (must be set to return to original branch)
	DeleteBranch bool `yaml:"delete_branch"`
	// New branch name
	BranchName string `yaml:"branch_name"`
	// Commit
	Commit bool `yaml:"commit"`
	// Commit message
	CommitMessage string `yaml:"commit_message"`
	// Whether to push changes
	Push bool `yaml:"push"`
	// The remote to push to
	Remote string `yaml:"remote"`
	// return to the original branch after the command is executed
	ReturnToOriginalBranch bool `yaml:"return_to_original_branch"`
}

type RepverTarget struct {
	// Path for the target
	Path string `yaml:"path"`
	// Pattern for the content of the target
	Pattern string `yaml:"pattern"`
}

func (c *RepverConfig) GetCommand(name string) (*RepverCommand, error) {
	for _, command := range c.Commands {
		if command.Name == name {
			return &command, nil
		}
	}
	return nil, fmt.Errorf("command %s not found", name)
}

// Get parameters across everything
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

func (g *RepverGit) BuildBranchName(vals map[string]string) string {
	branchName := g.BranchName
	for key, val := range vals {
		branchName = strings.ReplaceAll(branchName, "{{"+key+"}}", val)
	}

	return branchName
}

func (g *RepverGit) BuildCommitMessage(vals map[string]string) string {
	commitMessage := g.CommitMessage
	for key, val := range vals {
		commitMessage = strings.ReplaceAll(commitMessage, "{{"+key+"}}", val)
	}

	return commitMessage
}

func (g *RepverGit) GitOptionsSpecified() bool {
	return g.CreateBranch || g.DeleteBranch || g.Commit || g.Push || g.ReturnToOriginalBranch
}
