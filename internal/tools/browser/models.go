package browser

import (
	"context"

	"github.com/go-rod/rod/lib/launcher"
)

type BrowserConfig struct {
	launcher *launcher.Launcher
	rodUrl   string
}

type BrowserImpl interface {
	FetchWebContext(ctx context.Context, url string) (string, error)
}
