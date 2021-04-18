package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Peer struct {
	Address string `toml:address`
	Port    string `toml:port`
	Id      int    `toml:id`
}

type NodeConfig struct {
	Address string `toml:address`
	Port    string `toml:port`
	Id      int    `toml:id`
	Peers   []Peer `toml:peers`
}

type Client struct {
	Config *NodeConfig
}

func New(path string) (*Client, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg NodeConfig
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Config: &cfg,
	}

	return c, nil
}
