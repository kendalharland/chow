# chow

A framework for writing cross-platform, testable automation scripts.

[![CircleCI](https://circleci.com/gh/kharland/chow.svg?style=svg&circle-token=ebc3a281a614ce8198e0213295e4e2258cdcc7b0)](https://circleci.com/gh/kharland/chow)
[![codecov](https://codecov.io/gh/kharland/chow/branch/master/graph/badge.svg?token=eTT4V04m1C)](https://codecov.io/gh/kharland/chow)
[![Documentation](https://godoc.org/github.com/kharland/chow?status.svg)](http://godoc.org/github.com/kharland/chow)

![chow-logo](assets/chow-logo.png)


## Overview
We often need to write complex scripts that build, test & depoy software on multiple
platforms.  Plain Python/Go scripts can be hard to maintain as they grow and more people
join the project, and many shell environments are not supported on both Unix and Windows
systems.

Chow simplifies the process of writing and maintaining these scripts;  Chow tests generate
"expectation" files, which contain JSON representations of the set of commands that are
expected to execute when the script is run in production.  Expectation files should be
checked into your source tree and compared whenever changes are made to the script.  In
production, chow will actually execute those commands.

### Installation

```sh
go get -u go.kendal.io/chow
```

### Examples

#### Hello World

The following program calls the "echo" executable on the current PATH to print some text.

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
            // Command line arguments
            Args: []string{"-foo", "bar", "-baz", "bang"},
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

For more examples see the [chow-examples] project.

For an more in-depth user guide, see the [Wiki]

[chow-examples]: https://github.com/kharland/chow-examples
[Wiki]: https://github.com/kharland/chow/wiki