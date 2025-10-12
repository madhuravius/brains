package tools

func NewCommonToolsConfig(ignorePatterns []string) CommonToolsImpl {
	return &CommonToolsConfig{
		ignorePatterns: ignorePatterns,
	}
}
