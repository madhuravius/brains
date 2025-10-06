package core_test

import (
	"testing"

	"brains/internal/aws"
	"brains/internal/core"
	mockBrains "brains/internal/mock"

	"github.com/stretchr/testify/assert"
)

func TestCoreConfigLifecycle(t *testing.T) {
	orig := &aws.AWSConfig{}
	logger := &mockBrains.TestLogger{}

	c := core.NewCoreConfig(orig)
	c.SetLogger(logger)
	assert.NotNil(t, c)
}
