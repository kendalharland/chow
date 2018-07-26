// A basic hello world test to verify that the framework can execute binaries.
package main

import (
	"runtime"

	"go.kendal.io/chow"
)

var suffix = ""

// TODO: Consider providing this as part of the library.
func init() {
	suffix = ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
}

func Echo(text string) chow.StepProvider {
	return &chow.SelfProvider{
		Command: []string{"test_programs/bin/echo" + suffix, text},
	}
}

func main() {
	chow.Main(RunSteps)
}

func RunSteps(r chow.Runner) {
	r.Run("ehco_hello_world", Echo("Hello, World!"))
}
