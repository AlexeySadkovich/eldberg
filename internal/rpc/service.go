package rpc

import "github.com/AlexeySadkovich/eldberg/internal/node"

type Service interface {
	node.NodeService
	Ping() error
}
