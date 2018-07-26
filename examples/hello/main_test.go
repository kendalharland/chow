package main

import (
	"os"
	"testing"

	"go.kendal.io/chow"
)

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping example test in CI environment")
	}
}

func TestMain(t *testing.T) {
	skipCI(t)
	cfg := chow.TestConfig{Runnable: RunSteps}

	t.Run("default", func(t *testing.T) {
		cfg.Run(chow.TestCase{
			Name:   t.Name(),
			Output: chow.CreateExpectationFile(t),
		})
	})
}
