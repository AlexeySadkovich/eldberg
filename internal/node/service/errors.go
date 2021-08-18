package service

import "errors"

var (
	ErrInvalidTx    = errors.New("invalid transaction")
	ErrInvalidBlock = errors.New("invalid block")
)
