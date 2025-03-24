package repver

import (
	"flag"
	"fmt"
	"os"
)

var Debug bool
var DryRun bool
var UserCommand string

// Func load command line parameters to bool
func ParseParams() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	command := flag.String("command", "", "Command to execute")
	dryRun := flag.Bool("dryRun", false, "Dry run mode - shows changes without applying them")

	flag.Parse()

	Debug = *debug
	DryRun = *dryRun
	UserCommand = *command
}

func Debugln(format string, args ...interface{}) {
	if Debug {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
