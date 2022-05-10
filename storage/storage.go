package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Storage interface {
	Chain() ChainStorage

	// DHT implements storage required by
	// distributed hash table.
	//
	// See http://github.com/AlexeySadkovich/eldberg/internal/network/discover/dht.go
	DHT() *dhtStorage
}

type storage struct {
	db    *leveldb.DB
	chain *chainStorage
	dht   *dhtStorage
}

var _ Storage = (*storage)(nil)

func New(db *leveldb.DB) Storage {
	return &storage{
		db: db,
	}
}

func (s *storage) Chain() ChainStorage {
	if s.chain != nil {
		return s.chain
	}

	s.chain = &chainStorage{
		storage: s,
	}

	return s.chain
}

func (s *storage) DHT() *dhtStorage {
	if s.dht != nil {
		return s.dht
	}

	s.dht = &dhtStorage{
		storage: s,
	}

	return s.dht
}
