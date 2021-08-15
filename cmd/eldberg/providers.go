package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
)

func provideLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()

	return logger.Sugar()
}

func provideDB(config *config.Config) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(config.Node.Database, nil)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return db, nil
}
