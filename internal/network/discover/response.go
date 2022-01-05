package discover

import "time"

type awaitedResponse interface {
	received() chan struct{}
	timeout() chan struct{}
}

type response struct {
	typ messageType
	rec chan struct{} // response received
	exp chan struct{} // timeout expired
}

func makeResponse(msgType messageType) *response {
	return &response{
		typ: msgType,
		rec: make(chan struct{}),
		exp: make(chan struct{}),
	}
}

func (r *response) await() {
	timer := time.NewTimer(pingMaxTime)
	defer timer.Stop()

	select {
	case <-r.rec:
		return
	case <-timer.C:
		close(r.exp)
		return
	}
}

func (r *response) markReceived() {
	close(r.rec)
}

func (r *response) received() chan struct{} {
	return r.rec
}

func (r *response) timeout() chan struct{} {
	return r.exp
}
