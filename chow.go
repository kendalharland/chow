package chow

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Main runs the client application, and should be called immediately in main().
//
// Example Usage:
//
//     func main() {
//         Main(func(r Runner) {
//             ...
//         })
//     }
func Main(r Runnable) error {
	return runRunnable(r, os.Stdout, os.Stderr)
}

// Runner executes steps within an application.
//
// Example Usage:
//
//     result := runner.Run(common.Echo{"Hello World"})
//     fmt.Println("Stdout:", result.Stdout)
//     fmt.Println("Stderr:", result.Stderr)
//     fmt.Println("Exit code:", result.ExitCode)
type Runner interface {
	Run(stepName string, s Step) StepResult
}

// Runnable is the client application. This should be passed to Main().
type Runnable func(Runner)

// Step describes what what will happen in a step invocation.
//
// Command is the shell command to run.  Outputs is an optional list of paths that will
// exist after Command is run.  In production, an error is generated if any of the paths
// do not exist after running.  In tests, warnings are issued if a client attempts to read
// from a path that has not been declared by a previously executed step.
type Step struct {
	Command Command `json:"command"`
	Outputs Outputs `json:"outputs"`
}

// Command is an alias for a step's shell commands.
//
// See the README for guidelines about using paths in command arguments.
type Command = []string

// Outputs is an alias for a step's declared outputs.
//
// See the README for guidelines about using paths in outputs.
type Outputs = []string

// StepResult describes the output of a step execution.
type StepResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func Placeholder(contents string) string {
	if placeholders == nil {
		placeholders = make(map[string]io.WriteCloser)
	}

	id := fmt.Sprintf("%d", len(placeholders))
	tempFile, err := ioutil.TempFile("", id)
	if err != nil {
		panic(err)
	}

	_, err = tempFile.Write([]byte(contents))
	if err != nil {
		panic(err)
	}

	placeholders[id] = tempFile
	return "//ph/" + id
}
