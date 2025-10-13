package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
)

func LoadGitignore(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			pterm.Warning.Println(".gitignore not found, continuing")
			return []string{}, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// getPatternFromLine - this function is heavily adopted from https://github.com/sabhiram/go-gitignore/blob/master/ignore.go
// it helps determine what the actual regex to set up gitignore validity is
func getPatternFromLine(line string) (*regexp.Regexp, bool) {
	line = strings.TrimRight(line, "\r")

	if strings.HasPrefix(line, `#`) {
		return nil, false
	}

	line = strings.Trim(line, " ")

	if line == "" {
		return nil, false
	}

	negatePattern := false
	if line[0] == '!' {
		negatePattern = true
		line = line[1:]
	}

	if regexp.MustCompile(`^(\#|\!)`).MatchString(line) {
		line = line[1:]
	}

	if regexp.MustCompile(`([^\/+])/.*\*\.`).MatchString(line) && line[0] != '/' {
		line = "/" + line
	}

	line = regexp.MustCompile(`\.`).ReplaceAllString(line, `\.`)

	magicStar := "#$~"

	if strings.HasPrefix(line, "/**/") {
		line = line[1:]
	}
	line = regexp.MustCompile(`/\*\*/`).ReplaceAllString(line, `(/|/.+/)`)
	line = regexp.MustCompile(`\*\*/`).ReplaceAllString(line, `(|.`+magicStar+`/)`)
	line = regexp.MustCompile(`/\*\*`).ReplaceAllString(line, `(|/.`+magicStar+`)`)

	line = regexp.MustCompile(`\\\*`).ReplaceAllString(line, `\`+magicStar)
	line = regexp.MustCompile(`\*`).ReplaceAllString(line, `([^/]*)`)

	line = strings.ReplaceAll(line, "?", `\?`)

	line = strings.ReplaceAll(line, magicStar, "*")

	var expr = ""
	if strings.HasSuffix(line, "/") {
		expr = line + "(|.*)$"
	} else {
		expr = line + "(|/.*)$"
	}
	if strings.HasPrefix(expr, "/") {
		expr = "^(|/)" + expr[1:]
	} else {
		expr = "^(|.*/)" + expr
	}
	pattern, _ := regexp.Compile(expr)

	return pattern, negatePattern
}

func (c *CommonToolsConfig) IsIgnored(path string) bool {
	if strings.Contains("/"+path+"/", "/.git/") || path == ".git" || strings.HasSuffix(path, "/.git") {
		return true
	}

	if base := filepath.Base(path); base != "" {
		if _, found := defaultIgnoreNames[base]; found {
			return true
		}
	}

	// Replace OS-specific path separator.
	matchesPath := false
	for _, ip := range c.regexPatterns {
		if ip.pattern.MatchString(path) {
			// If this is a regular target (not negated with a gitignore
			// exclude "!" etc)
			if !ip.negate {
				matchesPath = true
			} else if matchesPath {
				// Negated pattern, and matchesPath is already set
				matchesPath = false
			}
		}
	}
	return matchesPath
}
