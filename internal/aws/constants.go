package aws

import (
	_ "embed"
)

//go:embed data/models_pricing.json
var rawModelsPricing []byte

// token limit is still a fixed safety bound (128 000)
const TokenLimit = 128000
