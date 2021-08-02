package vite

import "errors"

// Client errors
var (
	ErrCallParametersInvalid = errors.New("call parameters invalid")
	ErrCallOutputMarshal     = errors.New("call output marshal")
	ErrCallMethodInvalid     = errors.New("call method invalid")
)
