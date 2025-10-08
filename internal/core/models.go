package core

import (
	awsConfig "brains/internal/aws"
	brainsConfig "brains/internal/config"
	"brains/internal/tools/file_system"
)

type toolsConfig struct {
	fsToolConfig *file_system.FileSystemConfig
}

type CoreConfig struct {
	toolsConfig *toolsConfig
	awsConfig   *awsConfig.AWSConfig
	logger      brainsConfig.SimpleLogger
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
}

type ResearchModelResponse struct {
	ResearchActions ResearchActions `json:"research_actions"`
}

type ResearchModelResponseWithParameters struct {
	Name       string                `json:"name"`
	Parameters ResearchModelResponse `json:"parameters"`
}
