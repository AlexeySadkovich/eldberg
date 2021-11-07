package discover

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	alpha      = 3 // Concurrency
	keyBits    = 160
	bucketSize = 20

	refreshInterval    = 10 * time.Minute
	revalidateInterval = 10 * time.Second
)

type Table struct {
	mutex sync.Mutex
	self  *netnode
	nodes [][]*netnode
}

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

func newTable(ip string, port int) (*Table, error) {
	rand.Seed(time.Now().UnixNano())

	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("create ID: %w", err)
	}

	self := &netnode{
		id:   id,
		ip:   net.ParseIP(ip),
		port: port,
	}

	t := &Table{
		self: self,
	}

	for i := 0; i < keyBits; i++ {
		t.nodes = append(t.nodes, []*netnode{})
	}

	return t, nil
}

func newID() (nodeID, error) {
	var id nodeID
	_, err := rand.Read(id.Bytes())
	return id, err
}
