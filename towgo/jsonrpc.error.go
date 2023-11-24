package towgo

type Errors interface {
	Set(int64, string)
	NewChild(int64, string) *Error
	GetCode() int64
	Error() string
	AppendChild(Errors)
}

type Error struct {
	Message string      `json:"message"`
	Code    int64       `json:"code"`
	Data    interface{} `json:"data"`
}

func New() *Error {
	return &Error{}
}

func (e *Error) Set(code int64, message string) {
	e.Code = code
	e.Message = message
	if message == "" {
		e.Message = CODETYPES[code]
	}

}

func (e *Error) NewChild(code int64, message string) *Error {

	child := Error{
		Message: message,
		Code:    code,
	}
	if message == "" {
		child.Message = CODETYPES[code]
	}
	e.Data = child
	return &child
}

func (e *Error) AppendChild(es Errors) {

	e.Data = es

}

func (e *Error) GetCode() int64 {
	return e.Code
}

func (e *Error) Error() string {
	return e.Message
}
