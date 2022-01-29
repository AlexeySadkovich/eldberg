package discover

import "time"

type response struct {
	from nodeID
	typ  messageType
	rec  chan struct{} // response received
}

type awaitedResponse struct {
	wrapped *response
	exp     chan struct{} // timeout expired
}

func makeResponse(from nodeID, msgType messageType) *response {
	return &response{
		from: from,
		typ:  msgType,
		rec:  make(chan struct{}),
	}
}

func (r *response) await(timeout time.Duration) *awaitedResponse {
	awaited := &awaitedResponse{
		wrapped: r,
		exp:     make(chan struct{}),
	}

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case <-awaited.wrapped.rec:
			return
		case <-timer.C:
			close(awaited.exp)
			return
		}
	}()

	return awaited
}

func (r *response) markReceived() {
	close(r.rec)
}

func (r *response) received() chan struct{} {
	return r.rec
}

func (r *awaitedResponse) from() nodeID {
	return r.wrapped.from
}

func (r *awaitedResponse) received() chan struct{} {
	return r.wrapped.rec
}

func (r *awaitedResponse) timeout() chan struct{} {
	return r.exp
}
