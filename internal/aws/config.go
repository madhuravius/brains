package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pterm/pterm"

	brainsConfig "github.com/madhuravius/brains/internal/config"
)

type STSClient interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

var loadConfigFunc = config.LoadDefaultConfig
var newSTSClientFunc = func(cfg aws.Config, optFns ...func(*sts.Options)) STSClient {
	return sts.NewFromConfig(cfg, optFns...)
}

func NewAWSConfig(region string) *AWSConfig {
	modelsPricing, err := getModelsPricing()
	if err != nil {
		return nil
	}

	return &AWSConfig{
		region:  region,
		pricing: modelsPricing,
	}
}

func getModelsPricing() ([]ModelPricing, error) {
	var out []ModelPricing
	if err := json.Unmarshal(rawModelsPricing, &out); err != nil {
		pterm.Error.Printf("unable to load SDK config, %s\n", err.Error())
		return nil, err
	}
	return out, nil
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
func (a *AWSConfig) SetPricing(pricing []ModelPricing)     { a.pricing = pricing }
