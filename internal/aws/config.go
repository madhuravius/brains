package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"
)

type AWSConfig struct {
	cfg                 aws.Config
	defaultBedrockModel string
}

func NewAWSConfig(model string) *AWSConfig {
	return &AWSConfig{
		defaultBedrockModel: model,
	}
}

func (a *AWSConfig) SetAndValidateCredentials() bool {
	pterm.Info.Println("checking AWS credentials")
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		pterm.Error.Printf("unable to load SDK config, %s\n", err.Error())
		return false
	}
	a.cfg = cfg
	client := sts.NewFromConfig(cfg)
	resp, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		pterm.Error.Printf("credentials invalid: %s\n", err.Error())
		return false
	}
	pterm.Info.Printf("Valid credentials for ARN: %s (Account: %s)\n", *resp.Arn, *resp.Account)
	return true
}
