package file_system

type FileSystemConfig struct {
	ignorePatterns []string
}

func NewFileSystemConfig() (*FileSystemConfig, error) {
	ignorePatterns, err := LoadGitignore(".gitignore")
	if err != nil {
		return nil, err
	}

	return &FileSystemConfig{
		ignorePatterns: ignorePatterns,
	}, nil
}
