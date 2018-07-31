package chow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

func TestProdRunner_Run(t *testing.T) {
	// Expects that executing the given step produces the given step log.  Results in a
	// test failure if the actual log differs.
	expectOutput := func(t *testing.T, step Step, expected stepLog) {
		stderr := new(bytes.Buffer)
		startDir, _ := os.Getwd()

		// Run the program
		logs := &MemoryLogWriter{}
		runner := &prodRunner{
			startDir: startDir,
			stdout:   os.Stdout,
			stderr:   stderr,
			stepLog:  logs,
		}
		runner.Run("", step)

		// Verify step log is correct.
		actual := logs.Entries[0]
		expectLogsEqual(t, expected, actual)
	}

	// Expects that running the given steps produces an error. Results in a test failure
	// if no error is produced.
	expectError := func(t *testing.T, steps []Step) {
		if runRunnable(func(r Runner) {
			for _, step := range steps {
				r.Run("", step)
			}
		}, os.Stdout, os.Stderr) == nil {
			t.Fatalf("expected an error. got nil")
		}
	}

	// Test setup.
	echoPath := buildTestBinary(t, "echo")
	catPath := buildTestBinary(t, "cat")

	// Test teardown.
	defer func() {
		os.RemoveAll(echoPath)
		os.RemoveAll(catPath)
	}()

	t.Run("should run a command", func(t *testing.T) {
		input := Step{
			Command: []string{echoPath, "Hello, World!"},
		}

		output := stepLog{
			Step: Step{
				Command: []string{echoPath, "Hello, World!"},
			},
			StepResult: StepResult{
				Stdout: "Hello, World!",
			},
		}

		expectOutput(t, input, output)
	})

	t.Run("should convert path containing start dir in command", func(t *testing.T) {
		startDir, _ := os.Getwd()
		expectedPath := filepath.FromSlash(startDir + "/path/to/file")

		input := Step{
			Command: []string{echoPath, "//path/to/file"},
		}

		output := stepLog{
			Step: Step{
				Command: []string{echoPath, expectedPath},
			},
			StepResult: StepResult{
				Stdout: expectedPath,
			},
		}

		expectOutput(t, input, output)
	})

	t.Run("should convert path containing cwd in command if cwd == start dir", func(t *testing.T) {
		cwd, _ := os.Getwd()
		expectedPath := filepath.FromSlash(cwd + "/path/to/file")

		input := Step{
			Command: []string{echoPath, "//CWD/path/to/file"},
		}

		output := stepLog{
			Step: Step{
				Command: []string{echoPath, expectedPath},
			},
			StepResult: StepResult{
				Stdout: expectedPath,
			},
		}

		expectOutput(t, input, output)

	})

	t.Run("should convert placeholders", func(t *testing.T) {
		placeholder := Placeholder("def")
		placeholderID := strings.SplitN(placeholder, "//ph/", 2)[1]
		placeholderBackingFile := placeholders[placeholderID].(*os.File).Name()

		input := Step{
			Command: []string{catPath, placeholder},
		}

		output := stepLog{
			Step: Step{
				Command: []string{catPath, placeholderBackingFile},
			},
			StepResult: StepResult{
				Stdout: "def",
			},
		}

		expectOutput(t, input, output)
	})

	t.Run("should convert path containing cwd in command if cwd != start dir", func(t *testing.T) {
		// TODO: implement
	})

	t.Run("should not convert absolute path in command", func(t *testing.T) {
		input := Step{
			Command: []string{echoPath, "/absolute/path"},
		}

		output := stepLog{
			Step: Step{
				Command: []string{echoPath, filepath.FromSlash("/absolute/path")},
			},
			StepResult: StepResult{
				Stdout: filepath.FromSlash("/absolute/path"),
			},
		}

		expectOutput(t, input, output)
	})

	// TODO: Add tests for path conversion in outputs.

	t.Run("should error if a command fails to produce outputs", func(t *testing.T) {
		expectError(t, []Step{{
			Command: []string{echoPath},
			Outputs: []string{"missing.txt"},
		}})
	})

	t.Run("should error if a binary does not exist", func(t *testing.T) {
		expectError(t, []Step{{
			Command: []string{"i_dont_exist"},
		}})
	})

}

