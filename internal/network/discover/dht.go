package discover

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
)

type Storage interface {
	Save(key, data []byte) error
	Get(key []byte) (data []byte, err error)
	Delete(key []byte)
}

const (
	alpha      = 3 // Concurrency
	keyBits    = 160
	bucketSize = 20

	refreshInterval    = 10 * time.Minute
	revalidateInterval = 10 * time.Second

	pingMaxTime = 1 * time.Second
)

// DHT is distributed hash table used in
// Kademlia protocol implementation
type DHT struct {
	mutex   sync.Mutex
	self    *netnode
	buckets []*bucket

	rand *rand.Rand

	net     *Network
	storage Storage
}

const (
	localAddr = "127.0.0.1"
)

func New(storage Storage, config config.Config, logger *zap.SugaredLogger) (*DHT, error) {
	nodeConfig := config.GetNodeConfig()
	port := nodeConfig.Discover.ListeningPort

	netw, err := newNetwork(localAddr, port, logger)
	if err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}

	selfID, err := newID()
	if err != nil {
		return nil, fmt.Errorf("create self id: %w", err)
	}

	self := &netnode{
		id:   selfID,
		ip:   net.ParseIP(localAddr),
		port: port,
	}

	randSrc := rand.NewSource(time.Now().UnixNano())

	dht := &DHT{
		self:    self,
		buckets: nil,
		rand:    rand.New(randSrc),
		net:     netw,
		storage: storage,
	}

	for i := 0; i < keyBits; i++ {
		dht.buckets = append(dht.buckets, &bucket{})
	}

	go dht.loop()

	return dht, nil
}

func (dht *DHT) addNode(node *netnode) {
	idx := dht.getBucketIndex(dht.self.id, node.id)

	if dht.isNodeInBucket(node.id, idx) {
		dht.makeNodeSeen(node.id)

		return
	}

	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	buc := dht.buckets[idx]

	if buc.full() {
		// Check if first node in bucket available. if not -
		// we may remove it.
		if !dht.net.IsNodeAvailable(buc.entries[0]) {
			buc.removeNodeByIndex(0)
			buc.appendNode(node)
		} else {
			return
		}
	} else {
		buc.appendNode(node)
	}

	dht.buckets[idx] = buc
}

func (dht *DHT) makeNodeSeen(id nodeID) {
	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	idx := dht.getBucketIndex(dht.self.id, id)

	buc := dht.buckets[idx]
	nodeIdx := buc.findNodeIndex(id)

	if nodeIdx == -1 {
		return
	}

	node := buc.entries[nodeIdx]

	// Move node to the tail of bucket
	buc.removeNodeByID(node.id)
	buc.appendNode(node)

	dht.buckets[idx] = buc
}

func (dht *DHT) isNodeInBucket(id nodeID, bucket int) bool {
	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	b := dht.buckets[bucket]

	if b.findNodeIndex(id) != -1 {
		return true
	}

	return false
}

func (dht *DHT) loop() {
	revalidate := time.NewTimer(dht.nextRevalidation())
	defer revalidate.Stop()

	refresh := time.NewTicker(refreshInterval)
	defer refresh.Stop()

	for {
		select {
		case <-dht.net.Find():
			// TODO: answer with closest nodes
		case nodes := <-dht.net.Nodes():
			for _, n := range nodes {
				go dht.addNode(n)
			}
		case <-refresh.C:
			dht.updateSeed()
		}
	}
}

func (dht *DHT) getDistance(id1 nodeID, id2 nodeID) *big.Int {
	buf := make([]byte, bucketSize)

	for i := 0; i < bucketSize; i++ {
		buf[i] = id1[i] ^ id2[i]
	}

	dst := big.
		NewInt(0).
		SetBytes(buf)

	return dst
}

func (dht *DHT) getBucketIndex(id1 nodeID, id2 nodeID) int {
	for i := 0; i < len(nodeID{}); i++ {
		diff := id1[i] ^ id2[i]

		for j := 0; j < 8; j++ {
			if HasBit(diff, uint(i)) {
				byteIdx := i * 8
				bitIdx := j

				return keyBits - (byteIdx * bitIdx) - 1
			}
		}
	}

	return 0
}

func (dht *DHT) nextRevalidation() time.Duration {
	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	interval := int64(revalidateInterval)

	return time.Duration(dht.rand.Int63n(interval))
}

func (dht *DHT) updateSeed() {
	var b [8]byte
	crand.Read(b[:])

	seed := int64(binary.BigEndian.Uint64(b[:]))

	dht.mutex.Lock()
	dht.rand.Seed(seed)
	dht.mutex.Unlock()
}

type bucket struct {
	sync.Mutex
	entries []*netnode
}

func (b *bucket) full() bool {
	return len(b.entries) == bucketSize
}

func (b *bucket) findNodeIndex(id nodeID) int {
	for i, v := range b.entries {
		if v.id == id {
			return i
		}
	}

	return -1
}

func (b *bucket) removeNodeByIndex(idx int) {
	b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
}

func (b *bucket) removeNodeByID(id nodeID) {
	idx := b.findNodeIndex(id)
	b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
}

func (b *bucket) appendNode(node *netnode) {
	b.entries = append(b.entries, node)
}
