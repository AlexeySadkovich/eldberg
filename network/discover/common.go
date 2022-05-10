package discover

import (
	"fmt"
	"math/big"
	"math/rand"
	"net"
)

type (
	// netnode stores information about
	// network node
	netnode struct {
		ID   nodeID
		IP   net.IP
		Port int
	}

	nodeID [32]byte
)

var nilNodeID nodeID

func (n *netnode) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP, Port: n.Port}
}

func (id nodeID) Bytes() []byte {
	return id[:]
}

func (id nodeID) Hex() string {
	return fmt.Sprintf("%x", id)
}

// nodesByDistance is a list of nodes
// ordered by distance to target.
type nodesByDistance struct {
	entries []*netnode
	target  nodeID
}

func (n *nodesByDistance) appendUnique(nodes ...*netnode) {
	for _, v := range nodes {
		exists := false
		for _, vv := range n.entries {
			if v.ID == vv.ID {
				exists = true
				break
			}
		}
		if !exists {
			n.entries = append(n.entries, v)
		}
	}
}

func (n *nodesByDistance) remove(id nodeID) {
	for i, v := range n.entries {
		if v.ID == id {
			n.entries = append(n.entries[:i], n.entries[i+1:]...)
			return
		}
	}
}

func (n *nodesByDistance) Len() int {
	return len(n.entries)
}

func (n *nodesByDistance) Swap(i, j int) {
	n.entries[i], n.entries[j] = n.entries[j], n.entries[i]
}

func (n *nodesByDistance) Less(i, j int) bool {
	// Closure for calculating distance between nodes.
	distance := func(id1 nodeID, id2 nodeID) *big.Int {
		buf1 := new(big.Int).SetBytes(id1.Bytes())
		buf2 := new(big.Int).SetBytes(id2.Bytes())
		res := new(big.Int).Xor(buf1, buf2)
		return res
	}

	iDist := distance(n.entries[i].ID, n.target)
	jDist := distance(n.entries[j].ID, n.target)

	return iDist.Cmp(jDist) == -1
}

func newID() nodeID {
	var id nodeID
	rand.Read(id[:])
	return id
}
