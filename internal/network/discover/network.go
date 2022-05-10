package discover

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sort"
	"sync"

	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Network struct {
	mutex sync.Mutex
	self  *netnode
	conn  *net.UDPConn
	// pending stores all responses which we are waiting for.
	// Key is message.ID.
	pending map[uint64]*response
	// responding stores msg ids by nodes ids which are waiting
	// for response with nodes
	responding map[nodeID]uint64

	// findCh triggered when some node requests
	// our nodes list.
	// Access to it with FindNode().
	findCh chan *netnode
	// nodesCh sends out nodes gotten
	// after our ask sending to another node
	// or when some node requested us.
	// Access to it with Nodes().
	nodesCh chan []*netnode

	// counter increases on every sent message.
	counter uint64

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
		pending:    make(map[uint64]*response),
		responding: make(map[nodeID]uint64),
		findCh:     make(chan *netnode),
		nodesCh:    make(chan []*netnode),
		counter:    rand.Uint64(),
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
}

// FindNode iterates through network with FindNode message.
func (n *Network) FindNode(closest *nodesByDistance) (int64, []nodeID) {
	if len(closest.entries) == 0 {
		return 0, nil
	}

	contacted := make(map[nodeID]struct{})

	queryRest := false
	closestNode := closest.entries[0].ID

	unavailableNodes := make([]nodeID, 0)
	responses := make([]*awaitedResponse, 0, len(closest.entries))

	for {
		for i, node := range closest.entries {
			if i > alpha && !queryRest {
				break
			}

			if _, ok := contacted[node.ID]; ok {
				continue
			}
			contacted[node.ID] = struct{}{}

			resp, err := n.sendFindNode(node)
			if err != nil {
				unavailableNodes = append(unavailableNodes, node.ID)
				continue
			}
			awaited := resp.await(responseTimeout)

			responses = append(responses, awaited)
		}

		wg := sync.WaitGroup{}
		collected := make(chan []*netnode)
		// Start to wait responses from nodes
		addedTotal := atomic.NewInt64(0)
		for _, r := range responses {
			wg.Add(1)
			go func(r *awaitedResponse, wg *sync.WaitGroup) {
				defer wg.Done()

				select {
				case <-r.received():
					nodes, ok := r.data().([]*netnode)
					if !ok {
						n.logger.Debugf("received %T instead of nodes list", r.data())

						return
					}
					collected <- nodes
					addedTotal.Inc()
				case <-r.timeout():
					// Node is responding for too long, add it to list of unavailable nodes.
					unavailableNodes = append(unavailableNodes, r.from())
				}
			}(r, &wg)
		}

		for _, v := range unavailableNodes {
			closest.remove(v)
		}

		// Wait until all responses are proceeded and close
		// chan with collected nodes.
		go func() {
			wg.Wait()
			close(collected)
		}()

		if len(responses) > 0 {
			for nodes := range collected {
				closest.appendUnique(nodes...)
			}
		}

		if len(closest.entries) == 0 {
			return 0, nil
		}

		sort.Sort(closest)

		//  If the closest node is unchanged then iteration is done.
		if bytes.Compare(closest.entries[0].ID.Bytes(), closestNode.Bytes()) == 0 || queryRest {
			if !queryRest {
				queryRest = true
				continue
			}

			return addedTotal.Load(), unavailableNodes
		} else {
			closestNode = closest.entries[0].ID
		}
	}
}

// ping sends ping message and waits for answer.
func (n *Network) ping(node *netnode) (nodeID, bool) {
	resp, err := n.sendPing(node)
	if err != nil {
		return nilNodeID, false
	}
	awaited := resp.await(pingTimeout)

	select {
	case <-awaited.received():
		return awaited.from(), true
	case <-awaited.timeout():
		return nilNodeID, false
	}
}

// sendFindNode sends FindNode message to the peer node and starts pending response
// with list of nodes.
func (n *Network) sendFindNode(dst *netnode) (*response, error) {
	msg := &message{
		ID:   n.counter,
		From: n.self,
		Type: FindNode,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with node list
	resp := makeResponse(dst.ID, Nodes)
	n.addPending(msg.ID, resp)

	return resp, nil
}

// sendNodes sends message with nodes list to the peer node.
func (n *Network) sendNodes(dst *netnode, nodes []*netnode) error {
	msg := &message{
		From: n.self,
		Type: Nodes,
		Data: nodes,
	}

	n.mutex.Lock()
	msgID, ok := n.responding[dst.ID]
	n.mutex.Unlock()
	if ok {
		msg.ID = msgID
	} else {
		msg.ID = n.counter
	}

	if err := n.send(dst, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// sendPing sends message to the peer node and starts pending response
// with Pong message. Returns response, which can be used for awaiting.
func (n *Network) sendPing(dst *netnode) (*response, error) {
	msg := &message{
		ID:   n.counter,
		From: n.self,
		Type: Ping,
		Data: nil,
	}

	if err := n.send(dst, msg); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	// Start pending for the response with pong message
	resp := makeResponse(dst.ID, Pong)
	n.addPending(msg.ID, resp)

	return resp, nil
}

// sendPong sends Pong message to the node which pinged us.
func (n *Network) sendPong(dst *netnode, msgID uint64) error {
	msg := &message{
		ID:   msgID,
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

	n.mutex.Lock()
	n.counter++
	n.mutex.Unlock()

	return nil
}

// listen waits for incoming UDP packets.
func (n *Network) listen() {
	buf := make([]byte, 2048)
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
	msg, err := deserializeMessage(bytes.NewBuffer(data))
	if err != nil {
		err = fmt.Errorf("deserialize message: %w", err)
		n.logger.Debugf("handle message failed: %s", err)

		return
	}

	n.mutex.Lock()
	resp, ok := n.pending[msg.ID]
	n.mutex.Unlock()
	if !ok {
		n.handleRequest(msg.From, msg)

		return
	}
	resp.from = msg.From.ID
	resp.data = msg.Data
	resp.markReceived()

	n.removePending(msg.ID)
}

// handleRequest processes message from other node as request
// because we were not pending it.
// Caller must hold  mutex.
func (n *Network) handleRequest(node *netnode, msg *message) {
	switch msg.Type {
	case Ping:
		if err := n.sendPong(node, msg.ID); err != nil {
			n.logger.Debugf("send pong message: %s", err)

			return
		}
	case FindNode:
		n.responding[node.ID] = msg.ID
		n.findCh <- node
	}
}

func (n *Network) addPending(id uint64, resp *response) {
	n.mutex.Lock()
	n.pending[id] = resp
	n.mutex.Unlock()
}

func (n *Network) removePending(id uint64) {
	n.mutex.Lock()
	delete(n.pending, id)
	n.mutex.Unlock()
}
