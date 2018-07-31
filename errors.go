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
	panic(formatError(fatal, err, step))
}

func logWarning(message string, step Step) {
	fmt.Fprint(os.Stderr, formatError(warn, errors.New(message), step))
}

func formatError(level logLevel, err error, step Step) error {
	b := new(bytes.Buffer)
	fmt.Fprintln(b)
	fmt.Fprintf(b, "chow: %s: %v\n", level, err)
	if !reflect.DeepEqual(step, Step{}) {
		fmt.Fprintf(b, "IN STEP: %#v\n", step)
	}
	fmt.Fprintln(b)
	return errors.New(b.String())
}
