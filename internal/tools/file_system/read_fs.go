package file_system

import (
	"encoding/json"
	"fmt"
	"os"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/pterm/pterm"
)

func (f *FileSystemConfig) GetFileContents(path string) (string, error) {
	if f.commonTools.IsIgnored(path) {
		return "", nil
	}

	info, err := os.Stat(path)
	if err != nil {
		pterm.Warning.Printfln("failed to stat %s: %v", path, err)
		return "", err
	}
	if info.IsDir() {
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		pterm.Warning.Printfln("failed to read %s: %v", path, err)
		return "", err
	}

	return string(data), nil
}

func (f *FileSystemConfig) SetContextFromGlob(pattern string) (string, error) {
	files, err := doublestar.Glob(os.DirFS("."), pattern)
	if err != nil {
		return "", fmt.Errorf("failed to expand glob: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files matched pattern %s", pattern)
	}

	contents := make(map[string]string)

	for _, fpath := range files {
		data, err := f.GetFileContents(fpath)
		if err != nil {
			continue
		}
		if data == "" {
			continue
		}
		pterm.Debug.Printfln("added file to context: %s", fpath)
		contents[fpath] = data
	}

	contentData, err := json.Marshal(contents)
	if err != nil {
		pterm.Error.Printfln("failed to marshal file json map: %v", err)
		return "", err
	}

	return string(contentData), nil
}
