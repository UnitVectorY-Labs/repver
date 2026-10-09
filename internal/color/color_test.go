package color

import (
	"testing"
)

func TestWrapEnabled(t *testing.T) {
	Enabled = true
	got := Red("hello")
	want := "\033[31mhello\033[0m"
	if got != want {
		t.Errorf("Red(\"hello\") = %q, want %q", got, want)
	}
}

func TestWrapDisabled(t *testing.T) {
	Enabled = false
	defer func() { Enabled = true }()

	got := Red("hello")
	want := "hello"
	if got != want {
		t.Errorf("Red(\"hello\") with Enabled=false = %q, want %q", got, want)
	}
}

func TestAllColors(t *testing.T) {
	Enabled = true

	tests := []struct {
		name string
		fn   func(string) string
		code string
	}{
		{"Red", Red, "\033[31m"},
		{"Green", Green, "\033[32m"},
		{"Yellow", Yellow, "\033[33m"},
		{"Cyan", Cyan, "\033[36m"},
		{"Bold", Bold, "\033[1m"},
		{"BoldRed", BoldRed, "\033[1;31m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn("test")
			want := tc.code + "test" + "\033[0m"
			if got != want {
				t.Errorf("%s(\"test\") = %q, want %q", tc.name, got, want)
			}
		})
	}
}

func TestFormatFunctions(t *testing.T) {
	Enabled = true

	got := Redf("count: %d", 42)
	want := "\033[31mcount: 42\033[0m"
	if got != want {
		t.Errorf("Redf = %q, want %q", got, want)
	}

	got = Greenf("hello %s", "world")
	want = "\033[32mhello world\033[0m"
	if got != want {
		t.Errorf("Greenf = %q, want %q", got, want)
	}
}

func TestDisabledFormatFunctions(t *testing.T) {
	Enabled = false
	defer func() { Enabled = true }()

	got := Redf("count: %d", 42)
	want := "count: 42"
	if got != want {
		t.Errorf("Redf with Enabled=false = %q, want %q", got, want)
	}

	got = Boldf("hello %s", "world")
	want = "hello world"
	if got != want {
		t.Errorf("Boldf with Enabled=false = %q, want %q", got, want)
	}
}

func TestYellowfFormatFunction(t *testing.T) {
	origEnabled := Enabled
	defer func() { Enabled = origEnabled }()
	Enabled = true

	got := Yellowf("value: %d", 10)
	want := "\033[33mvalue: 10\033[0m"
	if got != want {
		t.Errorf("Yellowf = %q, want %q", got, want)
	}
}

func TestCyanfFormatFunction(t *testing.T) {
	origEnabled := Enabled
	defer func() { Enabled = origEnabled }()
	Enabled = true

	got := Cyanf("path: %s", "/tmp")
	want := "\033[36mpath: /tmp\033[0m"
	if got != want {
		t.Errorf("Cyanf = %q, want %q", got, want)
	}
}

func TestAllFormatFunctionsDisabled(t *testing.T) {
	origEnabled := Enabled
	defer func() { Enabled = origEnabled }()
	Enabled = false

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Redf", Redf("test %d", 1), "test 1"},
		{"Greenf", Greenf("test %s", "ok"), "test ok"},
		{"Yellowf", Yellowf("test %s", "warn"), "test warn"},
		{"Cyanf", Cyanf("test %s", "info"), "test info"},
		{"Boldf", Boldf("test %s", "bold"), "test bold"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s with Enabled=false = %q, want %q", tc.name, tc.got, tc.expected)
			}
		})
	}
}
