package network

import (
	"context"

	"github.com/AlexeySadkovich/eldberg/internal/rpc/server"
)

type NetworkService interface {
	AddPeer(address, url string) error
	RemovePeer(address string)
	PushBlock(block string)
}

type Network struct {
	server *server.Server
}

var _ NetworkService = (*Network)(nil)

func New(server *server.Server) NetworkService {
	return &Network{}
}

func (n *Network) Run(ctx context.Context) {
	n.server.Start(ctx)
}

func (n *Network) Stop() {
	n.server.Stop()
}

func (n *Network) AddPeer(address, url string) error {
	return nil
}

func (n *Network) RemovePeer(address string) {}

func (n *Network) PushBlock(block string) {}
