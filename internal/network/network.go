package network

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/internal/rpc/server"
)

type NetworkService interface {
	AddPeer(address, url string) error
	RemovePeer(address string) error
	PushBlock(block string)
}

type Network struct {
	server *server.Server
	logger *zap.SugaredLogger

	peers map[string]*Peer
}

var _ NetworkService = (*Network)(nil)

func New(lc fx.Lifecycle, server *server.Server, logger *zap.SugaredLogger) NetworkService {
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
		OnStop: func(c context.Context) error {
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
		err := fmt.Errorf("network.AddPeer: %w", err)
		return err
	}

	n.peers[address] = peer

	return nil
}

func (n *Network) RemovePeer(address string) error {
	delete(n.peers, address)
	return nil
}

func (n *Network) PushBlock(block string) {}
