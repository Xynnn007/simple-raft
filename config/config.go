package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type NodeConfig struct {
	Peers   map[string]Peer
	Address string
	Port    string
	Id      int
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
