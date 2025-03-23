package repver

import (
	"fmt"
	"regexp"
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
