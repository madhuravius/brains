package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
)

type mockSTSClient struct {
	output *sts.GetCallerIdentityOutput
	err    error
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return m.output, m.err
}

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
	mockClient := &mockSTSClient{
		output: &sts.GetCallerIdentityOutput{
			Arn:     aws.String("arn:aws:sts::123456789012:assumed-role/role-name"),
			Account: aws.String("123456789012"),
		},
		err: nil,
	}
	newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
		return mockClient
	}

	cfg := NewAWSConfig("model", "us-west-2")
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
		return &mockSTSClient{}
	}

	cfg := NewAWSConfig("model", "us-west-2")
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
	mockClient := &mockSTSClient{
		output: nil,
		err:    errors.New("invalid credentials"),
	}
	newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
		return mockClient
	}

	cfg := NewAWSConfig("model", "us-west-2")
	ok := cfg.SetAndValidateCredentials()
	assert.False(t, ok)
}
