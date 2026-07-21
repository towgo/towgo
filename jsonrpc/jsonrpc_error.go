package jsonrpc

// Errors describes nested JSON-RPC error values.
type Errors interface {
	// Set assigns an error code and message.
	Set(int64, string)
	// NewChild creates and attaches a child error.
	NewChild(int64, string) *Error
	// Gegcode returns the error code.
	Gegcode() int64
	// Error returns the error message.
	Error() string
	// AppendChild attaches a child error value.
	AppendChild(Errors)
}

// Error is the JSON-RPC error payload.
type Error struct {
	// Message is the human-readable error text.
	Message string `json:"message"`
	// Code is the JSON-RPC or application error code.
	Code int64 `json:"code"`
	// Data carries nested or structured error details.
	Data any `json:"data"`
}

// New creates an empty Error.
func New() *Error {
	return &Error{}
}

// Set assigns an error code and message.
func (e *Error) Set(code int64, message string) {
	e.Code = code
	e.Message = message
	if e.Message == "" {
		e.Message = CODETYPES[code]
		if e.Message == "" {
			e.Message = "unknown error"
		}
	}
}

// NewChild creates and attaches a child error.
func (e *Error) NewChild(code int64, message string) *Error {
	child := &Error{
		Code:    code,
		Message: message,
	}
	if child.Message == "" {
		child.Message = CODETYPES[code]
		if child.Message == "" {
			child.Message = "unknown error"
		}
	}
	e.Data = child
	return child
}

// AppendChild attaches a nested error value.
func (e *Error) AppendChild(es Errors) {
	e.Data = es
}

// Gegcode returns the error code.
func (e *Error) Gegcode() int64 {
	return e.Code
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

// codeTypes maps status codes to default error messages.
type codeTypes map[int64]string

// CODETYPES maps JSON-RPC and application status codes to default messages.
var CODETYPES = NewCodeTypes()

// NewCodeTypes creates the default JSON-RPC code message map.
func NewCodeTypes() codeTypes {
	return codeTypes{
		0:      "unknown error",
		200:    "ok",
		-32700: "Parse error",
		-32600: "Invalid Request",
		-32601: "Method not found",
		-32602: "Invalid params",
		-32603: "Internal error",
		-32000: "Server error",
		408:    "REQUEST_TIMEOUT",
		500:    "internal server error",
	}
}
