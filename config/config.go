package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Addresses map[string]Peer
}

type Client struct {
	Config *Config
}

func New(path string) (*Client, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Config: &cfg,
	}

	return c, nil
}
