package chow

import (
	"errors"
	"flag"
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

type TestConfig struct {
	Runnable Runnable
	Flags    *flag.FlagSet
}

// TODO: Define better logging functions and use those instead of these panics.
func (c *TestConfig) Run(tc TestCase) {
	if tc.Name == "" {
		panic(errors.New("name cannot be empty"))
	}
	if tc.Output == nil {
		panic(errors.New("test output writer cannot be nil"))
	}

	runner := &testRunner{
		Mocks:   tc.Mocks,
		stepLog: &JSONStepLogWriter{tc.Output},
	}

	// TODO: Set input args.
	c.Runnable(runner)
}

type TestCase struct {
	Name   string
	Args   []string
	Mocks  []Mock
	Output io.Writer
}

// TODO: Fix panics in this function.
func CreateExpectationFile(t *testing.T) *os.File {
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

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping example test in CI environment")
	}
}
