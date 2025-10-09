package aws

import (
	_ "embed"
)

//go:embed data/models_pricing.json
var rawModelsPricing []byte
