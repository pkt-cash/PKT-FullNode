package er

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
)

type err struct {
	e     error
	stack []byte
}

type R interface {
	String() string
	Wrapped() error
	Is(error) bool
}

func (e err) String() string {
	if e.stack != nil {
		fmt.Sprintf("%s\n%s", e.e.Error(), string(e.stack))
	}
	return e.e.Error()
}

func (e err) Wrapped() error {
	return e.e
}

func (e err) SetWrapped(ee error) {
	e.e = ee
}

func (e err) Is(er error) bool {
	return e.e == er
}

func CaptureStack() []byte {
	if "" == os.Getenv("ENABLE_STACKTRACE") {
		return nil
	}
	return debug.Stack()
}

func New(s string) R {
	return E(errors.New(s))
}

func Errorf(format string, a ...interface{}) R {
	return E(fmt.Errorf(format, a...))
}

func E(e error) R {
	if e == nil {
		return nil
	}
	return err{
		e:     e,
		stack: CaptureStack(),
	}
}
