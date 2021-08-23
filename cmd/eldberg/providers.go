package main

import (
	"fmt"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
)

func provideLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()

	return logger.Sugar()
}

func provideDB(config *config.Config) (*leveldb.DB, error) {
	path := filepath.Join(config.Node.Directory, config.Node.Database)

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return db, nil
}
