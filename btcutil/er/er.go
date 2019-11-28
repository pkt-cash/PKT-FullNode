package er

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

var stacktraceDisabled = []string{"No stack, ENABLE_STACKTRACE not set"}

// GenericErrorType is for packages with only one or two error codes
// which don't make sense having their own error type
var GenericErrorType = NewErrorType("er.GenericErrorType")

// ErrorCode is a code for identifying a particular type of fault.
// Error codes can have a numeric code identifier or they can not.
type ErrorCode struct {
	Detail string
	Number int
	Type   *ErrorType
}

type typedErr struct {
	messages []string
	errType  *ErrorType
	code     *ErrorCode
	err      R
}

// Error type

// ErrorType is a generic type of error, each type can have many error codes
type ErrorType struct {
	Name       string
	codeLookup map[int]*ErrorCode
	Codes      []*ErrorCode
}

// NewErrorType creates a new error type, it must be identified by name.
// For example: var MyError er.ErrorType = NewErrorType("mypackage.MyError")
func NewErrorType(ident string) ErrorType {
	return ErrorType{
		Name:       ident,
		codeLookup: make(map[int]*ErrorCode),
	}
}

func (c *ErrorCode) Is(err R) bool {
	if err == nil {
		return false
	}
	if te, ok := err.(typedErr); ok {
		return te.code == c
	}
	return false
}

func (c *ErrorCode) New(info string, err R) R {
	var messages []string
	if info == "" {
		messages = []string{c.Detail}
	} else {
		messages = []string{c.Detail, info}
	}
	if err == nil {
		err = New("")
	} else if te, ok := err.(typedErr); ok {
		if te.code == c {
			if info != "" {
				te.messages = append(messages, te.messages...)
			}
			return te
		}
	}
	return typedErr{
		messages: messages,
		errType:  c.Type,
		code:     c,
		err:      err,
	}
}

func (e *ErrorType) Is(err R) bool {
	if err == nil {
		return false
	}
	if te, ok := err.(typedErr); ok {
		return te.errType == e
	}
	return false
}

func (e *ErrorType) Decode(err R) *ErrorCode {
	if err == nil {
		return nil
	}
	if te, ok := err.(typedErr); ok {
		return te.code
	}
	return nil
}

// NumberedCode constructs a new error code with a number.
func (e *ErrorType) newErrorCode(
	number int,
	hasNumber bool,
	info string,
	detail string,
) *ErrorCode {
	var result *ErrorCode
	var header string
	if hasNumber {
		header = fmt.Sprintf("%s(%d)", info, number)
	} else {
		header = fmt.Sprintf("%s", info)
	}
	if detail != "" {
		header = header + ": " + detail
	}
	result = &ErrorCode{
		Detail: header,
		Type:   e,
		Number: number,
	}
	if hasNumber {
		e.codeLookup[number] = result
	}
	e.Codes = append(e.Codes, result)
	return result
}

func (c *ErrorCode) Default() R {
	return c.New("", nil)
}

func (e *ErrorType) Code(info string) *ErrorCode {
	return e.newErrorCode(0, false, info, "")
}

func (e *ErrorType) CodeWithDetail(info string, detail string) *ErrorCode {
	return e.newErrorCode(0, false, info, detail)
}

func (e *ErrorType) CodeWithNumber(info string, number int) *ErrorCode {
	return e.newErrorCode(number, true, info, "")
}

func (e *ErrorType) CodeWithNumberAndDetail(info string, number int, detail string) *ErrorCode {
	return e.newErrorCode(number, true, info, detail)
}

func (e *ErrorType) NumberToCode(number int) *ErrorCode {
	return e.codeLookup[number]
}

func (te typedErr) AddMessage(m string) {
	te.messages = append([]string{m}, te.messages...)
}

func (te typedErr) Message() string {
	tem := te.err.Message()
	if tem == "" {
		return strings.Join(te.messages, ": ")
	}
	return fmt.Sprintf("%s: %s", strings.Join(te.messages, ": "), te.err.Message())
}

func (te typedErr) HasStack() bool {
	return te.err.HasStack()
}

func (te typedErr) Stack() []string {
	return te.err.Stack()
}

func (te typedErr) String() string {
	if te.err.HasStack() {
		return fmt.Sprintf("%s\n%s", te.Message(), strings.Join(te.err.Stack(), "\n"))
	}
	return te.Message()
}

func (te typedErr) Wrapped0() error {
	return te.err.Wrapped0()
}

func (te typedErr) Native() error {
	return errors.New(te.String())
}

//////
/// er.R
//////

type R interface {
	Message() string
	Stack() []string
	HasStack() bool
	String() string
	Wrapped0() error
	Native() error
	AddMessage(m string)
}

type err struct {
	messages []string
	e        error
	bstack   []byte
	stack    []string
}

type errAsNative struct {
	e err
}

func (e errAsNative) Error() string {
	return e.e.String()
}

func (e err) HasStack() bool {
	return e.bstack != nil
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

func (e err) AddMessage(m string) {
	if e.messages == nil {
		e.messages = []string{m, e.e.Error()}
	} else {
		e.messages = append([]string{m}, e.messages...)
	}
}

func (e err) Message() string {
	if e.messages == nil {
		return e.e.Error()
	}
	return strings.Join(e.messages, ", ")
}

func (e err) String() string {
	if e.bstack != nil {
		return fmt.Sprintf("%s\n%s", e.Message(), strings.Join(e.Stack(), "\n"))
	}
	return e.Message()
}

func (e err) Wrapped0() error {
	return e.e
}

func (e err) Native() error {
	return errAsNative{e: e}
}

//////

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

func Native(err R) error {
	if err == nil {
		return nil
	}
	return err.Native()
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
	if en, ok := e.(errAsNative); ok {
		return en.e
	}
	return err{
		e:      e,
		bstack: captureStack(),
	}
}

func CauseOf(err R) R {
	if err == nil {
		return err
	}
	if te, ok := err.(typedErr); ok {
		return te.err
	}
	return err
}

func equals(e, r R, fuzzy bool) bool {
	if e == nil || r == nil {
		return e == nil && r == nil
	}
	if te, ok := e.(typedErr); ok {
		if tr, ok := r.(typedErr); ok {
			return te.code == tr.code
		}
		// e is a typed error, r isn't
		return false
	}
	if ee, ok := e.(err); ok {
		if rr, ok := r.(err); ok {
			if ee.e != nil && rr.e != nil {
				// e and r are made by er.E(), check they wrap the same underlying error type
				if ee.e == rr.e {
					return true
				}
				if fuzzy {
					return reflect.TypeOf(ee.e) == reflect.TypeOf(rr.e)
				}
			}
			// doesn't wrap anything, therefore every er.R is unique
			// TODO(cjd): Maybe we want to grab the pc of the caller to er.E() and
			// store that so we can consider the same 2 errors which were made in an identical
			// location.
			return false
		}
		// differing types
		return false
	}
	panic("I don't know what error type this is: " + reflect.TypeOf(e).Name())
}

func Equals(e, r R) bool {
	return equals(e, r, false)
}
func FuzzyEquals(e, r R) bool {
	return equals(e, r, true)
}
