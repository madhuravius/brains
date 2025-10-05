package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

func (b *BrainsConfig) initLogger(enabled bool) error {
	b.logger.enabled = enabled
	if !enabled {
		return nil
	}
	f, err := os.OpenFile(".brains.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
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
	data, err := os.ReadFile(".brains.log")
	if err != nil {
		if len(b.logger.mem) == 0 {
			return ""
		}
		return strings.Join(b.logger.mem, "\n")
	}
	return string(data)
}

func (b *BrainsConfig) PrintLogs() {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
		glamour.WithColorProfile(termenv.ANSI256),
	)

	logCtx := b.GetLogContext()

	logCtx = strings.ReplaceAll(logCtx, "[REQUEST]", "# __[❓ REQUEST]__ \n")
	logCtx = strings.ReplaceAll(logCtx, "[RESPONSE]", "# __[⚡ RESPONSE]__ \n")
	logCtx = strings.ReplaceAll(logCtx, "[RESPONSE FOR CODE]", "# __[🧠 RESPONSE FOR CODE]__ \n")

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
		_ = os.Remove(".brains.log")
	}
	f, err := os.OpenFile(".brains.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	b.logger.file = f
	b.logger.mem = nil
	return nil
}
