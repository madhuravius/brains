package core

import (
	"context"

	awsConfig "github.com/madhuravius/brains/internal/aws"
	brainsConfig "github.com/madhuravius/brains/internal/config"
	"github.com/madhuravius/brains/internal/tools/browser"
	"github.com/madhuravius/brains/internal/tools/file_system"
)

type CoreImpl interface {
	AskFlow(ctx context.Context, llmRequest *LLMRequest) error
	CodeFlow(ctx context.Context, llmRequest *LLMRequest) error
	ValidateBedrockConfiguration(modelID string) bool

	SetLogger(l brainsConfig.SimpleLogger)
	GetAWSConfig() awsConfig.AWSImpl
	SetAWSConfig(a awsConfig.AWSImpl)
}

type toolsConfig struct {
	browserToolConfig browser.BrowserImpl
	fsToolConfig      file_system.FileSystemImpl
}

type CoreConfig struct {
	awsImpl      awsConfig.AWSImpl
	brainsConfig brainsConfig.BrainsConfigImpl
	logger       brainsConfig.SimpleLogger
	toolsConfig  *toolsConfig
}

type LLMRequest struct {
	Glob                string
	ModelID             string
	PersonaInstructions string
	Prompt              string
}
type Researchable interface {
	SetFileMapData(filePath, filePathData string)
	SetResearchData(url, data string)
}
type RepoMappable interface {
	SetRepoMapContext(repoMap string)
}
type FileMapData map[string]string
type ResearchData map[string]string

type CommonData struct {
	ResearchData
	FileMapData
	RepoMapContext string
}

type AskData struct {
	*CommonData
}
type askDataDAGFunction func(inputs map[string]string) (string, error)

type CodeData struct {
	*CommonData
	CodeModelResponse *CodeModelResponse
}
type codeDataDAGFunction func(inputs map[string]string) (string, error)

type InitialContextSettable interface {
	generateInitialContextRun() string
}

type Hydratable interface {
	IsHydrated() bool
}

type HasParameters[T any] interface {
	GetParameters() T
}

type CodeUpdate struct {
	Path    string `json:"path"`
	OldCode string `json:"old_code"`
	NewCode string `json:"new_code"`
}

type AddCodeFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type RemoveCodeFile struct {
	Path string `json:"path"`
}

type CodeModelResponse struct {
	MarkdownSummary string           `json:"markdown_summary"`
	CodeUpdates     []CodeUpdate     `json:"code_updates"`
	AddCodeFiles    []AddCodeFile    `json:"add_code_files,omitempty"`
	RemoveCodeFiles []RemoveCodeFile `json:"remove_code_files,omitempty"`
}

type CodeModelResponseWithParameters struct {
	Name       string            `json:"name"`
	Parameters CodeModelResponse `json:"parameters"`
}

type ResearchActions struct {
	UrlsRecommended []string `json:"urls_recommended"`
	FilesRequested  []string `json:"files_requested"`
}

type ResearchModelResponse struct {
	MarkdownSummary string          `json:"markdown_summary"`
	ResearchActions ResearchActions `json:"research_actions"`
}

type ResearchModelResponseWithParameters struct {
	Name       string                `json:"name"`
	Parameters ResearchModelResponse `json:"parameters"`
}
