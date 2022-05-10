package discover

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNetwork_PingPong(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	id1 := newID()
	n1, err := newNetwork(id1, "127.0.0.1", 7000, logger.Sugar())
	require.NoError(t, err)
	go n1.listen()
	id2 := newID()
	n2, err := newNetwork(id2, "127.0.0.1", 7001, logger.Sugar())
	require.NoError(t, err)
	go n2.listen()

	resp, err := n1.sendPing(n2.Self())
	require.NoError(t, err)
	awaited := resp.await(pingTimeout)

	select {
	case <-awaited.received():
	case <-awaited.timeout():
		require.Fail(t, "response time out")
	}
}

func TestNetwork_SendNodes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	id1 := newID()
	n1, err := newNetwork(id1, "127.0.0.1", 7000, logger.Sugar())
	require.NoError(t, err)
	go n1.listen()
	id2 := newID()
	n2, err := newNetwork(id2, "127.0.0.1", 7001, logger.Sugar())
	require.NoError(t, err)
	go n2.listen()

	expNodes := []*netnode{
		{
			IP:   net.ParseIP("127.0.0.2"),
			Port: 2222,
		},
		{
			IP:   net.ParseIP("127.0.0.3"),
			Port: 2222,
		},
	}

	resp, err := n2.sendFindNode(n1.Self())
	require.NoError(t, err)
	awaited := resp.await(responseTimeout)

	<-n1.OnFindNode()
	n1.SendNodes(n2.Self(), expNodes)

	select {
	case <-awaited.received():
		nodes, ok := awaited.data().([]*netnode)
		require.True(t, ok)
		require.Equal(t, expNodes, nodes)
	case <-awaited.timeout():
		require.Fail(t, "response time out")
	case <-time.After(3 * time.Second):
		require.Fail(t, "wait time out")
	}
}
