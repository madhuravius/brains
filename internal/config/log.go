package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

func (b *BrainsConfig) InitLogger(enabled bool) error {
	b.logger.enabled = enabled
	if !enabled {
		return nil
	}
	logDir := filepath.Dir(LogPath)
	err := os.MkdirAll(logDir, 0o750)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	b.logger.file = f
	return nil
}

func (b *BrainsConfig) LogMessage(msg string) {
	if !b.logger.enabled || b.logger.file == nil {
		return
	}
	b.logger.mu.Lock()
	_, _ = b.logger.file.WriteString(msg + "\n")
	b.logger.mu.Unlock()
}

func (b *BrainsConfig) GetLogContext() string {
	data, err := os.ReadFile(LogPath)
	if err != nil {
		return ""
	}

	newline := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		newline = "\r\n"
	}

	re := regexp.MustCompile(`(?m)(?:^[ \t]*\r?\n){3,}`)
	out := re.ReplaceAll(data, []byte(newline+newline))

	return string(out)
}

func (b *BrainsConfig) PrintLogs() {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
		glamour.WithColorProfile(termenv.ANSI256),
	)

	logCtx := b.GetLogContext()

	logCtx = strings.ReplaceAll(logCtx, "[REQUEST]", "[❓ REQUEST]")
	logCtx = strings.ReplaceAll(logCtx, "[RESPONSE]", "[⚡ RESPONSE]")
	logCtx = strings.ReplaceAll(logCtx, "[RESPONSE FOR CODE]", "[🧠 RESPONSE FOR CODE]")
	logCtx = strings.ReplaceAll(logCtx, "[RESPONSE FOR RESEARCH]", "[🔍 RESPONSE FOR RESEARCH]")

	rendered, _ := r.Render(logCtx)
	fmt.Println(rendered)
}

func (b *BrainsConfig) Reset() error {
	if !b.logger.enabled {
		return nil
	}
	b.logger.mu.Lock()
	defer b.logger.mu.Unlock()

	if b.logger.file != nil {
		_ = b.logger.file.Close()
		_ = os.Remove(LogPath)
	}
	f, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	b.logger.file = f
	return nil
}
