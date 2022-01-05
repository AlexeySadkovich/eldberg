package discover

import (
	"fmt"
	"math/rand"
	"net"
)

type (
	// netnode stores information about
	// network node
	netnode struct {
		id   nodeID
		ip   net.IP
		port int
	}

	nodeID [32]byte
)

func (n *netnode) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.ip, Port: n.port}
}

func (id nodeID) Bytes() []byte {
	return id[:]
}

func (id nodeID) Hex() string {
	return fmt.Sprintf("%x", id)
}

func newID() (nodeID, error) {
	var id nodeID
	_, err := rand.Read(id.Bytes())
	return id, err
}
