package tools

func NewCommonToolsConfig(ignorePatterns []string) CommonToolsImpl {
	return &CommonToolsConfig{
		ignorePatterns: ignorePatterns,
	}
}

func (c *CommonToolsConfig) SetIgnorePatterns(ignorePatterns []string) {
	c.ignorePatterns = ignorePatterns
}
