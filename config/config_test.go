package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := Config()
	assert.NotNil(t, cfg)

	cfg.SetupLogging()
	assert.NotNil(t, cfg.log)
}
