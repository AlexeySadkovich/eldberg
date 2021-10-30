package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Storage interface {
	Chain() ChainStorage
}

type storage struct {
	db    *leveldb.DB
	chain *chainStorage
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
