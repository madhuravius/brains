package file_system

import "github.com/madhuravius/brains/internal/tools"

func NewFileSystemConfig() (FileSystemImpl, error) {
	ignorePatterns, err := tools.LoadGitignore(".gitignore")
	if err != nil {
		return nil, err
	}

	return &FileSystemConfig{
		commonTools: tools.NewCommonToolsConfig(ignorePatterns),
	}, nil
}
