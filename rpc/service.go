package rpc

import (
	"github.com/AlexeySadkovich/eldberg/node"
)

type Service interface {
	node.Service
	Ping() error
}
