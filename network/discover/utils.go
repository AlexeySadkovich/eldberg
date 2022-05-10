package discover

// IsTemporaryErr checks is network error is temporary
func IsTemporaryErr(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	e, ok := err.(temporary)

	return ok && e.Temporary()
}

func HasBit(b byte, pos uint) bool {
	pos = 7 - pos
	val := b & (1 << pos)

	return val > 0
}
