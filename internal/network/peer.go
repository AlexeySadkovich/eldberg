package network

import (
	"errors"
	"fmt"

	"github.com/AlexeySadkovich/eldberg/internal/node"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/client"
)

type Peer struct {
	address string
	url     string
	client  node.NodeService
}

var (
	ErrPeerAnavailbale = errors.New("peer anavailable")
)

func NewPeer(address, url string) (*Peer, error) {
	client, err := client.NewClient(address)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Peer{
		address: address,
		url:     url,
		client:  client,
	}, nil
}
