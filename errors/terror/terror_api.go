package terror

import (
	"fmt"
	"github.com/towgo/towgo/errors/tcode"
)

// New creates and returns an error which is formatted from given text.
func New(text string) error {
	return &Error{
		stack: callers(),
		text:  text,
		code:  tcode.CodeNil,
	}
}

// Newf returns an error that formats as the given format and args.
func Newf(format string, args ...interface{}) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  tcode.CodeNil,
	}
}

// NewSkip creates and returns an error which is formatted from given text.
// The parameter `skip` specifies the stack callers skipped amount.
func NewSkip(skip int, text string) error {
	return &Error{
		stack: callers(skip),
		text:  text,
		code:  tcode.CodeNil,
	}
}

// NewSkipf returns an error that formats as the given format and args.
// The parameter `skip` specifies the stack callers skipped amount.
func NewSkipf(skip int, format string, args ...interface{}) error {
	return &Error{
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  tcode.CodeNil,
	}
}

// Wrap wraps error with text. It returns nil if given err is nil.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  text,
		code:  Code(err),
	}
}

// Wrapf returns an error annotating err with a stack trace at the point Wrapf is called, and the format specifier.
// It returns nil if given `err` is nil.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  Code(err),
	}
}

// WrapSkip wraps error with text. It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func WrapSkip(skip int, err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  text,
		code:  Code(err),
	}
}

// WrapSkipf wraps error with text that is formatted with given format and args. It returns nil if given err is nil.
// The parameter `skip` specifies the stack callers skipped amount.
// Note that it does not lose the error code of wrapped error, as it inherits the error code from it.
func WrapSkipf(skip int, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(skip),
		text:  fmt.Sprintf(format, args...),
		code:  Code(err),
	}
}
