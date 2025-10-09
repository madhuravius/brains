package core_test

import (
	"testing"

	"github.com/madhuravius/brains/internal/aws"
	"github.com/madhuravius/brains/internal/core"
	mockBrains "github.com/madhuravius/brains/internal/mock"

	"github.com/stretchr/testify/assert"
)

func TestCoreConfigLifecycle(t *testing.T) {
	orig := &aws.AWSConfig{}
	logger := &mockBrains.TestLogger{}

	c := core.NewCoreConfig(orig)
	c.SetLogger(logger)
	assert.NotNil(t, c)
}
