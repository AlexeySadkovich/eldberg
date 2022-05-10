// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// NodeServiceClient is the client API for NodeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeServiceClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingReply, error)
	ConnectPeer(ctx context.Context, in *PeerRequest, opts ...grpc.CallOption) (*PeerReply, error)
	DisconnectPeer(ctx context.Context, in *PeerRequest, opts ...grpc.CallOption) (*PeerReply, error)
	AcceptTransaction(ctx context.Context, in *TxRequest, opts ...grpc.CallOption) (*TxReply, error)
	AcceptBlock(ctx context.Context, in *BlockRequest, opts ...grpc.CallOption) (*BlockReply, error)
}

type nodeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeServiceClient(cc grpc.ClientConnInterface) NodeServiceClient {
	return &nodeServiceClient{cc}
}

func (c *nodeServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingReply, error) {
	out := new(PingReply)
	err := c.cc.Invoke(ctx, "/rpc.NodeService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) ConnectPeer(ctx context.Context, in *PeerRequest, opts ...grpc.CallOption) (*PeerReply, error) {
	out := new(PeerReply)
	err := c.cc.Invoke(ctx, "/rpc.NodeService/ConnectPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) DisconnectPeer(ctx context.Context, in *PeerRequest, opts ...grpc.CallOption) (*PeerReply, error) {
	out := new(PeerReply)
	err := c.cc.Invoke(ctx, "/rpc.NodeService/DisconnectPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AcceptTransaction(ctx context.Context, in *TxRequest, opts ...grpc.CallOption) (*TxReply, error) {
	out := new(TxReply)
	err := c.cc.Invoke(ctx, "/rpc.NodeService/AcceptTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AcceptBlock(ctx context.Context, in *BlockRequest, opts ...grpc.CallOption) (*BlockReply, error) {
	out := new(BlockReply)
	err := c.cc.Invoke(ctx, "/rpc.NodeService/AcceptBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeServiceServer is the server API for NodeService service.
// All implementations must embed UnimplementedNodeServiceServer
// for forward compatibility
type NodeServiceServer interface {
	Ping(context.Context, *PingRequest) (*PingReply, error)
	ConnectPeer(context.Context, *PeerRequest) (*PeerReply, error)
	DisconnectPeer(context.Context, *PeerRequest) (*PeerReply, error)
	AcceptTransaction(context.Context, *TxRequest) (*TxReply, error)
	AcceptBlock(context.Context, *BlockRequest) (*BlockReply, error)
	mustEmbedUnimplementedNodeServiceServer()
}

// UnimplementedNodeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeServiceServer struct {
}

func (UnimplementedNodeServiceServer) Ping(context.Context, *PingRequest) (*PingReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedNodeServiceServer) ConnectPeer(context.Context, *PeerRequest) (*PeerReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConnectPeer not implemented")
}
func (UnimplementedNodeServiceServer) DisconnectPeer(context.Context, *PeerRequest) (*PeerReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisconnectPeer not implemented")
}
func (UnimplementedNodeServiceServer) AcceptTransaction(context.Context, *TxRequest) (*TxReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptTransaction not implemented")
}
func (UnimplementedNodeServiceServer) AcceptBlock(context.Context, *BlockRequest) (*BlockReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptBlock not implemented")
}
func (UnimplementedNodeServiceServer) mustEmbedUnimplementedNodeServiceServer() {}

// UnsafeNodeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeServiceServer will
// result in compilation errors.
type UnsafeNodeServiceServer interface {
	mustEmbedUnimplementedNodeServiceServer()
}

func RegisterNodeServiceServer(s grpc.ServiceRegistrar, srv NodeServiceServer) {
	s.RegisterService(&NodeService_ServiceDesc, srv)
}

func _NodeService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.NodeService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_ConnectPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).ConnectPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.NodeService/ConnectPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).ConnectPeer(ctx, req.(*PeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_DisconnectPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).DisconnectPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.NodeService/DisconnectPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).DisconnectPeer(ctx, req.(*PeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AcceptTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AcceptTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.NodeService/AcceptTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AcceptTransaction(ctx, req.(*TxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AcceptBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AcceptBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.NodeService/AcceptBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AcceptBlock(ctx, req.(*BlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeService_ServiceDesc is the grpc.ServiceDesc for NodeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.NodeService",
	HandlerType: (*NodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _NodeService_Ping_Handler,
		},
		{
			MethodName: "ConnectPeer",
			Handler:    _NodeService_ConnectPeer_Handler,
		},
		{
			MethodName: "DisconnectPeer",
			Handler:    _NodeService_DisconnectPeer_Handler,
		},
		{
			MethodName: "AcceptTransaction",
			Handler:    _NodeService_AcceptTransaction_Handler,
		},
		{
			MethodName: "AcceptBlock",
			Handler:    _NodeService_AcceptBlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/rpc/service.proto",
}