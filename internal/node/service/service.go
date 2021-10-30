package service

import (
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

var _ node.Service = (*Node)(nil)

type NodeParams struct {
	fx.In

	Holder  *holder.Holder
	Network *network.Network
	Chain   *blockchain.Chain

	Logger *zap.SugaredLogger
}

func New(p NodeParams) node.Service {
	return &Node{
		holder:  p.Holder,
		network: p.Network,
		chain:   p.Chain,
		logger:  p.Logger,
	}
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

func (n *Node) AcceptTransaction(data []byte) error {
	tx := blockchain.EmptyTransaction()

	if err := tx.Deserialize(data); err != nil {
		err = fmt.Errorf("node.AcceptTransaction: failed to deserialize transaction: %w", err)
		n.logger.Debug(err)
		return err
	}

	if !tx.IsValid() {
		err := fmt.Errorf("node.AcceptTransaction: %w", node.ErrInvalidTx)
		n.logger.Debug(err)
		return err
	}

	lastHash, err := n.chain.GetLastHash()
	if err != nil {
		err = fmt.Errorf("node.AcceptTransaction: transaction not accepted: %w", err)
		n.logger.Debug(err)
		return err
	}

	if n.chain.GetCurrentBlock() == nil {
		block := blockchain.NewBlock(n.holder.Address(), lastHash)
		n.chain.SetCurrentBlock(block)
	}

	if err := n.chain.GetCurrentBlock().AddTransaction(tx); err != nil {
		err = fmt.Errorf("node.AddTransaction: %w", err)
		n.logger.Debug(err)
		return err
	}

	if n.chain.GetCurrentBlock().Fullness() == config.TXSLIMIT {
		block, err := n.chain.GetCurrentBlock().Serialize()
		if err != nil {
			err = fmt.Errorf("node.AddTransaction: failed to serialize block %w", err)
			n.logger.Debug(err)
			return err
		}

		if err := n.network.PushBlock(block); err != nil {
			err = fmt.Errorf("node.AddTransaction: failed to push block %w", err)
			n.logger.Debug(err)
			return err
		}
	}

	return nil
}

func (n *Node) AcceptBlock(data []byte) error {
	block := blockchain.EmptyBlock()

	if err := block.Deserialize(data); err != nil {
		err = fmt.Errorf("node.AcceptBlock: failed to deserialize block: %w", err)
		n.logger.Debug(err)
		return err
	}

	if !block.IsValid() {
		err := fmt.Errorf("node.AcceptBlock: %w", node.ErrInvalidBlock)
		n.logger.Debug(err)
		return err
	}

	return n.chain.AddBlock(block)
}
