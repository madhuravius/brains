package file_system

import "github.com/madhuravius/brains/internal/tools"

type FileSystemConfig struct {
	commonTools tools.CommonToolsImpl
}

func NewFileSystemConfig() (*FileSystemConfig, error) {
	ignorePatterns, err := tools.LoadGitignore(".gitignore")
	if err != nil {
		return nil, err
	}

	return &FileSystemConfig{
		commonTools: tools.NewCommonToolsConfig(ignorePatterns),
	}, nil
}
