package file_system

import "github.com/madhuravius/brains/internal/tools"

type FileSystemConfig struct {
	commonTools tools.CommonToolsImpl
}

type FileSystemImpl interface {
	CreateFile(filePath, fileContents string) error
	DeleteFile(filePath string) error
	GetFileContents(path string) (string, error)
	GetFileTree(root string) (string, error)
	SetContextFromGlob(pattern string) (string, error)
	UpdateFile(filePath, oldContent, newContent string, interactive bool) (bool, error)
}