func TestTestRunner_Run(t *testing.T) {
	// Expects that executing the given steps w/ the given mocks produces the given step
	// log.  Results in a test failure if the actual log differs.
	expectOutput := func(t *testing.T, step []Step, mocks []Mock, expected []stepLog) {
		// Execute the program.
		output := &MemoryLogWriter{}
		runner := &testRunner{Mocks: mocks, stepLog: output}
		for i := range step {
			runner.Run("step_"+fmt.Sprint(i), step[i])
		}

		// Verify the results.
		actual := output.Entries
		if len(expected) != len(actual) {
			t.Fatalf("expected %v Got %v", expected, actual)
		}

		for i := 0; i < len(expected); i++ {
			expectLogsEqual(t, expected[i], actual[i])
		}
	}

	t.Run("step output should be mocked", func(t *testing.T) {
		t.Run("when a mock step name matches", func(t *testing.T) {
			step := Step{Command: []string{"command"}}
			inputs := []Step{step, step, step}

			mocks := []Mock{{
				Step: "step_1",
				Result: StepResult{
					Stdout:   "mocked stdout",
					Stderr:   "mocked stderr",
					ExitCode: 31,
				},
			}}

			result := []stepLog{{
				StepName:   "step_0",
				Step:       step,
				StepResult: StepResult{},
			}, {
				StepName: "step_1",
				Step:     step,
				StepResult: StepResult{
					Stdout:   "mocked stdout",
					Stderr:   "mocked stderr",
					ExitCode: 31,
				},
			}, {
				StepName:   "step_2",
				Step:       step,
				StepResult: StepResult{},
			}}

			expectOutput(t, inputs, mocks, result)
		})
	})

	t.Run("step output should be empty", func(t *testing.T) {
		t.Run("when there are no mocks", func(t *testing.T) {
			inputs := []Step{{
				Command: []string{"command", "arg1", "arg2"},
				Outputs: []string{"output"},
			}}

			result := []stepLog{{
				StepName: "step_0",
				Step:     inputs[0],
			}}

			expectOutput(t, inputs, []Mock{}, result)
		})
		t.Run("when a mock step name does not match", func(t *testing.T) {
			inputs := []Step{{
				Command: []string{"command", "arg1", "arg2"},
				Outputs: []string{"output"},
			}}

			result := []stepLog{{
				StepName: "step_0",
				Step:     inputs[0],
			}}

			expectOutput(t, inputs, []Mock{}, result)
		})
	})

	// TODO: Add tests for when there are mutliple matching mocks
	// TODO: Add tests showing that multiple mocks can be used with multiple steps.
}

func TestJSONStepWriter(t *testing.T) {
	buffer := new(bytes.Buffer)
	w := &JSONStepLogWriter{buffer}

	input := stepLog{
		Step: Step{
			Command: []string{"a"},
			Outputs: []string{"b"},
		},
		StepResult: StepResult{
			Stdout:   "c",
			Stderr:   "d",
			ExitCode: 3,
		},
	}

	w.Write(input)
	var output stepLog
	if err := json.Unmarshal(buffer.Bytes(), &output); err != nil {
		t.Fatal(err)
	}

	expectLogsEqual(t, input, output)
}

func TestTestWorkflow(t *testing.T) {
	cfg := TestConfig{
		Runnable: func(r Runner) {
			r.Run("echo", Step{Command: []string{"echo"}})
		},
		Flags: nil,
	}

	t.Run("default", func(t *testing.T) {
		var buf bytes.Buffer
		cfg.Run(t, TestCase{
			Name:   t.Name(),
			Output: &buf,
		})

		expectedLog, _ := json.MarshalIndent(
			stepLog{
				StepName: "echo",
				Step: Step{
					Command: []string{"echo"},
				},
				StepResult: StepResult{},
			},
			"", "    ")

		expected := strings.TrimSpace(string(expectedLog))
		output := strings.TrimSpace(buf.String())

		if output != expected {
			t.Fatalf("expected\n%s\nbut got\n%s\n", expected, output)
		}
	})
}

func TestCreateExpectationFile(t *testing.T) {
	skipCI(t)

	t.Run("nested", func(t *testing.T) {
		cwd, _ := os.Getwd()
		expected := filepath.FromSlash(cwd + "/expectations/TestCreateExpectationFile.nested.expected.json")
		file := createExpectationFile(t)
		if file.Name() != expected {
			t.Fatalf("expected path %s\nbut got %s\n", file.Name(), expected)
		}
		file.Close()
		os.RemoveAll(expected)
	})
}

func buildTestBinary(t *testing.T, tool string) string {
	cmd := exec.Command("go", "build", "go.kendal.io/chow/test_binaries/"+tool)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}

	path := tool
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	return path
}

func expectLogsEqual(t *testing.T, expected, actual stepLog) {
	if !stepLogsEqual(expected, actual) {
		b := new(bytes.Buffer)
		diffs := pretty.Diff(expected, actual)

		fmt.Fprintln(b, "actual output differs from expected at:")
		for _, diff := range diffs {
			fmt.Fprintf(b, "-%s\n", diff)
		}
		fmt.Fprintf(b, "expected: %v\nactual: %v",
			pretty.Sprint(expected), pretty.Sprint(actual))

		t.Errorf(b.String())
	}
}

func stepLogsEqual(a, b stepLog) bool {
	return a.StepName == b.StepName &&
		reflect.DeepEqual(a.Step, b.Step) &&
		strings.TrimSpace(a.StepResult.Stdout) == strings.TrimSpace(a.StepResult.Stdout) &&
		strings.TrimSpace(a.StepResult.Stderr) == strings.TrimSpace(b.StepResult.Stderr) &&
		a.StepResult.ExitCode == b.StepResult.ExitCode
}

type MemoryLogWriter struct {
	Entries []stepLog
}

func (w *MemoryLogWriter) Write(s stepLog) error {
	w.Entries = append(w.Entries, s)
	return nil
}
