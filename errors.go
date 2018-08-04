package chow

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
)

func logFatal(message string, err error, step Step) {
	err = fmt.Errorf("%s: %v", message, err.Error())
	formatted := formatError("FATAL", err, step)
	panic(formatted)
}

func logWarning(message string, step Step) {
	formatted := formatError("WARN", errors.New(message), step)
	fmt.Fprint(os.Stderr, formatted)
}

func formatError(level string, err error, step Step) error {
	b := new(bytes.Buffer)
	fmt.Fprintln(b)
	fmt.Fprintf(b, "chow: %s: %v\n", level, err)
	if !reflect.DeepEqual(step, Step{}) {
		fmt.Fprintf(b, "IN STEP: %#v\n", step)
	}
	fmt.Fprintln(b)
	return errors.New(b.String())
}
