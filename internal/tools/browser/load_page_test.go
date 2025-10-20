package browser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madhuravius/brains/internal/tools/browser"
)

func TestFetchWebContextSuccess(t *testing.T) {
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

	b, err := browser.NewBrowserConfig()
	assert.NoError(t, err)
	txt, err := b.FetchWebContext(context.Background(), srv.URL)
	assert.NoError(t, err)
	assert.Contains(t, txt, "Hello World from Test")
}

func TestFetchWebContextInvalidURL(t *testing.T) {
	assert.Panics(t, func() {
		b, _ := browser.NewBrowserConfig()
		_, _ = b.FetchWebContext(context.Background(), "http://[::1]:invalid")
	})
}
