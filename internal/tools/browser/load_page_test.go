package browser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"

	"brains/internal/tools/browser"
)

func init() {
	launcher.NewBrowser().MustGet()
	launcher.New().NoSandbox(true).MustLaunch()
}

func TestFetchWebContextSuccess(t *testing.T) {
	os.Setenv("ROD_ARGS", "--no-sandbox --disable-setuid-sandbox")
	html := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
    <p>Hello World from Test</p>
</body>
</html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	txt, err := browser.FetchWebContext(context.Background(), srv.URL)
	assert.NoError(t, err)
	assert.Contains(t, txt, "Hello World from Test")
}

func TestFetchWebContextInvalidURL(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = browser.FetchWebContext(context.Background(), "http://[::1]:invalid")
	})
}
