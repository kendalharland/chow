// A basic hello world test to verify that the framework can execute binaries.
package main

import (
	"go.kendal.io/chow"
)

func main() {
	chow.Main(RunSteps)
}

func RunSteps(r chow.Runner) {
	r.Run("echo hello_world", chow.Step{
		Command: []string{"echo", "Hello, World!"},
	})
	r.Run("echo hello_world", chow.Step{
		Command: []string{"echo", "Hello, Again!"},
	})
}
