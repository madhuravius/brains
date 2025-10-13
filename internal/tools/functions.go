package tools

func NewCommonToolsConfig(ignorePatterns []string) CommonToolsImpl {
	commonTools := CommonToolsConfig{
		ignorePatterns: ignorePatterns,
	}
	commonTools.HydrateRegexIgnorePatterns()
	return &commonTools
}

func (c *CommonToolsConfig) SetIgnorePatterns(ignorePatterns []string) {
	c.ignorePatterns = ignorePatterns
	c.HydrateRegexIgnorePatterns()
}

func (c *CommonToolsConfig) HydrateRegexIgnorePatterns() {
	patterns := make([]*regexPattern, len(c.ignorePatterns))
	for idx, line := range c.ignorePatterns {
		regexPatt, negateRegex := getPatternFromLine(line)
		patterns[idx] = &regexPattern{
			pattern: regexPatt,
			negate:  negateRegex,
		}
	}
	c.regexPatterns = patterns
}
