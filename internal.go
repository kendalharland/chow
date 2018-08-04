package chow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

var placeholders map[string]io.WriteCloser

// stepLog describes a step invocation.
//
// This is logged to the console in production and serialized into an
// expectation file when testing.
type stepLog struct {
	StepName   string     `json:"step_name"`
	Step       Step       `json:"step"`
	StepResult StepResult `json:"result"`
}

func runRunnable(r Runnable, stdout io.Writer, stderr io.Writer) (err error) {
	// The framework will panic if any fatal errors occur. Recover from these panics so we
	// can report errors gracefully.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	startDir, err := os.Getwd()
	if err != nil {
		logFatal("failed to get working directory", err, Step{})
	}

	runner := &prodRunner{
		startDir: startDir,
		stdout:   stdout,
		stderr:   stderr,
	}

	// Run the program.
	r(runner)
	return
}

type prodRunner struct {
	currentStep Step
	startDir    string
	stdout      io.Writer
	stderr      io.Writer
	stepOutput  io.Writer
}

// Run implements Runner
func (r *prodRunner) Run(name string, step Step) StepResult {
	r.currentStep = step

	if err := r.convertAnyPaths(r.currentStep.Command); err != nil {
		logFatal("failed to convert paths in step command", err, r.currentStep)
	}
	if err := r.convertAnyPaths(r.currentStep.Outputs); err != nil {
		logFatal("failed to convert paths in step outputs", err, r.currentStep)
	}

	child := exec.Command(r.currentStep.Command[0], r.currentStep.Command[1:]...)

	// Capture stdout & stderr. We still want to print the child's output for easy
	// debugging, so we also stream to the current stdout and stderr.
	outWriter := &recordingWriter{Delegate: r.stdout}
	errWriter := &recordingWriter{Delegate: r.stderr}
	child.Stdout = outWriter
	child.Stderr = errWriter

	if err := child.Start(); err != nil {
		logFatal("failed to start child process", err, r.currentStep)
	}

	var exitCode int
	if err := child.Wait(); err != nil {
		exitCode = err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus()
	}

	// Ensure outputs exist, fail otherwise.
	var missingOutputs []string
	for _, output := range r.currentStep.Outputs {
		_, err := os.Stat(output)
		if err != nil && os.IsNotExist(err) {
			missingOutputs = append(missingOutputs, output)
		}
	}

	if len(missingOutputs) > 0 {
		err := fmt.Errorf("ouputs are missing: %#v", missingOutputs)
		logFatal("declared outputs missing after step execution", err, r.currentStep)
	}

	// Log the result
	stepLog := stepLog{
		StepName: name,
		Step:     r.currentStep,
		StepResult: StepResult{
			Stdout:   outWriter.String(),
			Stderr:   errWriter.String(),
			ExitCode: exitCode,
		},
	}

	encoder := json.NewEncoder(r.stepOutput)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stepLog); err != nil {
		logFatal("failed to log step", err, r.currentStep)
	}

	return stepLog.StepResult
}

// Converts the input path to an absolute path for the current platform.
func (r *prodRunner) convertAnyPaths(args []string) error {
	for i, p := range args {
		// Current working directory
		if strings.HasPrefix(p, "//cwd/") {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %v", err)
			}

			suffix := strings.SplitN(p, "//cwd/", 2)[1]
			args[i] = filepath.FromSlash(wd + "/" + suffix)
			continue
		}

		// Placeholder
		if strings.HasPrefix(p, "//ph/") {
			id := strings.SplitN(p, "//ph/", 2)[1]
			file := placeholders[id]
			// TODO: Find a way to close the file handle.
			args[i] = file.(*os.File).Name()
			continue
		}

		// Start dir
		if strings.HasPrefix(p, "///") {
			suffix := strings.SplitN(p, "///", 2)[1]
			r.startDir = strings.TrimRight(r.startDir, "/")
			args[i] = filepath.FromSlash(r.startDir + "/" + suffix)
			continue
		}

		// Ignore absolute paths, relative paths and non-path arguments.
	}

	return nil
}

// A Runner that records step invocations for testing.
//
// This generates an "expectation", which is a serialized chain of step logs
// representing the set of commands that would run in production.
type testRunner struct {
	Mocks      []Mock
	callCounts map[string]int
	stepLogs   []stepLog
}

// Run implements Runner
//
// This is called directly by the client's production code.
func (r *testRunner) Run(name string, step Step) StepResult {
	if r.callCounts == nil {
		r.callCounts = make(map[string]int)
	}

	// Record that this step has been called one more time.
	if i, ok := r.callCounts[name]; ok {
		name = fmt.Sprintf("%s %d", name, i)
	}
	r.callCounts[name]++

	// If there's a mock return value for the step, return it.  It's possible the user
	// registered multiple mocks in their test; In this case, the first one registered
	// wins because we search the list of mocks from 0...end.
	var stepResult StepResult
	for i, mock := range r.Mocks {
		if mock.Step == name {
			stepResult = mock.Result
			// Prevent the mock from matching other steps by removing it.
			r.Mocks = append(r.Mocks[:i], r.Mocks[i+1:]...)
			break
		}
	}

	r.stepLogs = append(r.stepLogs, stepLog{name, step, stepResult})
	return stepResult
}

func (*testRunner) registerPlaceholder(content string) string {
	return "[placeholder]"
}

type recordingWriter struct {
	Delegate io.Writer
	buf      bytes.Buffer
}

func (w *recordingWriter) Write(b []byte) (int, error) {
	if n, err := w.buf.Write(b); err != nil {
		return n, err
	}
	return w.Delegate.Write(b)
}

func (w *recordingWriter) String() string {
	return w.buf.String()
}
