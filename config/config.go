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
	GetNodeConfig() *NodeConfig
	GetChainConfig() *ChainConfig
	GetHolderConfig() *HolderConfig
}

type config struct {
	Node   *NodeConfig   `yaml:"node"`
	Chain  *ChainConfig  `yaml:"chain"`
	Holder *HolderConfig `yaml:"holder"`
}

type NodeConfig struct {
	Port      int    `yaml:"port"`
	Directory string `yaml:"dir"`
	PeersPath string `yaml:"peers"`

	Control []struct {
		Protocol string `yaml:"protocol"`
		Port     int    `yaml:"port"`
	} `yaml:"control"`

	Discover struct {
		Address        string `yaml:"address"`
		ListeningPort  int    `yaml:"listeningPort"`
		BootstrapNodes []Node `yaml:"bootnodes"`
		UseStun        bool   `yaml:"useStun"`
		StunAddr       string `yaml:"stunAddr"`
	} `yaml:"discover"`
}

type ChainConfig struct {
	Database string `yaml:"database"`
	Genesis  struct {
		Value float64 `yaml:"value"`
	} `yaml:"genesis"`
}

type HolderConfig struct {
	PrivateKeyPath string `yaml:"privateKey"`
}

type Node struct {
	Addr string `yaml:"addr"`
	Port int    `yaml:"port"`
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

func (c *config) GetNodeConfig() *NodeConfig {
	return c.Node
}

func (c *config) GetChainConfig() *ChainConfig {
	return c.Chain
}

func (c *config) GetHolderConfig() *HolderConfig {
	return c.Holder
}
