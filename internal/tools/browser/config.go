package browser

import (
	"os"

	"github.com/go-rod/rod/lib/launcher"
)

func newBrowser() (*launcher.Launcher, string) {
	l := launcher.New().Headless(true)

	if os.Getenv("DISABLE_ROD_SANDBOX") == "true" {
		l.Set("no-sandbox").
			Set("disable-setuid-sandbox").
			Set("disable-dev-shm-usage").
			Set("disable-gpu")
	}

	rodUrl := l.MustLaunch()
	return l, rodUrl
}

func NewBrowserConfig() (BrowserImpl, error) {
	l, rodUrl := newBrowser()
	return &BrowserConfig{
		launcher: l,
		rodUrl:   rodUrl,
	}, nil
}
