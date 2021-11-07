package discover

// IsTemporaryErr checks is network error is temporary
func IsTemporaryErr(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	e, ok := err.(temporary)

	return ok && e.Temporary()
}
