package repver

import (
	"flag"
	"fmt"
	"os"
)

var Debug bool
var DryRun bool
var UserCommand string

// ParseParams initializes the command-line flags and sets the global variables
func ParseParams() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	command := flag.String("command", "", "Command to execute")
	dryRun := flag.Bool("dry-run", false, "Dry run mode - shows changes without applying them")

	flag.Parse()

	Debug = *debug
	DryRun = *dryRun
	UserCommand = *command
}

// Debugln prints debug messages to stderr if Debug mode is enabled
func Debugln(format string, args ...interface{}) {
	if Debug {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
