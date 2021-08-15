package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type StorageService interface {
	Chain() IChainStorage
}

type Storage struct {
	db    *leveldb.DB
	chain *ChainStorage
}

var _ StorageService = (*Storage)(nil)

func New(db *leveldb.DB) StorageService {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Chain() IChainStorage {
	if s.chain != nil {
		return s.chain
	}

	s.chain = &ChainStorage{
		storage: s,
	}

	return s.chain
}
