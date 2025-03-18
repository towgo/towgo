// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package terror

import (
	"github.com/towgo/towgo/errors/tcode"
)

// Code returns the error code.
// It returns CodeNil if it has no error code.
func (err *Error) Code() tcode.Code {
	if err == nil {
		return tcode.CodeNil
	}
	if err.code == tcode.CodeNil {
		return Code(err.Unwrap())
	}
	return err.code
}

// SetCode updates the internal code with given code.
func (err *Error) SetCode(code tcode.Code) {
	if err == nil {
		return
	}
	err.code = code
}
