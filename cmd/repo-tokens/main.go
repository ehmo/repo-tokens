package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ehmo/repo-tokens/internal/badge"
	"github.com/ehmo/repo-tokens/internal/readme"
)

var version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit()
			return
		case "version":
			fmt.Println("repo-tokens", version)
			return
		case "--action":
			runAction()
			return
		}
	}

	var (
		badgePath    string
		ctxWin       int
		encoding     string
		jsonOut      bool
		detectOnly   bool
		showVersion  bool
		repoURL      string
		includeTests bool
		topN         int
		updReadme    string
		marker       string
	)

	flag.StringVar(&badgePath, "badge", "", "Write SVG badge to this path")
	flag.IntVar(&ctxWin, "context-window", 200_000, "Context window size for percentage")
	flag.StringVar(&encoding, "encoding", "cl100k_base", "Tiktoken encoding name")
	flag.BoolVar(&jsonOut, "json", false, "Output results as JSON")
	flag.BoolVar(&detectOnly, "detect", false, "Only show detected projects, don't count")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.StringVar(&repoURL, "repo-url", "", "URL for badge link")
	flag.BoolVar(&includeTests, "include-tests", false, "Include test files in count")
	flag.IntVar(&topN, "top", 0, "Show top N files by token count")
	flag.StringVar(&updReadme, "update-readme", "", "Update token markers in this README file")
	flag.StringVar(&marker, "marker", "token-count", "HTML comment marker name for README")

	flag.Usage = func() {
		w := os.Stderr
		fmt.Fprintf(w, "repo-tokens %s — Count codebase tokens for LLM context awareness\n\n", version)
		fmt.Fprintf(w, "Usage:\n")
		fmt.Fprintf(w, "  repo-tokens [flags] [path]     Count tokens\n")
		fmt.Fprintf(w, "  repo-tokens init [path]        Set up workflow + badge for a repo\n")
		fmt.Fprintf(w, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintln(w, "  repo-tokens                          # Count tokens in current directory")
		fmt.Fprintln(w, "  repo-tokens ~/my-project              # Count a specific path")
		fmt.Fprintln(w, "  repo-tokens --badge badge.svg         # Generate SVG badge")
		fmt.Fprintln(w, "  repo-tokens --top 10                  # Show largest files")
		fmt.Fprintln(w, "  repo-tokens --json                    # JSON output")
		fmt.Fprintln(w, "  repo-tokens --detect                  # Show detected projects")
		fmt.Fprintln(w, "  repo-tokens --context-window 128000   # GPT-4 context")
		fmt.Fprintln(w, "  repo-tokens init                      # Set up GitHub Action")
	}
	flag.Parse()

	if showVersion {
		fmt.Println("repo-tokens", version)
		return
	}

	target := "."
	if flag.NArg() > 0 {
		target = flag.Arg(0)
	}
	target, err := filepath.Abs(target)
	if err != nil {
		fatal("invalid path: %v", err)
	}

	s := scan(ScanParams{
		Target:        target,
		ContextWindow: ctxWin,
		Encoding:      encoding,
		IncludeTests:  includeTests,
	})

	if detectOnly {
		printDetected(s.Projects, target, jsonOut)
		return
	}

	if jsonOut {
		printJSONResults(s)
	} else {
		printTable(s, target)
		if topN > 0 {
			printTopFiles(s.Results, topN)
		}
	}

	if badgePath != "" {
		if err := badge.Write(badgePath, s.TotalTokens, s.ContextWindow, repoURL); err != nil {
			fatal("badge: %v", err)
		}
		fmt.Printf("\nBadge written to %s (%s, %d%%)\n",
			badgePath, badge.FormatTokens(s.TotalTokens), badge.Percentage(s.TotalTokens, s.ContextWindow))
	}

	if updReadme != "" {
		url := repoURL
		if url == "" {
			url = "https://github.com/ehmo/repo-tokens"
		}
		text := badge.Text(s.TotalTokens, s.ContextWindow)
		if ok, err := readme.UpdateMarkers(updReadme, marker, text, url, badgePath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: readme: %v\n", err)
		} else if ok {
			fmt.Printf("\nREADME updated: %s\n", updReadme)
		}
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
