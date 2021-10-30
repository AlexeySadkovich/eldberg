package controller

import (
	"context"

	"github.com/AlexeySadkovich/eldberg/internal/node"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/pb"
)

type NodeServiceController struct {
	pb.UnimplementedNodeServiceServer
	nodeService node.Service
}

func New(node node.Service) *NodeServiceController {
	return &NodeServiceController{
		nodeService: node,
	}
}

func (s *NodeServiceController) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingReply, error) {
	message := &pb.PingReply{Message: "pong"}
	return message, nil
}

func (s *NodeServiceController) ConnectPeer(ctx context.Context, request *pb.PeerRequest) (*pb.PeerReply, error) {
	if err := s.nodeService.ConnectPeer(request.Address, request.Url); err != nil {
		return nil, err
	}

	message := &pb.PeerReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) DisconnectPeer(ctx context.Context, request *pb.PeerRequest) (*pb.PeerReply, error) {
	s.nodeService.DisconnectPeer(request.Address)

	message := &pb.PeerReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) AcceptTransaction(ctx context.Context, request *pb.TxRequest) (*pb.TxReply, error) {
	if err := s.nodeService.AcceptTransaction(request.Data); err != nil {
		return nil, err
	}

	message := &pb.TxReply{
		Status: "ok",
	}

	return message, nil
}

func (s *NodeServiceController) AcceptBlock(ctx context.Context, request *pb.BlockRequest) (*pb.BlockReply, error) {
	if err := s.nodeService.AcceptBlock(request.Data); err != nil {
		return nil, err
	}

	message := &pb.BlockReply{
		Status: "ok",
	}

	return message, nil
}
