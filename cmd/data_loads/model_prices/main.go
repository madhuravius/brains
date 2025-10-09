package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/pterm/pterm"

	brainsAws "github.com/madhuravius/brains/internal/aws"
)

const REGION = "us-east-1"

type ModelDetails struct {
	ModelID      string
	ProviderName string
}

type PricingDetails struct {
	InputCostPer1kTokens  float64
	OutputCostPer1kTokens float64
}

type PriceData struct {
	Product ProductInfo `json:"product"`
	Terms   struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				Unit         string `json:"unit"`
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

type ProductInfo struct {
	Attributes struct {
		ModelName     string `json:"model"`
		InferenceType string `json:"inferenceType"`
	} `json:"attributes"`
}

func getAvailableModels(ctx context.Context, cfg aws.Config) (map[string]ModelDetails, error) {
	client := bedrock.NewFromConfig(cfg)
	resp, err := client.ListFoundationModels(ctx, &bedrock.ListFoundationModelsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list foundation models: %w", err)
	}

	models := make(map[string]ModelDetails)
	for _, summary := range resp.ModelSummaries {
		if summary.ModelName != nil && summary.ModelId != nil {
			models[aws.ToString(summary.ModelName)] = ModelDetails{
				ModelID:      aws.ToString(summary.ModelId),
				ProviderName: aws.ToString(summary.ProviderName),
			}
		}
	}
	return models, nil
}

func getBedrockPricing(ctx context.Context) (map[string]PricingDetails, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for pricing: %w", err)
	}

	client := pricing.NewFromConfig(cfg)
	paginator := pricing.NewGetProductsPaginator(client, &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonBedrock"),
	})

	modelPrices := make(map[string]PricingDetails)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next pricing page: %w", err)
		}

		for _, priceJSON := range page.PriceList {
			var priceData PriceData
			if err := json.Unmarshal([]byte(priceJSON), &priceData); err != nil {
				log.Printf("warning: failed to unmarshal price data: %v", err)
				continue
			}

			modelName := strings.TrimSpace(priceData.Product.Attributes.ModelName)
			if modelName == "" {
				continue
			}

			currentPrices := modelPrices[modelName]
			for _, term := range priceData.Terms.OnDemand {
				for _, dim := range term.PriceDimensions {
					price, err := strconv.ParseFloat(dim.PricePerUnit.USD, 64)
					if err != nil {
						continue
					}
					switch {
					case strings.Contains(priceData.Product.Attributes.InferenceType, "Input"):
						currentPrices.InputCostPer1kTokens = price
					case strings.Contains(priceData.Product.Attributes.InferenceType, "Output"):
						currentPrices.OutputCostPer1kTokens = price
					}
				}
			}
			modelPrices[modelName] = currentPrices
		}
	}
	return modelPrices, nil
}

func main() {
	fmt.Printf("fetching data for region: %s\n\n", REGION)
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(REGION))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	var wg sync.WaitGroup
	var availableModels map[string]ModelDetails
	var modelPricing map[string]PricingDetails
	var modelsErr, pricingErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		availableModels, modelsErr = getAvailableModels(ctx, cfg)
		if modelsErr == nil {
			log.Println("✅ finished fetching available models.")
		}
	}()

	go func() {
		defer wg.Done()
		modelPricing, pricingErr = getBedrockPricing(ctx)
		if pricingErr == nil {
			log.Println("✅ finished fetching pricing information.")
		}
	}()

	wg.Wait()

	if modelsErr != nil {
		log.Fatalf("error fetching available models: %v", modelsErr)
	}
	if pricingErr != nil {
		log.Fatalf("error fetching pricing information: %v", pricingErr)
	}

	var combinedData []brainsAws.AggregatedModelPricing

	for modelName, details := range availableModels {
		for priceName, prices := range modelPricing {
			if strings.Contains(strings.ToLower(modelName), strings.ToLower(priceName)) ||
				strings.Contains(strings.ToLower(priceName), strings.ToLower(modelName)) {
				combinedData = append(combinedData, brainsAws.AggregatedModelPricing{
					ModelName:             modelName,
					ModelID:               details.ModelID,
					ProviderName:          details.ProviderName,
					InputCostPer1kTokens:  prices.InputCostPer1kTokens,
					OutputCostPer1kTokens: prices.OutputCostPer1kTokens,
				})
				break
			}
		}
	}

	if len(combinedData) == 0 {
		fmt.Println("\n⚠️ no models with corresponding pricing found.")
		fmt.Printf("this could be because:\n")
		fmt.Printf("  1. model access is not enabled for your account in the '%s' region.\n", REGION)
		fmt.Printf("  2. the pricing API did not return data for this region.\n")
		return
	}

	sort.Slice(combinedData, func(i, j int) bool {
		return combinedData[i].ModelName < combinedData[j].ModelName
	})

	fmt.Println("\n--- combined model and pricing data ---")
	for _, item := range combinedData {
		fmt.Printf("model name: %s\n", item.ModelName)
		fmt.Printf("  - model id: %s\n", item.ModelID)
		fmt.Printf("  - provider: %s\n", item.ProviderName)
		fmt.Printf("  - input cost / 1k tokens: $%.6f\n", item.InputCostPer1kTokens)
		fmt.Printf("  - output cost / 1k tokens: $%.6f\n", item.OutputCostPer1kTokens)
		fmt.Println("----------------------------------------")
	}

	jsonBytes, err := json.MarshalIndent(combinedData, "", "  ")
	if err != nil {
		pterm.Error.Printf("failed to marshal pricing data to json: %v\n", err)
	} else {
		if mkErr := os.MkdirAll("data", 0o750); mkErr != nil {
			pterm.Error.Printf("failed to create data directory: %v\n", mkErr)
		} else {
			if writeErr := os.WriteFile("internal/aws/data/models_pricing.json", jsonBytes, 0o644); writeErr !=
				nil {
				pterm.Error.Printf("failed to write models_pricing.json: %v\n", writeErr)
			} else {
				pterm.Success.Println("pricing data written to data/models_pricing.json")
			}
		}
	}
}
