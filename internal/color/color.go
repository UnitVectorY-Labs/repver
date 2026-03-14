package color

import (
	"fmt"
	"os"
)

// Enabled controls whether color output is active. It defaults to true
// and is set to false when the --no-color flag is passed or the NO_COLOR
// environment variable is set (any value, including empty).
var Enabled = true

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		Enabled = false
	}
}

// ANSI escape codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	boldRed = "\033[1;31m"
)

// wrap returns s wrapped with the given ANSI code when color is enabled.
func wrap(code, s string) string {
	if !Enabled {
		return s
	}
	return code + s + reset
}

// Red returns s colored red.
func Red(s string) string { return wrap(red, s) }

// Green returns s colored green.
func Green(s string) string { return wrap(green, s) }

// Yellow returns s colored yellow.
func Yellow(s string) string { return wrap(yellow, s) }

// Cyan returns s colored cyan.
func Cyan(s string) string { return wrap(cyan, s) }

// Bold returns s in bold.
func Bold(s string) string { return wrap(bold, s) }

// BoldRed returns s in bold red.
func BoldRed(s string) string { return wrap(boldRed, s) }

// Redf returns a red formatted string.
func Redf(format string, a ...interface{}) string { return Red(fmt.Sprintf(format, a...)) }

// Greenf returns a green formatted string.
func Greenf(format string, a ...interface{}) string { return Green(fmt.Sprintf(format, a...)) }

// Yellowf returns a yellow formatted string.
func Yellowf(format string, a ...interface{}) string { return Yellow(fmt.Sprintf(format, a...)) }

// Cyanf returns a cyan formatted string.
func Cyanf(format string, a ...interface{}) string { return Cyan(fmt.Sprintf(format, a...)) }

// Boldf returns a bold formatted string.
func Boldf(format string, a ...interface{}) string { return Bold(fmt.Sprintf(format, a...)) }
