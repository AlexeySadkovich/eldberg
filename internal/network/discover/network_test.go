package discover

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net"
	"testing"
	"time"
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

	select {
	case <-resp.received():
	case <-resp.timeout():
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

	resp, err := n2.sendFindNodeAwaited(n1.Self())

	n1.SendNodes(n2.Self(), expNodes)
	require.NoError(t, err)

	var resNodes []*netnode
	select {
	case resNodes = <-n2.nodesCh:
	case <-resp.timeout():
		require.Fail(t, "response time out")
	case <-time.After(2 * time.Second):
		require.Fail(t, "wait time out")
	}

	require.Equal(t, expNodes, resNodes)
}
