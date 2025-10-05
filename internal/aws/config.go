package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"

	brainsConfig "brains/internal/config"
)

type STSClient interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type AWSConfig struct {
	cfg                 aws.Config
	defaultBedrockModel string
	region              string
	invoker             BedrockInvoker
	logger              brainsConfig.SimpleLogger
}

var loadConfigFunc = config.LoadDefaultConfig
var newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
	return sts.NewFromConfig(cfg, optFns...)
}

func NewAWSConfig(model string, region string) *AWSConfig {
	return &AWSConfig{
		defaultBedrockModel: model,
		region:              region,
	}
}

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
	pterm.Info.Println("valid credentials")
	return true
}

func (a *AWSConfig) GetConfig() aws.Config                 { return a.cfg }
func (a *AWSConfig) SetLogger(l brainsConfig.SimpleLogger) { a.logger = l }
