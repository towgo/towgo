package utils

import (
	"github.com/towgo/towgo/errors/tcode"
	"github.com/towgo/towgo/errors/terror"
)

// Throw throws out an exception, which can be caught be TryCatch or recover.
func Throw(exception interface{}) {
	panic(exception)
}

// Try implements try... logistics using internal panic...recover.
// It returns error if any exception occurs, or else it returns nil.
func Try(try func()) (err error) {
	if try == nil {
		return
	}
	defer func() {
		if exception := recover(); exception != nil {
			if v, ok := exception.(error); ok && terror.HasStack(v) {
				err = v
			} else {
				err = terror.NewCodef(tcode.CodeInternalPanic, "%+v", exception)
			}
		}
	}()
	try()
	return
}

// TryCatch implements `try...catch..`. logistics using internal `panic...recover`.
// It automatically calls function `catch` if any exception occurs and passes the exception as an error.
// If `catch` is given nil, it ignores the panic from `try` and no panic will throw to parent goroutine.
//
// But, note that, if function `catch` also throws panic, the current goroutine will panic.
func TryCatch(try func(), catch func(exception error)) {
	if try == nil {
		return
	}
	if exception := Try(try); exception != nil && catch != nil {
		catch(exception)
	}
}
