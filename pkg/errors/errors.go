package errors

import (
	"errors"
	"fmt"
)

type ErrNo int

const (
	NoErr ErrNo = iota
	EForeign
	EInternal
	ECtxClosed
	ELogin
	EConnect
	EListMailbox
	EListMail
	ESelect
)

var _no2str = map[ErrNo]string{
	NoErr:        "no error",
	EForeign:     "foreign error",
	EInternal:    "",
	ECtxClosed:   "context closed",
	ELogin:       "login failed, check your usename and password",
	EConnect:     "connect to server failed",
	EListMailbox: "list mailboxes failed",
	EListMail:    "list mail failed",
	ESelect:      "select mailbox failed",
}

type Error struct {
	ErrNo ErrNo
	Err   error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", _no2str[e.ErrNo], e.Err.Error())
	}
	return _no2str[e.ErrNo]
}

func new(no ErrNo, internal error) *Error {
	return &Error{
		ErrNo: no,
		Err:   internal,
	}
}

func New(s string) *Error {
	return &Error{
		ErrNo: EInternal,
		Err:   errors.New(s),
	}
}

func No(err error) ErrNo {
	if e, ok := err.(*Error); ok {
		return e.ErrNo
	}
	return EForeign
}

func Foreign(err error) *Error {
	return &Error{
		ErrNo: EForeign,
		Err:   err,
	}
}

func CtxClosed() *Error {
	return new(ECtxClosed, nil)
}

func Connect(internal error) *Error {
	return new(EConnect, internal)
}

func ListMailbox(internal error) *Error {
	return new(EListMailbox, internal)
}

func ListMail(internal error) *Error {
	return new(EListMail, internal)
}

func Select(internal error) *Error {
	return new(ESelect, internal)
}
