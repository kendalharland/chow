# chow

A framework for writing cross-platform, testable automation scripts.

[![CircleCI](https://circleci.com/gh/kharland/chow.svg?style=svg&circle-token=ebc3a281a614ce8198e0213295e4e2258cdcc7b0)](https://circleci.com/gh/kharland/chow)
[![codecov](https://codecov.io/gh/kharland/chow/branch/master/graph/badge.svg?token=eTT4V04m1C)](https://codecov.io/gh/kharland/chow)
[![Documentation](https://godoc.org/github.com/kharland/chow?status.svg)](http://godoc.org/github.com/kharland/chow)

![chow-logo](assets/chow-logo.png)


## Overview
We often need to write complex scripts that build, test & depoy software on multiple
platforms.  Bash is not an option if you need support for Windows, and plain Python/Go
scripts can be hard to maintain as they grow, and more people work on the project.  Every
time an if-statement is added to a script, you have to make sure the commands executed in
the new conditional branch are what you expect, for each platform.

Chow simplifies the process of writing and maintaining these scripts:
* In testing, chow generates a serialized representation of the commands that your script
  would execute.
* In production, chow executes that list of calls.

## Usage

### Writing a script

The following is a program that calls the "echo" executable on the current PATH to print
some text.

```go
package main

import "go.kendal.io/chow"

func main() {
	chow.Main(RunSteps)
}

func RunSteps(r chow.Runner) {
	r.Run("echo hello_world", chow.Step{
        Command: []string{"echo", "Hello, World!"},
    })
}
```

### Testing
Chow tests generate "expectation" files, which contain JSON representations of the set of
commands that are expected to execute when the script is run in production.  Expectation
files should be checked into your source tree and compared whenever changes are made to
the script.  If the expecation is the same, you can be sure your script will execute
correctly in production.

Chow tests are written like normal go tests, and can be added alongside any unit tests you
decide to write.  We can write a basic test for the above code like so:

```go
package main

import (
	"testing"
	"go.kendal.io/chow"
)

func TestMain(t *testing.T) {
    cfg := chow.TestConfig{Runnable: RunSteps}

	t.Run("default", func(t *testing.T) {
		cfg.Run(chow.TestCase{
            Name:   t.Name(),

            // Command line arguments
            Args: []string{"-foo", "bar", "-baz", "bang"},

            // Where to write the expectation.
            // Replace with any io.Writer to stream elsewhere.
			Output: chow.CreateExpectationFile(t),
		})
	})
}
```

This test produces the expectation file:

```json
{
    "step_name": "echo hello_world",
    "step": {
        "command": [
            "echo",
            "Hello, World!"
        ],
        "outputs": null
    },
    "result": {
        "stdout": "",
        "stderr": "",
        "exit_code": 0
    }
}
```

For more examples see the [chow-examples] project

[chow-examples]: https://github.com/kharland/chow-examples
