package service

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/network"
	"github.com/AlexeySadkovich/eldberg/internal/node"
)

type Node struct {
	holder  *holder.Holder
	network *network.Network
	chain   *blockchain.Chain

	logger *zap.SugaredLogger
}

var _ node.NodeService = (*Node)(nil)

type NodeParams struct {
	fx.In

	Holder  *holder.Holder
	Network *network.Network
	Chain   *blockchain.Chain

	Logger *zap.SugaredLogger
}

func New(lc fx.Lifecycle, p NodeParams) node.NodeService {
	node := &Node{
		holder:  p.Holder,
		network: p.Network,
		chain:   p.Chain,
		logger:  p.Logger,
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
	if err := n.network.AddPeer(address, url); err != nil {
		err = fmt.Errorf("node.ConnectPeer: %w", err)
		n.logger.Debug(err)
		return err
	}

	return nil
}

func (n *Node) DisconnectPeer(address string) error {
	return n.network.RemovePeer(address)
}

func (n *Node) AcceptTransaction(data string) error {
	tx := new(blockchain.Transaction)

	if err := tx.Deserialize(data); err != nil {
		err = fmt.Errorf("node.AcceptTransaction: failed to deserialize transaction: %w", err)
		n.logger.Debug(err)
		return err
	}

	if !tx.IsValid() {
		err := fmt.Errorf("node.AcceptTransaction: %w", ErrInvalidTx)
		n.logger.Debug(err)
		return err
	}

	lastHash, err := n.chain.GetLastHash()
	if err != nil {
		err = fmt.Errorf("node.AcceptTransaction: transaction not accepted: %w", err)
		n.logger.Debug(err)
		return err
	}

	if n.chain.CurrentBlock == nil {
		n.chain.CurrentBlock = blockchain.NewBlock(n.holder.Address(), lastHash)
	}

	if err := n.chain.CurrentBlock.AddTransaction(tx); err != nil {
		err = fmt.Errorf("node.AddTransaction: %w", err)
		n.logger.Debug(err)
		return err
	}

	if n.chain.CurrentBlock.Fullness() == config.TXSLIMIT {
		block := n.chain.CurrentBlock.Serialize()
		n.network.PushBlock(block)
	}

	return nil
}

func (n *Node) AcceptBlock(data string) error {
	block := new(blockchain.Block)

	if err := block.Deserialize(data); err != nil {
		err = fmt.Errorf("node.AcceptBlock: failed to deserialize block: %w", err)
		n.logger.Debug(err)
		return err
	}

	if !block.IsValid() {
		err := fmt.Errorf("node.AcceptBlock: %w", ErrInvalidBlock)
		n.logger.Debug(err)
		return err
	}

	return n.chain.AddBlock(block)
}
