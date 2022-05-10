package consensus

// Algorithm represents set of methods required for
// mining
type Algorithm interface {
	Run()
	Stop()
	Done() <-chan struct{}
}

type Miner struct {
	algo Algorithm
	stop chan struct{}
}

func NewMiner(algo Algorithm) *Miner {
	miner := &Miner{
		algo: algo,
		stop: make(chan struct{}),
	}

	return miner
}

func (m *Miner) Mine() {
	go m.algo.Run()

	select {
	case <-m.algo.Done():
		return
	}
}

func (m *Miner) Abort() {
	m.algo.Stop()
}
