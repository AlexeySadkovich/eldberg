package service

import (
	"context"

	"go.uber.org/fx"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/network"
	"github.com/AlexeySadkovich/eldberg/internal/node"
)

type Node struct {
	holder  *holder.Holder
	network *network.Network
	chain   *blockchain.Chain
}

var _ node.NodeService = (*Node)(nil)

type NodeParams struct {
	fx.In

	Holder  *holder.Holder
	Network *network.Network
	Chain   *blockchain.Chain
}

func New(lc fx.Lifecycle, p NodeParams) node.NodeService {
	node := &Node{
		holder:  p.Holder,
		network: p.Network,
		chain:   p.Chain,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go node.network.Run(ctx)
			return nil
		},
		OnStop: func(c context.Context) error {
			node.network.Stop()
			return nil
		},
	})

	return node
}

func (n *Node) ConnectPeer(address string, url string) error {
	panic("not implemented")
}

func (n *Node) DisconnectPeer(address string) {
	panic("not implemented")
}

func (n *Node) AcceptTransaction(data string) error {
	panic("not implemented")
}

func (n *Node) AcceptBlock(data string) error {
	panic("not implemented")
}
