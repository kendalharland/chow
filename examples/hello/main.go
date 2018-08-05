// A basic hello world test to verify that the framework can execute binaries.
package main

import (
	"flag"

	"go.kendal.io/chow"
)

var name string

func main() {
	flags := flag.FlagSet{}
	flags.StringVar(&name, "name", "Anonymous", "The user to greet")
	chow.Main(RunSteps, &flags)
}

func RunSteps(r chow.Runner) {
	r.Run("echo "+name, chow.Step{
		Command: []string{"echo", "Hello, " + name + "!"},
	})
}
