package client

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/internal/rpc"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/pb"
)

type NodeClient struct {
	client pb.NodeServiceClient
}

var ErrUnknownAnswer = errors.New("unknown answer")

func NewClient(addr string) (rpc.Service, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	client := pb.NewNodeServiceClient(conn)

	return &NodeClient{client: client}, nil
}

func (n *NodeClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	msg, err := n.client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return err
	}

	if msg.Message != "pong" {
		return ErrUnknownAnswer
	}

	return nil
}

func (n *NodeClient) ConnectPeer(address, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	req := &pb.PeerRequest{
		Address: address,
		Url:     url,
	}
	msg, err := n.client.ConnectPeer(ctx, req)
	if err != nil {
		return err
	}

	if msg.Status != "ok" {
		return fmt.Errorf("%s", msg.Detail)
	}

	return nil
}

func (n *NodeClient) DisconnectPeer(address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	req := &pb.PeerRequest{
		Address: address,
	}
	msg, err := n.client.DisconnectPeer(ctx, req)
	if err != nil {
		return err
	}

	if msg.Status != "ok" {
		return fmt.Errorf("%s", msg.Detail)
	}

	return nil
}

func (n *NodeClient) AcceptTransaction(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	req := &pb.TxRequest{
		Data: data,
	}
	msg, err := n.client.AcceptTransaction(ctx, req)
	if err != nil {
		return err
	}

	if msg.Status != "ok" {
		return fmt.Errorf("%s", msg.Detail)
	}

	return nil
}

func (n *NodeClient) AcceptBlock(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cancel()

	req := &pb.BlockRequest{
		Data: data,
	}
	msg, err := n.client.AcceptBlock(ctx, req)
	if err != nil {
		return err
	}

	if msg.Status != "ok" {
		return fmt.Errorf("%s", msg.Detail)
	}

	return nil
}
