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

// Runner executes Steps.
//
// Example Usage:
//
//     result := runner.Run("echo_hello_world", Step{
//         Command: []string{"echo", "Hello World"}),
//     })
//     fmt.Println("Stdout:", result.Stdout)
//     fmt.Println("Stderr:", result.Stderr)
//     fmt.Println("Exit code:", result.ExitCode)
type Runner interface {
	Run(stepName string, s Step) StepResult
}

// Runnable is the client application. This should be passed to Main().
type Runnable func(Runner)

// Step describes a shell command to run.
//
// Outputs is an optional list of paths that will exist after Command is run. In
// production, it is a fatal error if any of the paths do not exist after the
// step is run.  In tests, warnings are issued if a client attempts to read from
// a path that was not declared by any previous step.
type Step struct {
	Command []string `json:"command"`
	Outputs []string `json:"outputs"`
}

// StepResult describes the output of a step execution.
type StepResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Placeholder returns a unique ID that serves as a "placeholder" for a file.
//
// It's cumbersome to ensure that a program's various file and directory names
// do not clash, especially when using third party libraries.  Rather than
// creating filenames one's self, Placeholder should be used in both library and
// application code to produce unique filenames.  The ID is converted to a path
// automatically when passed in the Command or Outputs of a step.
//
// To read or write to a placeholder directly - for example, using
// ioutil.WriteFile or ioutil.ReadFile - you must first call PlaceholderPath to
// resolve the ID to its underlying filepath.
func Placeholder(contents string) string {
	if placeholders == nil {
		placeholders = make(map[string]io.WriteCloser)
	}

	id := fmt.Sprintf("%d", len(placeholders))
	tempFile, err := ioutil.TempFile("", id)
	if err != nil {
		panic(err)
	}

	if _, err = tempFile.Write([]byte(contents)); err != nil {
		panic(err)
	}

	placeholders[id] = tempFile
	return "//ph/" + id
}

// PlaceholderPath returns the filepath represented by the given placeholder ID.
func PlaceholderPath(id string) string {
	tempFile, ok := placeholders[id].(*os.File)
	if !ok {
		panic(fmt.Errorf("unkown placeholder ID: %v", id))
	}
	return tempFile.Name()
}
