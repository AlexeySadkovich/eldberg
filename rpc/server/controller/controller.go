package controller

import (
	"context"
	"github.com/AlexeySadkovich/eldberg/node"
	pb2 "github.com/AlexeySadkovich/eldberg/rpc/pb"
)

type NodeServiceController struct {
	pb2.UnimplementedNodeServiceServer
	nodeService node.Service
}

func New(node node.Service) *NodeServiceController {
	return &NodeServiceController{
		nodeService: node,
	}
}

func (s *NodeServiceController) Ping(ctx context.Context, request *pb2.PingRequest) (*pb2.PingReply, error) {
	message := &pb2.PingReply{Message: "pong"}
	return message, nil
}

func (s *NodeServiceController) ConnectPeer(ctx context.Context, request *pb2.PeerRequest) (*pb2.PeerReply, error) {
	if err := s.nodeService.ConnectPeer(request.Address, request.Url); err != nil {
		return nil, err
	}

	message := &pb2.PeerReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) DisconnectPeer(ctx context.Context, request *pb2.PeerRequest) (*pb2.PeerReply, error) {
	s.nodeService.DisconnectPeer(request.Address)

	message := &pb2.PeerReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) AcceptTransaction(ctx context.Context, request *pb2.TxRequest) (*pb2.TxReply, error) {
	if err := s.nodeService.AcceptTransaction(request.Data); err != nil {
		return nil, err
	}

	message := &pb2.TxReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) AcceptBlock(ctx context.Context, request *pb2.BlockRequest) (*pb2.BlockReply, error) {
	if err := s.nodeService.AcceptBlock(request.Data); err != nil {
		return nil, err
	}

	message := &pb2.BlockReply{
		Status: "ok",
	}

	return message, nil
}
