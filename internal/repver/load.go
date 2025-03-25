package repver

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads a YAML file from the specified path and parses it into a RepverConfig structure
func Load(filePath string) (*RepverConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Parse(string(data))
}

// Parse takes a YAML string and parses it into a RepverConfig structure
func Parse(yamlContent string) (*RepverConfig, error) {
	config := &RepverConfig{}
	err := yaml.Unmarshal([]byte(yamlContent), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
