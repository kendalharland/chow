package chow

import (
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
	Run(stepName string, s StepProvider) StepResult
}

// Runnable is the client application. This should be passed to Main().
type Runnable func(Runner)

// StepProvider generates Steps.
//
// StepProviders are passed to Runner.Run to execute shell commands.
type StepProvider interface {
	Create() Step
}

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

// StepLog is a log-entry for a step invocation.
//
// This is logged to the console in production and serialized into an expectation file
// during tests.
type StepLog struct {
	StepName   string     `json:"step_name"`
	Step       Step       `json:"step"`
	StepResult StepResult `json:"result"`
}

// SelfProvider adapts a Step as a StepProvider.
type SelfProvider Step

func (s *SelfProvider) Create() Step {
	return Step(*s)
}

// NoArgProvider adapts the name of a shell command as a StepProvider.
//
// Example Usage:
//
//    cmd := NoArgProvider('true')
//    runner.Run(cmd)
type NoArgProvider string

func (c NoArgProvider) Create() Step {
	return Step{Command: []string{string(c)}}
}
