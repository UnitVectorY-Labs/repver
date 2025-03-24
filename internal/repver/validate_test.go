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
