package discover

import (
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

type Network struct {
	mutex   sync.Mutex
	address *net.UDPAddr
	conn    *net.UDPConn
	// pending stores all nodes which from we
	// are waiting for the response.
	pending map[nodeID]*response
	// responding stores all nodes which are waiting
	// for the response with nodes list from our node.
	responding map[nodeID]*netnode

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

func newNetwork(ip string, port int, logger *zap.SugaredLogger) (*Network, error) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen UDP: %w", err)
	}

	n := &Network{
		address:    addr,
		conn:       conn,
		pending:    make(map[nodeID]*response),
		responding: make(map[nodeID]*netnode),
		findCh:     make(chan *netnode),
		nodesCh:    make(chan []*netnode),
		logger:     logger,
	}

	return n, nil
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

func (n *Network) SendNodes(node *netnode, nodes []*netnode) {
	if err := n.sendNodes(node, nodes); err != nil {
		n.logger.Debugf("send nodes error: %s", err)

		return
	}

	// Delete node from responding nodes list if there
	// was not any error during sending.
	delete(n.responding, node.id)
}

func (n *Network) FindNode(closest *nodesByDistance) (int64, []nodeID) {
	contacted := make(map[nodeID]struct{})
	nothingCloser := false

	unavailableNodes := make([]nodeID, 0)
	responses := make([]*awaitedResponse, 0, len(closest.entries))

	for i, node := range closest.entries {
		if i > alpha && !nothingCloser {
			break
		}

		if _, ok := contacted[node.id]; ok {
			continue
		}
		contacted[node.id] = struct{}{}

		resp, err := n.sendFindNode(node)
		if err != nil {
			unavailableNodes = append(unavailableNodes, node.id)
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
				// Node responding for too long, add it to unavailable
				unavailableNodes = append(unavailableNodes, r.from())
			}
		}(r, &wg)
	}

	wg.Wait()

	for _, v := range unavailableNodes {
		closest.remove(v)
	}

	if !nothingCloser && len(unavailableNodes) == 0 {
		return 0, nil
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

// sendFindNode sends message to the peer node and starts pending response
// with list of nodes.
func (n *Network) sendFindNode(node *netnode) (*awaitedResponse, error) {
	msg := &message{
		ID:   node.id,
		Type: FindNode,
		Data: nil,
	}

	if err := n.send(node, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with node list
	resp := makeResponse(node.id, Nodes)
	n.addPending(node.id, resp)
	awaited := resp.await(responseTimeout)

	return awaited, nil
}

// sendNodes sends message with nodes list to the peer node.
func (n *Network) sendNodes(node *netnode, nodes []*netnode) error {
	msg := &message{
		ID:   node.id,
		Type: Nodes,
		Data: nodes,
	}

	if err := n.send(node, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// sendPing sends message to the peer node and starts pending response
// with Pong message. Returns awaited response, which can be used to check response state.
func (n *Network) sendPing(node *netnode) (*awaitedResponse, error) {
	msg := &message{
		ID:   node.id,
		Type: Ping,
		Data: nil,
	}

	if err := n.send(node, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with pong message
	resp := makeResponse(node.id, Pong)
	n.addPending(node.id, resp)
	awaited := resp.await(pingTimeout)

	return awaited, nil
}

// sendPong sends Pong message to the node which pinged us.
func (n *Network) sendPong(node *netnode) error {
	msg := &message{
		ID:   node.id,
		Type: Pong,
		Data: nil,
	}

	if err := n.send(node, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (n *Network) send(node *netnode, msg *message) error {
	addr := node.addr()
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

// read waits for incoming UDP packets.
func (n *Network) read() {
	buf := make([]byte, maxPacketSize)
	for {
		_, from, err := n.conn.ReadFromUDP(buf)
		if IsTemporaryErr(err) {
			n.logger.Debugf("read from UDP temp error: %s", err)
		} else if err != nil {
			if !errors.Is(err, io.EOF) {
				n.logger.Errorf("read from UDP: %s", err)
			}

			return
		}

		if err := n.handleMessage(from, buf); err != nil {
			n.logger.Debugf("handle message failed: %s", err)
		}
	}
}

// handleMessage handles all incoming messages and decides
// which way this message should be processed.
func (n *Network) handleMessage(from *net.UDPAddr, data []byte) error {
	msg, err := deserializeMessage(data)
	if err != nil {
		return fmt.Errorf("deserialize message: %w", err)
	}

	node := &netnode{
		id:   msg.ID,
		ip:   from.IP,
		port: from.Port,
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	resp, ok := n.pending[msg.ID]
	if !ok {
		n.handleRequest(node, msg)

		return nil
	}
	resp.markReceived()

	if resp.typ == msg.Type {
		n.handleResponse(node, msg)
	} else {
		n.handleRequest(node, msg)
	}

	return nil
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

	n.removePending(node.id)
}

// handleRequest processes message from other node as request
// because we were not pending it.
func (n *Network) handleRequest(node *netnode, msg *message) {
	switch msg.Type {
	case Ping:
		if err := n.sendPong(node); err != nil {
			n.logger.Debugf("send pong message: %s", err)
		}
	case FindNode:
		n.findCh <- node
		n.responding[node.id] = node
	}
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
