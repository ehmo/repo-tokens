package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ehmo/repo-tokens/internal/badge"
	"github.com/ehmo/repo-tokens/internal/readme"
)

func runAction() {
	path := env("INPUT_PATH", ".")
	ctxWin := envInt("INPUT_CONTEXT_WINDOW", 200_000)
	encoding := env("INPUT_ENCODING", "cl100k_base")
	badgePath := env("INPUT_BADGE_PATH", ".github/badges/tokens.svg")
	readmePath := env("INPUT_README", "README.md")
	marker := env("INPUT_MARKER", "token-count")
	includeTests := env("INPUT_INCLUDE_TESTS", "false") == "true"

	target, err := filepath.Abs(path)
	if err != nil {
		fatal("invalid path: %v", err)
	}

	s := scan(ScanParams{
		Target:        target,
		ContextWindow: ctxWin,
		Encoding:      encoding,
		IncludeTests:  includeTests,
	})

	text := badge.Text(s.TotalTokens, s.ContextWindow)
	fmt.Printf("Files: %d, Tokens: %d, Badge: %s\n", s.TotalFiles, s.TotalTokens, text)
	for _, r := range s.Results {
		fmt.Printf("  %s (%s): %d files, %s tokens, %d%%\n",
			r.Project.Path, r.Project.Preset.Name, r.Files,
			badge.FormatTokens(r.Tokens), badge.Percentage(r.Tokens, s.ContextWindow))
	}

	if badgePath != "" {
		repoURL := gitHubRepoURL()
		if err := badge.Write(badgePath, s.TotalTokens, s.ContextWindow, repoURL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: badge: %v\n", err)
		} else {
			fmt.Printf("Badge written to %s\n", badgePath)
		}
	}

	if readmePath != "" {
		if _, err := os.Stat(readmePath); err == nil {
			url := gitHubRepoURL()
			if url == "" {
				url = "https://github.com/ehmo/repo-tokens"
			}
			if ok, err := readme.UpdateMarkers(readmePath, marker, text, url, badgePath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: readme: %v\n", err)
			} else if ok {
				fmt.Println("README updated")
			}
		}
	}

	if out := os.Getenv("GITHUB_OUTPUT"); out != "" {
		jd, _ := json.Marshal(jsonOutputData{
			TotalFiles:    s.TotalFiles,
			TotalTokens:   s.TotalTokens,
			ContextWindow: s.ContextWindow,
			Percentage:    badge.Percentage(s.TotalTokens, s.ContextWindow),
		})
		if f, err := os.OpenFile(out, os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			fmt.Fprintf(f, "tokens=%d\n", s.TotalTokens)
			fmt.Fprintf(f, "percentage=%d\n", badge.Percentage(s.TotalTokens, s.ContextWindow))
			fmt.Fprintf(f, "badge=%s\n", text)
			fmt.Fprintf(f, "json=%s\n", jd)
			f.Close()
		}
	}
}

func gitHubRepoURL() string {
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		return "https://github.com/" + repo
	}
	return ""
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
