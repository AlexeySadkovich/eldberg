package discover

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
)

type Storage interface {
	Save(key, data []byte, replication, expiration time.Time) error
	Get(key []byte) (data []byte, err error)
	Delete(key []byte)
}

// DHT is distributed hash table used in
// Kademlia protocol implementation
type DHT struct {
	table   *Table
	net     *Network
	storage Storage
}

const (
	localAddr = "127.0.0.1"
)

func New(storage Storage, config config.Config, logger *zap.SugaredLogger) (*DHT, error) {
	nodeConfig := config.GetNodeConfig()
	port := nodeConfig.Discover.ListeningPort

	table, err := newTable(localAddr, port)
	if err != nil {
		return nil, fmt.Errorf("create hash table: %w", err)
	}

	netw, err := newNetwork(localAddr, port, logger)
	if err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}

	dht := &DHT{
		table:   table,
		net:     netw,
		storage: storage,
	}

	return dht, nil
}
