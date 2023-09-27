package services

import "fmt"


type Error struct {
    orig error
    msg string
    code ErrorCode
}

type ErrorCode uint

const (
    ErrorCodeUnknown ErrorCode = iota
    ErrorCodeNotFound
    ErrorCodeCellTaken
    ErrorCodeIllegalMove
    ErrorCodeInvalidArgument
)

func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error{
    return &Error{
        code: code,
        orig: orig,
        msg: fmt.Sprintf(format, a...),
    }
}

func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
    return WrapErrorf(nil, code, format, a...)
}

func (e *Error) Error() string {
    if e.orig != nil {
        return fmt.Sprintf("%s, %v", e.msg, e.orig)
    }
    return e.msg
}


func (e *Error) Unwrap() error {
    return e.orig
}

func (e *Error) Code() ErrorCode {
    return e.code
}
