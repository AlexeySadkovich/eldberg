package client

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/AlexeySadkovich/eldberg/internal/node"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/pb"
)

type NodeClient struct {
	client pb.NodeServiceClient
}

func NewClient(addr string) (node.NodeService, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	client := pb.NewNodeServiceClient(conn)

	return &NodeClient{client: client}, nil
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
