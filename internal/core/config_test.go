package core_test

import (
	"testing"

	"github.com/madhuravius/brains/internal/aws"
	brainsConfig "github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/core"
	mockBrains "github.com/madhuravius/brains/internal/mock"

	"github.com/stretchr/testify/assert"
)

func TestCoreConfigLifecycle(t *testing.T) {
	orig := &aws.AWSConfig{}
	brainsCfg := &brainsConfig.BrainsConfig{}
	logger := &mockBrains.TestLogger{}

	c := core.NewCoreConfig(orig, brainsCfg)
	c.SetLogger(logger)
	assert.NotNil(t, c)
}
