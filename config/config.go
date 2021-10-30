package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/AlexeySadkovich/eldberg/internal/utils"
)

const (
	TXSLIMIT = 13

	CtxTimeout = 2 * time.Second
)

type Config struct {
	Node struct {
		Port           int    `yaml:"port"`
		Directory      string `yaml:"dir"`
		PeersPath      string `yaml:"peers"`
		PrivateKeyPath string `yaml:"privateKey"`
		Database       string `yaml:"database"`
		Control        struct {
			Port int `yaml:"port"`
		} `yaml:"control"`
	} `yaml:"node"`
}

func New() (*Config, error) {
	file, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	cfg := new(Config)

	if err := yaml.Unmarshal(file, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if !utils.IsDirectoryExists(cfg.Node.Directory) {
		if err := utils.CreateDirectory(cfg.Node.Directory); err != nil {
			return nil, fmt.Errorf("create node directory: %w", err)
		}
	}

	return cfg, nil
}
