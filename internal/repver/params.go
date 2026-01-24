package repver

import (
	"flag"
	"fmt"
	"os"
)

var Debug bool
var DryRun bool
var UserCommand string
var Exists bool

// ParseParams initializes the command-line flags and sets the global variables
func ParseParams() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	command := flag.String("command", "", "Command to execute")
	dryRun := flag.Bool("dry-run", false, "Dry run mode - shows changes without applying them")
	exists := flag.Bool("exists", false, "Check whether .repver exists and contains the specified command")

	flag.Parse()

	Debug = *debug
	DryRun = *dryRun
	UserCommand = *command
	Exists = *exists
}

// Debugln prints debug messages to stderr if Debug mode is enabled
func Debugln(format string, args ...interface{}) {
	if Debug {
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
