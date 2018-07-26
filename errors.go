package chow

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
)

type logLevel string

const (
	fatal logLevel = "FATAL"
	warn           = "WARNING"
)

func logFatal(message string, err error) {
	err = fmt.Errorf("%s: %v", message, err.Error())
	panic(formatError(fatal, err))
}

func logWarning(message string) {
	fmt.Fprint(os.Stderr, formatError(warn, errors.New(message)))
}

func formatError(level logLevel, err error) error {
	b := new(bytes.Buffer)
	fmt.Fprintln(b)
	fmt.Fprintf(b, "chow: %s: %v\n", level, err)
	if !reflect.DeepEqual(currentStep, Step{}) {
		fmt.Fprintf(b, "         IN STEP: %#v\n", currentStep)
	}
	fmt.Fprintln(b)
	return errors.New(b.String())
}
