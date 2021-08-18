package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"

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

func (n *NodeClient) Ping() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	msg, err := n.client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return false, err
	}

	if msg.Message != "pong" {
		return false, ErrUnknownAnswer
	}

	return true, nil
}

func (n *NodeClient) ConnectPeer(address string, url string) error {
	panic("not implemented")
}

func (n *NodeClient) DisconnectPeer(address string) {
	panic("not implemented")
}

func (n *NodeClient) AcceptTransaction(data string) error {
	panic("not implemented")
}

func (n *NodeClient) AcceptBlock(data string) error {
	panic("not implemented")
}
