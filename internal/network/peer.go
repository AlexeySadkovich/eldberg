package network

import (
	"fmt"

	"github.com/AlexeySadkovich/eldberg/internal/rpc"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/client"
)

type Peer struct {
	address string
	url     string
	client  rpc.Service
}

func NewPeer(address, url string) (*Peer, error) {
	cli, err := client.NewClient(address)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	p := &Peer{
		address: address,
		url:     url,
		client:  cli,
	}

	if ok := p.Ping(); !ok {
		return nil, ErrPeerUnavailable
	}

	return p, nil
}

func (p *Peer) Ping() bool {
	if err := p.client.Ping(); err != nil {
		return false
	}

	return true
}

func (p *Peer) OfferBlock(block []byte) error {
	if ok := p.Ping(); !ok {
		return ErrPeerUnavailable
	}

	if err := p.client.AcceptBlock(block); err != nil {
		return err
	}

	return nil
}
