package discover

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net"
	"sort"
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

	pingTimeout     = 1 * time.Second
	responseTimeout = 2 * time.Second
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

	closeCh chan struct{}
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
		dht.buckets = append(dht.buckets, &bucket{lastRefresh: time.Now()})
	}

	go dht.timers()
	go dht.loop()

	return dht, nil
}

func (dht *DHT) iterate(target nodeID) {
	closest := dht.getClosestNodes(alpha, target, nil)
	if len(closest.entries) == 0 {
		return
	}

	bucIdx := dht.getBucketIndex(target, dht.self.id)
	dht.buckets[bucIdx].resetRefreshTime()

	dht.net.FindNode(closest)
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

func (dht *DHT) getClosestNodes(num int, target nodeID, ignored []*netnode) *nodesByDistance {
	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	idx := dht.getBucketIndex(dht.self.id, target)
	idxList := []int{idx}
	l := idx - 1
	r := idx + 1

	for len(idxList) < keyBits {
		if l >= 0 {
			idxList = append(idxList, l)
		}
		if r < keyBits {
			idxList = append(idxList, r)
		}
	}

	nodes := &nodesByDistance{target: target}
	toAdd := num

	for toAdd > 0 && len(idxList) > 0 {
		idx, idxList = idxList[0], idxList[1:]

		for _, n := range dht.buckets[idx].entries {
			ignore := false
			for _, v := range ignored {
				if n.id == v.id {
					ignore = true
					break
				}
			}

			if !ignore {
				nodes.appendUnique(n)
				toAdd--
				if toAdd == 0 {
					break
				}
			}
		}
	}

	sort.Sort(nodes)

	return nodes
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

	return b.findNodeIndex(id) != -1
}

func (dht *DHT) doRefresh() {
	for i, b := range dht.buckets {
		if b.expired() {
			nodeID := dht.getRandomID(i)
			dht.iterate(nodeID)
		}
	}
}

func (dht *DHT) timers() {
	refresh := time.NewTicker(refreshInterval)
	defer refresh.Stop()

	for {
		select {
		case <-refresh.C:
			dht.updateSeed()
			dht.doRefresh()
		case <-dht.closeCh:
			return
		}
	}
}

func (dht *DHT) loop() {
	for {
		select {
		case node := <-dht.net.OnFindNode():
			closest := dht.getClosestNodes(bucketSize, node.id, nil)
			dht.net.SendNodes(node, closest.entries)
		case nodes := <-dht.net.Nodes():
			for _, n := range nodes {
				go dht.addNode(n)
			}
		case <-dht.closeCh:
			return
		}
	}
}

func (dht *DHT) getRandomID(bucket int) nodeID {
	dht.mutex.Lock()
	defer dht.mutex.Unlock()

	var id []byte

	byteIdx := bucket / 8
	bitIdx := bucket % 8

	for i := 0; i < byteIdx; i++ {
		id = append(id, dht.self.id[i])
	}

	var b byte
	for i := 0; i < 8; i++ {
		var bit bool
		if i < bitIdx {
			bit = HasBit(dht.self.id[byteIdx], uint(i))
		} else {
			bit = rand.Intn(2) == 1
		}

		if bit {
			pos := 7 - i
			b += byte(math.Pow(2, float64(pos)))
		}
	}

	id = append(id, b)

	for i := byteIdx + 1; i < len(nodeID{}); i++ {
		r := byte(rand.Intn(256))
		id = append(id, r)
	}

	var nID nodeID
	copy(nID[:], id)

	return nID
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

				return keyBits - (byteIdx + bitIdx) - 1
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
	entries     []*netnode
	lastRefresh time.Time
}

func (b *bucket) expired() bool {
	return time.Since(b.lastRefresh) > refreshInterval
}

func (b *bucket) resetRefreshTime() {
	b.lastRefresh = time.Now()
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
