package storage

import (
	"errors"
	"fmt"
)

type IChainStorage interface {
	GetHeight() int
	Save([]byte, []byte) error
	GetLast() ([]byte, []byte, error)
	GetLastHash() ([]byte, error)
}

type ChainStorage struct {
	storage *Storage
}

var ErrBlockNotFound = errors.New("block not found")

func (s *ChainStorage) Save(hash, data []byte) error {
	err := s.storage.db.Put(hash, data, nil)
	if err != nil {
		return fmt.Errorf("put: %w", err)
	}

	return nil
}

func (s *ChainStorage) GetHeight() int {
	var height int

	iter := s.storage.db.NewIterator(nil, nil)
	for iter.Next() {
		height++
	}

	return height
}

func (s *ChainStorage) GetLast() ([]byte, []byte, error) {
	iter := s.storage.db.NewIterator(nil, nil)

	if ok := iter.Last(); !ok {
		return nil, nil, ErrBlockNotFound
	}

	return iter.Key(), iter.Value(), nil
}

func (s *ChainStorage) GetLastHash() ([]byte, error) {
	iter := s.storage.db.NewIterator(nil, nil)

	if ok := iter.Last(); !ok {
		return nil, ErrBlockNotFound
	}

	return iter.Key(), nil
}
