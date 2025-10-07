package browser

import (
	"context"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-shiori/go-readability"
)

func FetchWebContext(ctx context.Context, url string) (string, error) {
	rodUrl := launcher.New().
		Headless(true).
		Set("no-sandbox").
		Set("disable-setuid-sandbox").
		MustLaunch()

	browser := rod.New().ControlURL(rodUrl).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(url)
	page.MustWaitLoad()

	el, err := page.Element("html")
	if err != nil {
		return "", fmt.Errorf("failed to find html element: %w", err)
	}
	html, err := el.HTML()
	if err != nil {
		return "", fmt.Errorf("failed to get html: %w", err)
	}

	u, err := neturl.Parse(url)
	if err != nil {
		return "", fmt.Errorf("invalid page URL: %w", err)
	}

	article, err := readability.FromReader(strings.NewReader(html), u)
	if err != nil {
		return "", fmt.Errorf("readability extraction failed: %w", err)
	}

	// strip all the extra content
	text := article.TextContent
	reNewline := regexp.MustCompile(`\n|\r`)
	cleanedText := reNewline.ReplaceAllString(text, " ")

	reMultipleSpaces := regexp.MustCompile(`\s+`)
	cleanedText = reMultipleSpaces.ReplaceAllString(cleanedText, " ")

	return cleanedText, nil
}
