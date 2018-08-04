package chow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock is used to mock a step invocation.
//
// Step specifies the name of the step to mock.  Return is the step result to return.
// Mocks should be installed from a TestBuilder, like so:
//
//     config.NewTest(func(b *TestBuilder) {
//        b.Name("test_name")
//        b.Mock(Mock{
//          Step: "step_name",
//          Return: StepResult{
//              Stdout: "mocked output",
//          }
//        })
//     })
type Mock struct {
	Step   string
	Result StepResult
}

// TestCase specifies how an application should be exected in testing.
//
// Name is the name of this test case, and will be embedded in the name of the expecation
// file. Command-line flags can be set with `Args`.  The output of individual steps can
// be mocked via `Mocks`.   When two mocks match a given step, the one that was added the
// added the earliest is used.  For debugging or streaming, you may substitute any
// io.Writer for `Output`.  If a value is given, no expectation file will be generated for
// this test case.
type TestCase struct {
	Name   string
	Args   []string
	Mocks  []Mock
	Output io.Writer
}

// TestConfig is used to run a test suite for an application.
//
// Runnable is the application's implementation.
type TestConfig struct {
	Runnable Runnable
}

// Run implements Runner.
func (c *TestConfig) Run(t *testing.T, tc TestCase) {
	if t.Name() == "" {
		panic(errors.New("test case name cannot be empty"))
	}

	if tc.Output == nil {
		tc.Output = createExpectationFile(t)
	}

	runner := &testRunner{Mocks: tc.Mocks}
	c.Runnable(runner)

	encoder := json.NewEncoder(tc.Output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(runner.stepLogs); err != nil {
		panic(fmt.Errorf("failed to marshal expectation: %v", err))
	}
}

// TODO: Fix panics in this function.
func createExpectationFile(t *testing.T) *os.File {
	// Generate test directory if it doesn't exist.
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("could not get current directory: %v", err))
	}

	// Generate output directory.
	outDir := filepath.Join(cwd, "expectations")
	if err := os.MkdirAll(outDir, os.FileMode(os.O_APPEND)); err != nil {
		panic(fmt.Errorf("could not create %s: %v", outDir, err))
	}

	// Generate output file.
	basename := strings.Replace(t.Name(), "/", ".", -1) + ".expected.json"
	outPath := filepath.Join(outDir, basename)
	outFile, err := os.Create(outPath)
	if err != nil {
		panic(fmt.Errorf("could not create %s: %v", outPath, err))
	}

	return outFile
}

// Skips a test when running on CI, since we can't do file I/O.
func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping example test in CI environment")
	}
}
