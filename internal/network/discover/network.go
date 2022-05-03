package discover

import (
	"bytes"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

type Network struct {
	mutex sync.Mutex
	self  *netnode
	conn  *net.UDPConn
	// pending stores all nodes which from we
	// are waiting for the response.
	pending map[nodeID]*response
	// responding stores all nodes which are waiting
	// for the response with nodes list from our node.
	responding map[nodeID]*netnode
	// unknown stores addresses which may represent not known nodes.
	unknown map[string]struct{}

	// findCh triggered when some node requests
	// our nodes list.
	// Access to it with FindNode().
	findCh chan *netnode
	// nodesCh sends out nodes gotten
	// after our ask sending to another node
	// or when some node requested us.
	// Access to it with Nodes().
	nodesCh chan []*netnode

	logger *zap.SugaredLogger
}

const (
	maxPacketSize = 1024
)

func newNetwork(id nodeID, ip string, port int, logger *zap.SugaredLogger) (*Network, error) {
	self := &netnode{
		ID:   id,
		IP:   net.ParseIP(ip),
		Port: port,
	}

	conn, err := net.ListenUDP("udp", self.addr())
	if err != nil {
		return nil, fmt.Errorf("listen UDP: %w", err)
	}

	n := &Network{
		self:       self,
		conn:       conn,
		pending:    make(map[nodeID]*response),
		responding: make(map[nodeID]*netnode),
		unknown:    make(map[string]struct{}),
		findCh:     make(chan *netnode),
		nodesCh:    make(chan []*netnode),
		logger:     logger,
	}

	return n, nil
}

func (n *Network) Self() *netnode {
	return n.self
}

func (n *Network) shutdown() error {
	if err := n.conn.Close(); err != nil {
		return fmt.Errorf("close connection: %w", err)
	}

	return nil
}

// OnFindNode returns chan which sends node on new FindNode message from it.
func (n *Network) OnFindNode() chan *netnode {
	return n.findCh
}

// Nodes returns chan which sends lists of new netnodes.
func (n *Network) Nodes() chan []*netnode {
	return n.nodesCh
}

func (n *Network) SendNodes(dst *netnode, nodes []*netnode) {
	if err := n.sendNodes(dst, nodes); err != nil {
		n.logger.Debugf("send nodes error: %s", err)

		return
	}

	// Delete node from responding nodes list if there
	// was not any error during sending.
	delete(n.responding, dst.ID)
}

// BroadcastFindNode performs asking for nodes for all passed nodes and
// returns list of unavailable nodes.
func (n *Network) BroadcastFindNode(nodes []*netnode) []nodeID {
	unavailableNodes := make([]nodeID, 0)

	for _, v := range nodes {
		// Check if id is unknown for this node to avoid using nil id as key.
		if bytes.Compare(v.ID.Bytes(), nilNodeID.Bytes()) == 0 {
			if err := n.sendFindNode(v); err != nil {
				continue
			}
			// Add node to unknown list to await response without knowledge of its id
			n.addUnknown(v)
		} else {
			_, err := n.sendFindNodeAwaited(v)
			if err != nil {
				unavailableNodes = append(unavailableNodes, v.ID)
				continue
			}
		}
	}

	return unavailableNodes
}

func (n *Network) FindNode(closest *nodesByDistance) (int64, []nodeID) {
	contacted := make(map[nodeID]struct{})

	unavailableNodes := make([]nodeID, 0)
	responses := make([]*awaitedResponse, 0, len(closest.entries))

	for i, node := range closest.entries {
		if i > alpha {
			break
		}

		if _, ok := contacted[node.ID]; ok {
			continue
		}
		contacted[node.ID] = struct{}{}

		resp, err := n.sendFindNodeAwaited(node)
		if err != nil {
			unavailableNodes = append(unavailableNodes, node.ID)
			continue
		}

		responses = append(responses, resp)
	}

	// Start to wait responses from nodes
	addedTotal := atomic.NewInt64(0)
	wg := sync.WaitGroup{}
	for _, r := range responses {
		wg.Add(1)
		go func(r *awaitedResponse, wg *sync.WaitGroup) {
			defer wg.Done()

			select {
			case <-r.received():
				addedTotal.Inc()
			case <-r.timeout():
				// Node is responding for too long, add it to list of unavailable nodes.
				unavailableNodes = append(unavailableNodes, r.from())
			}
		}(r, &wg)
	}

	wg.Wait()

	for _, v := range unavailableNodes {
		closest.remove(v)
	}

	return addedTotal.Load(), unavailableNodes
}

func (n *Network) IsNodeAvailable(node *netnode) bool {
	resp, err := n.sendPing(node)
	if err != nil {
		return false
	}

	select {
	case <-resp.received():
		return true
	case <-resp.timeout():
		return false
	}
}

// sendFindNodeAwaited sends FindNode message to the peer node and starts pending response
// with list of nodes.
func (n *Network) sendFindNodeAwaited(dst *netnode) (*awaitedResponse, error) {
	msg := &message{
		From: n.self,
		Type: FindNode,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with node list
	resp := makeResponse(dst.ID, Nodes)
	n.addPending(dst.ID, resp)
	awaited := resp.await(responseTimeout)

	return awaited, nil
}

// sendFindNode sends FindNode message to the peer node and doesn't start pending for the response.
func (n *Network) sendFindNode(dst *netnode) error {
	msg := &message{
		From: n.self,
		Type: FindNode,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// sendNodes sends message with nodes list to the peer node.
func (n *Network) sendNodes(dst *netnode, nodes []*netnode) error {
	msg := &message{
		From: n.self,
		Type: Nodes,
		Data: nodes,
	}

	if err := n.send(dst, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// sendPing sends message to the peer node and starts pending response
// with Pong message. Returns awaited response, which can be used to check response state.
func (n *Network) sendPing(dst *netnode) (*awaitedResponse, error) {
	msg := &message{
		From: n.self,
		Type: Ping,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with pong message
	resp := makeResponse(dst.ID, Pong)
	n.addPending(dst.ID, resp)
	awaited := resp.await(pingTimeout)

	return awaited, nil
}

// sendPong sends Pong message to the node which pinged us.
func (n *Network) sendPong(dst *netnode) error {
	msg := &message{
		From: n.self,
		Type: Pong,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (n *Network) send(dst *netnode, msg *message) error {
	addr := dst.addr()

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	data, err := serializeMessage(msg)
	if err != nil {
		return fmt.Errorf("serialize message: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

// listen waits for incoming UDP packets.
func (n *Network) listen() {
	buf := make([]byte, maxPacketSize)
	for {
		_, _, err := n.conn.ReadFromUDP(buf)
		if IsTemporaryErr(err) {
			n.logger.Debugf("read from UDP temp error: %s", err)
		} else if err != nil {
			if !errors.Is(err, io.EOF) {
				n.logger.Errorf("read from UDP: %s", err)
			}

			continue
		}

		go n.handleMessage(buf)
	}
}

// handleMessage handles all incoming messages and decides
// which way this message should be processed.
func (n *Network) handleMessage(data []byte) {
	msg, err := deserializeMessage(data)
	if err != nil {
		err = fmt.Errorf("deserialize message: %w", err)
		n.logger.Debugf("handle message failed: %s", err)
	}

	unknown := n.isUnknown(msg.From)
	if unknown {
		switch {
		case msg.isRequest():
			n.handleRequest(msg.From, msg)
		case msg.isResponse():
			n.handleResponse(msg.From, msg)
		}

		return
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	resp, ok := n.pending[msg.From.ID]
	if !ok {
		n.handleRequest(msg.From, msg)

		return
	}
	resp.markReceived()

	if resp.typ == msg.Type {
		n.handleResponse(msg.From, msg)
	} else {
		n.handleRequest(msg.From, msg)
	}
}

// handleResponse processes message from other node as response
// because we were pending it.
func (n *Network) handleResponse(node *netnode, msg *message) {
	switch msg.Type {
	case Pong:
	case Nodes:
		nodes, ok := msg.Data.([]*netnode)
		if !ok {
			n.logger.Debugf("received %T instead of nodes list", msg.Data)

			return
		}

		// send new nodes to DHT
		n.nodesCh <- nodes
	default:
	}

	n.removePending(node.ID)
}

// handleRequest processes message from other node as request
// because we were not pending it.
func (n *Network) handleRequest(node *netnode, msg *message) {
	switch msg.Type {
	case Ping:
		if err := n.sendPong(node); err != nil {
			n.logger.Debugf("send pong message: %s", err)

			return
		}
	case FindNode:
		n.findCh <- node
		n.responding[node.ID] = node
	}
}

func (n *Network) addUnknown(node *netnode) {
	k := node.addr().String()
	n.mutex.Lock()
	n.unknown[k] = struct{}{}
	n.mutex.Unlock()
}

func (n *Network) isUnknown(node *netnode) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	k := node.addr().String()
	_, ok := n.unknown[k]

	return ok
}

func (n *Network) removeUnknown(node *netnode) {
	k := node.addr().String()
	n.mutex.Lock()
	delete(n.unknown, k)
	n.mutex.Unlock()
}

func (n *Network) addPending(id nodeID, resp *response) {
	n.mutex.Lock()
	n.pending[id] = resp
	n.mutex.Unlock()
}

func (n *Network) removePending(id nodeID) {
	n.mutex.Lock()
	delete(n.pending, id)
	n.mutex.Unlock()
}
