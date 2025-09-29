package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"
)

type STSClient interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// AWSConfig holds configuration needed to interact with AWS Bedrock.
type AWSConfig struct {
	cfg                 aws.Config
	defaultBedrockModel string
	region              string
	invoker             BedrockInvoker
}

var loadConfigFunc = config.LoadDefaultConfig
var newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
	return sts.NewFromConfig(cfg, optFns...)
}

// NewAWSConfig creates a new AWSConfig with the supplied model ID and region.
func NewAWSConfig(model string, region string) *AWSConfig {
	return &AWSConfig{
		defaultBedrockModel: model,
		region:              region,
	}
}

// SetAndValidateCredentials loads the AWS SDK configuration for the configured
// region and validates the credentials by calling STS GetCallerIdentity.
func (a *AWSConfig) SetAndValidateCredentials() bool {
	pterm.Info.Println("checking AWS credentials")
	cfg, err := loadConfigFunc(context.Background(), config.WithRegion(a.region))
	if err != nil {
		pterm.Error.Printf("unable to load SDK config, %s\n", err.Error())
		return false
	}
	a.cfg = cfg
	client := newSTSClientFunc(cfg)
	_, err = client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		pterm.Error.Printf("credentials invalid: %s\n", err.Error())
		return false
	}
	pterm.Info.Println("Valid credentials")
	return true
}
