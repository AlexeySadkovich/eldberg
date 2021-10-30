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

type Config interface {
	GetNodeConfig() *nodeConfig
	GetChainConfig() *chainConfig
	GetHolderConfig() *holderConfig
}

type config struct {
	Node   *nodeConfig   `yaml:"node"`
	Chain  *chainConfig  `yaml:"chain"`
	Holder *holderConfig `yaml:"holder"`
}

type nodeConfig struct {
	Port      int    `yaml:"port"`
	Directory string `yaml:"dir"`
	PeersPath string `yaml:"peers"`
	Control   struct {
		Port int `yaml:"port"`
	} `yaml:"control"`
}

type chainConfig struct {
	Database string `yaml:"database"`
	Genesis  struct {
		Value float64 `yaml:"value"`
	} `yaml:"genesis"`
}

type holderConfig struct {
	PrivateKeyPath string `yaml:"privateKey"`
}

func New() (Config, error) {
	file, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	cfg := new(config)

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

func (c *config) GetNodeConfig() *nodeConfig {
	return c.Node
}

func (c *config) GetChainConfig() *chainConfig {
	return c.Chain
}

func (c *config) GetHolderConfig() *holderConfig {
	return c.Holder
}
