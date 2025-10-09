package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"

	mockBrains "github.com/madhuravius/brains/internal/mock"
)

func TestSetAndValidateCredentialsSuccess(t *testing.T) {
	origLoad := loadConfigFunc
	origNewSTS := newSTSClientFunc
	defer func() {
		loadConfigFunc = origLoad
		newSTSClientFunc = origNewSTS
	}()

	loadConfigFunc = func(ctx context.Context, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}
	mockClient := &mockBrains.MockSTSClient{
		Output: &sts.GetCallerIdentityOutput{
			Arn:     aws.String("arn:aws:sts::123456789012:assumed-role/role-name"),
			Account: aws.String("123456789012"),
		},
		Err: nil,
	}
	newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
		return mockClient
	}

	cfg := NewAWSConfig("us-west-2")
	ok := cfg.SetAndValidateCredentials()
	assert.True(t, ok)
}

func TestSetAndValidateCredentialsLoadError(t *testing.T) {
	origLoad := loadConfigFunc
	origNewSTS := newSTSClientFunc
	defer func() {
		loadConfigFunc = origLoad
		newSTSClientFunc = origNewSTS
	}()

	loadConfigFunc = func(ctx context.Context, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load error")
	}
	newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
		return &mockBrains.MockSTSClient{}
	}

	cfg := NewAWSConfig("us-west-2")
	ok := cfg.SetAndValidateCredentials()
	assert.False(t, ok)
}

func TestSetAndValidateCredentialsInvalidCredentials(t *testing.T) {
	origLoad := loadConfigFunc
	origNewSTS := newSTSClientFunc
	defer func() {
		loadConfigFunc = origLoad
		newSTSClientFunc = origNewSTS
	}()

	loadConfigFunc = func(ctx context.Context, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}
	mockClient := &mockBrains.MockSTSClient{
		Output: nil,
		Err:    errors.New("invalid credentials"),
	}
	newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
		return mockClient
	}

	cfg := NewAWSConfig("us-west-2")
	ok := cfg.SetAndValidateCredentials()
	assert.False(t, ok)
}
