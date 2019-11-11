package er

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

var stacktraceDisabled = []string{"No stack, ENABLE_STACKTRACE not set"}

type err struct {
	e      error
	bstack []byte
	stack  []string
}

type R interface {
	Message() string
	Stack() []string
	String() string
	Wrapped0() error
	SetWrapped(e error)
	Native() error
}

func (e err) Stack() []string {
	if e.stack == nil {
		if e.bstack != nil {
			e.stack = strings.Split(string(e.bstack), "\n")
		} else {
			e.stack = stacktraceDisabled
		}
	}
	return e.stack
}

func (e err) Message() string {
	return e.e.Error()
}

func (e err) String() string {
	if e.bstack != nil {
		fmt.Sprintf("%s\n%s", e.e.Error(), strings.Join(e.Stack(), "\n"))
	}
	return e.e.Error()
}

func (e err) Wrapped0() error {
	return e.e
}

func (e err) SetWrapped(ee error) {
	e.e = ee
}

func (e err) Native() error {
	return errors.New(e.String())
}

func captureStack() []byte {
	if "" == os.Getenv("ENABLE_STACKTRACE") {
		return nil
	}
	return debug.Stack()
}

func Wrapped(err R) error {
	if err == nil {
		return nil
	}
	return err.Wrapped0()
}

func New(s string) R {
	return err{
		e:      errors.New(s),
		bstack: captureStack(),
	}
}

func Errorf(format string, a ...interface{}) R {
	return err{
		e:      fmt.Errorf(format, a...),
		bstack: captureStack(),
	}
}

func E(e error) R {
	if e == nil {
		return nil
	}
	return err{
		e:      e,
		bstack: captureStack(),
	}
}
