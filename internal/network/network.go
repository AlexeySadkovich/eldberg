package network

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/internal/rpc/server"
)

type NetworkService interface {
	AddPeer(address, url string) error
	RemovePeer(address string)
	PushBlock(block string)
}

type Network struct {
	server *server.Server
	logger *zap.SugaredLogger

	peers map[string]*Peer
}

var _ NetworkService = (*Network)(nil)

func New(server *server.Server, logger *zap.SugaredLogger) NetworkService {
	return &Network{
		server: server,
		logger: logger,
		peers:  make(map[string]*Peer),
	}
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
		err := fmt.Errorf("network.AddPeer: new peer: %w", err)
		n.logger.Debug(err)
		return err
	}

	n.peers[address] = peer

	return nil
}

func (n *Network) RemovePeer(address string) {
	delete(n.peers, address)
}

func (n *Network) PushBlock(block string) {}
