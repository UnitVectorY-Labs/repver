package repver

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load loads a configuration from the specified file path
func Load(filePath string) (*RepverConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Parse(string(data))
}

// Parse parses YAML content into a RepverConfig structure
func Parse(yamlContent string) (*RepverConfig, error) {
	config := &RepverConfig{}
	err := yaml.Unmarshal([]byte(yamlContent), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
