package network

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexeySadkovich/eldberg/node"
	"github.com/AlexeySadkovich/eldberg/rpc/server"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Network struct {
	server *server.Server
	logger *zap.SugaredLogger

	peers map[string]*Peer
}

func New(lc fx.Lifecycle, server *server.Server, logger *zap.SugaredLogger) *Network {
	netw := &Network{
		server: server,
		logger: logger,
		peers:  make(map[string]*Peer),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting network...")
			go netw.Run(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			netw.Stop()
			return nil
		},
	})

	return netw
}

func (n *Network) Run(ctx context.Context) {
	n.server.Start(ctx)
}

func (n *Network) Stop() {
	n.server.Stop()
}

func (n *Network) AddPeer(address, url string) error {
	peer, err := NewPeer(address, url)
	if err != nil {
		return fmt.Errorf("network.AddPeer: %w", err)
	}

	n.peers[address] = peer

	return nil
}

func (n *Network) RemovePeer(address string) error {
	delete(n.peers, address)
	return nil
}

func (n *Network) PushBlock(block []byte) error {
	push := new(pusher)
	wg := sync.WaitGroup{}

	for _, peer := range n.peers {
		wg.Add(1)
		go n.offerBlockToPeer(block, peer, push, &wg)
	}

	wg.Wait()

	if push.invalid() {
		return ErrBlockNotAccepted
	}

	return nil
}

func (n *Network) offerBlockToPeer(block []byte, peer *Peer, push *pusher, wg *sync.WaitGroup) {
	defer wg.Done()

	push.incTotal()
	if err := peer.OfferBlock(block); err != nil {
		if errors.Is(err, ErrPeerUnavailable) {
			return
		}

		if errors.Is(err, node.ErrInvalidBlock) {
			push.incInvalid()
		}
	}
}

// pusher controls process of pushing
// block to network
type pusher struct {
	sync.Mutex
	total         int
	invalidations int
}

func (p *pusher) incTotal() {
	p.Lock()
	p.total++
	p.Unlock()
}

func (p *pusher) incInvalid() {
	p.Lock()
	p.invalidations++
	p.Unlock()
}

func (p *pusher) invalid() bool {
	diff := float64(p.total - p.invalidations)
	delta := (diff / float64(p.invalidations)) * 100

	// delta is percentage of peers which didn't
	// accept block.
	// TODO: update invalidations check
	if delta > 30 {
		return true
	}

	return false
}
