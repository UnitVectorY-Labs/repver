package repver

import "testing"

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
		{"nested named groups", `(?P<first>(?P<inner>\d+))`, true},

		// Invalid cases:
		{"first group is unnamed", `(\d+)-(?P<second>\w+)`, false},
		{"second group is unnamed", `(?P<first>\d+)-(\w+)`, false},
		{"two unnamed capturing groups", `(a(b))`, false},
		{"second capturing group unnamed", `(?P<first>\d+)(\w+)`, false},
		{"invalid regex syntax", `(?P<first\d+`, false},
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
