package config

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

var (
	path = "../test/peers.toml"
)

func Test_New(t *testing.T) {
	cfg, err := New(path)
	assert.NoError(t, err)

	for k, v := range cfg.Config.Addresses {
		log.Infof("%s : %s", k, v)
	}
}
