package repver

import "testing"

func TestValidateCommandName(t *testing.T) {
	tests := []struct {
		name    string
		command string
		valid   bool
	}{
		// Valid cases:
		{"one character", "a", true},
		{"two characters", "ab", true},
		{"mixed case", "AbcDef", true},
		{"mixed case with numbers", "AbcDef123", true},
		{"thirty characters", "abcdefghijklmnopqrstuvwxyz1234", true},

		// Invalid cases:
		{"empty string", "", false},
		{"thirty one characters", "abcdefghijklmnopqrstuvwxyz12345678901", false},
		{"special characters", "abc!@#$%^&*()", false},
		{"spaces", "abc def", false},
		{"dashes", "abc-def", false},
		{"underscores", "abc_def", false},
		{"mixed case with special characters", "Abc!@#$%^&*()", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCommandName(tc.command)
			if tc.valid && err != nil {
				t.Errorf("command: %q should be valid but got error: %v", tc.command, err)
			} else if !tc.valid && err == nil {
				t.Errorf("command: %q should be invalid but no error was returned", tc.command)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		// Valid cases:
		{"no groups", `^abc$`, true},
		{"two named groups", `^(?P<first>\d+)-(?P<second>\w+)$`, true},
		{"one non-capturing group, one named group", `^(?:\d+)-(?P<second>\w+)$`, true},

		// Invalid cases:
		{"first group is unnamed", `^(\d+)-(?P<second>\w+)$`, false},
		{"second group is unnamed", `^(?P<first>\d+)-(\w+)$`, false},
		{"nested named groups", `^(?P<first>(?P<inner>\d+))$`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePattern(tc.pattern)
			if (err == nil) != tc.valid {
				t.Errorf("pattern: %q, expected valid: %v, got error: %v", tc.pattern, tc.valid, err)
			}
		})
	}
}

func TestValidateNamedGroups(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		// Valid cases:
		{"no groups", `abc`, true},
		{"two named groups", `(?P<first>\d+)-(?P<second>\w+)`, true},
		{"one non-capturing group, one named group", `(?:\d+)-(?P<second>\w+)`, true},

		// Invalid cases:
		{"first group is unnamed", `(\d+)-(?P<second>\w+)`, false},
		{"second group is unnamed", `(?P<first>\d+)-(\w+)`, false},
		{"two unnamed capturing groups", `(a(b))`, false},
		{"second capturing group unnamed", `(?P<first>\d+)(\w+)`, false},
		{"invalid regex syntax", `(?P<first\d+`, false},
		{"nested named groups", `(?P<first>(?P<inner>\d+))`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNamedGroups(tc.pattern)
			if (err == nil) != tc.valid {
				t.Errorf("pattern: %q, expected valid: %v, got error: %v", tc.pattern, tc.valid, err)
			}
		})
	}
}

