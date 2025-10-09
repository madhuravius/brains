package aws

import (
	"fmt"

	"github.com/pterm/pterm"
)

func (c *AWSConfig) pricingFor(modelID string) (modelPricing, bool) {
	for _, p := range c.pricing {
		if p.ModelID == modelID {
			return p, true
		}
	}
	return modelPricing{}, false
}

func (a *AWSConfig) PrintCost(usage map[string]any, modelID string) {
	p := modelPricing{}
	if val, ok := a.pricingFor(modelID); ok {
		p = val
	}
	promptTokens, completionTokens := 0, 0
	if v, ok := usage["prompt_tokens"]; ok {
		if n, ok := v.(float64); ok {
			promptTokens = int(n)
		}
	}
	if v, ok := usage["completion_tokens"]; ok {
		if n, ok := v.(float64); ok {
			completionTokens = int(n)
		}
	}
	cost := (float64(promptTokens)/1000.0)*p.InputCostPer1kTokens + (float64(completionTokens)/1000.0)*p.
		OutputCostPer1kTokens
	pterm.Info.Printf("estimated cost for this request: $%.6f (prompt %d, completion %d)\n", cost, promptTokens,
		completionTokens)
}

func (a *AWSConfig) PrintContext(usage map[string]any) {
	// token limit is still a fixed safety bound (128 000)
	const tokenLimit = 128000
	promptTokens, completionTokens := 0, 0
	if v, ok := usage["prompt_tokens"]; ok {
		if n, ok := v.(float64); ok {
			promptTokens = int(n)
		}
	}
	if v, ok := usage["completion_tokens"]; ok {
		if n, ok := v.(float64); ok {
			completionTokens = int(n)
		}
	}
	total := promptTokens + completionTokens
	pterm.Info.Printf("current context used: %d tokens (limit %d)\n", total, tokenLimit)
}

func (a *AWSConfig) PrintPricing(modelID string) error {
	tableData := pterm.TableData{{
		"Model ID",
		"Model Name",
		"Input Cost / 1k Tokens",
		"Output Cost / 1k Tokens",
	}}
	var activeModel modelPricing
	for _, p := range a.pricing {
		if modelID == p.ModelID {
			activeModel = p
		}
		tableData = append(tableData, []string{
			p.ModelID,
			p.ModelName,
			fmt.Sprintf("%f", p.InputCostPer1kTokens),
			fmt.Sprintf("%f", p.OutputCostPer1kTokens),
		})
	}
	if err := pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render(); err != nil {
		return err
	}
	fmt.Println()
	pterm.DefaultSection.Println("Active Model Pricing")
	pterm.Info.Printfln("Model ID: %s\nModel Name: %s\nInput Cost / 1k Tokens: %f\nOutput Cost / 1k Tokens: %f", activeModel.ModelID, activeModel.ModelName, activeModel.InputCostPer1kTokens, activeModel.OutputCostPer1kTokens)

	return nil
}
