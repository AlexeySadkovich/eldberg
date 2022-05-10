package network

import "errors"

var (
	ErrPeerUnavailable  = errors.New("peer unavailable")
	ErrBlockNotAccepted = errors.New("block not accepted (invalid)")
)
