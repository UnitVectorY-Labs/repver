package repver

import (
	"flag"
	"fmt"
	"os"
)

var Debug bool

var Command string

// Func load command line parameters to bool
func ParseParams() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	command := flag.String("command", "", "Command to execute")

	flag.Parse()

	Debug = *debug
	Command = *command
}

func Debugln(format string, args ...interface{}) {
	if Debug {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