func TestValidateParam(t *testing.T) {
	tests := []struct {
		name  string
		param RepverParam
		valid bool
	}{
		// Valid cases:
		{
			"basic param",
			RepverParam{Name: "version", Pattern: `^.*$`},
			true,
		},
		{
			"semantic version param",
			RepverParam{Name: "version", Pattern: `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$`},
			true,
		},
		{
			"simple version param",
			RepverParam{Name: "v", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)$`},
			true,
		},

		// Invalid cases:
		{
			"empty name",
			RepverParam{Name: "", Pattern: `^.*$`},
			false,
		},
		{
			"empty pattern",
			RepverParam{Name: "version", Pattern: ""},
			false,
		},
		{
			"pattern without anchors",
			RepverParam{Name: "version", Pattern: `.*`},
			false,
		},
		{
			"invalid regex pattern",
			RepverParam{Name: "version", Pattern: `^[invalid$`},
			false,
		},
		{
			"name with special chars",
			RepverParam{Name: "my-param", Pattern: `^.*$`},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if (err == nil) != tc.valid {
				t.Errorf("param: %+v, expected valid: %v, got error: %v", tc.param, tc.valid, err)
			}
		})
	}
}

func TestValidateTransform(t *testing.T) {
	tests := []struct {
		name      string
		command   RepverCommand
		transform string
		valid     bool
	}{
		// Valid cases:
		{
			"transform with valid groups",
			RepverCommand{
				Name: "test",
				Params: []RepverParam{
					{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`},
				},
			},
			"{{major}}.{{minor}}",
			true,
		},
		{
			"transform with all groups",
			RepverCommand{
				Name: "test",
				Params: []RepverParam{
					{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`},
				},
			},
			"{{major}}.{{minor}}.{{patch}}",
			true,
		},

		// Invalid cases:
		{
			"transform with unknown group",
			RepverCommand{
				Name: "test",
				Params: []RepverParam{
					{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)$`},
				},
			},
			"{{major}}.{{minor}}.{{patch}}",
			false,
		},
		{
			"transform without params",
			RepverCommand{
				Name:   "test",
				Params: []RepverParam{},
			},
			"{{major}}.{{minor}}",
			false,
		},
		{
			"transform without placeholders",
			RepverCommand{
				Name: "test",
				Params: []RepverParam{
					{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)$`},
				},
			},
			"static-text",
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.command.validateTransform(tc.transform)
			if (err == nil) != tc.valid {
				t.Errorf("transform: %q, expected valid: %v, got error: %v", tc.transform, tc.valid, err)
			}
		})
	}
}

func TestParamExtractNamedGroups(t *testing.T) {
	tests := []struct {
		name           string
		param          RepverParam
		value          string
		expectedGroups map[string]string
		shouldError    bool
	}{
		{
			"extract semver groups",
			RepverParam{Name: "version", Pattern: `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$`},
			"1.26.0",
			map[string]string{"major": "1", "minor": "26", "patch": "0"},
			false,
		},
		{
			"extract two groups",
			RepverParam{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)$`},
			"2.10",
			map[string]string{"major": "2", "minor": "10"},
			false,
		},
		{
			"value does not match pattern",
			RepverParam{Name: "version", Pattern: `^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`},
			"invalid",
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups, err := tc.param.ExtractNamedGroups(tc.value)
			if tc.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				for k, v := range tc.expectedGroups {
					if groups[k] != v {
						t.Errorf("expected group %q = %q, got %q", k, v, groups[k])
					}
				}
			}
		})
	}
}

func TestParamValidateValue(t *testing.T) {
	tests := []struct {
		name        string
		param       RepverParam
		value       string
		shouldError bool
	}{
		{
			"valid semver",
			RepverParam{Name: "version", Pattern: `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$`},
			"1.26.0",
			false,
		},
		{
			"invalid semver",
			RepverParam{Name: "version", Pattern: `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$`},
			"1.26",
			true,
		},
		{
			"valid simple pattern",
			RepverParam{Name: "name", Pattern: `^[a-z]+$`},
			"hello",
			false,
		},
		{
			"invalid simple pattern",
			RepverParam{Name: "name", Pattern: `^[a-z]+$`},
			"Hello123",
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.ValidateValue(tc.value)
			if tc.shouldError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tc.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestApplyTransform(t *testing.T) {
	tests := []struct {
		name     string
		template string
		groups   map[string]string
		expected string
	}{
		{
			"simple transform",
			"{{major}}.{{minor}}",
			map[string]string{"major": "1", "minor": "26", "patch": "0"},
			"1.26",
		},
		{
			"full transform",
			"{{major}}.{{minor}}.{{patch}}",
			map[string]string{"major": "1", "minor": "26", "patch": "0"},
			"1.26.0",
		},
		{
			"transform with prefix",
			"v{{major}}.{{minor}}",
			map[string]string{"major": "2", "minor": "0"},
			"v2.0",
		},
		{
			"empty template",
			"",
			map[string]string{"major": "1"},
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ApplyTransform(tc.template, tc.groups)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
