package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ehmo/repo-tokens/internal/badge"
	"github.com/ehmo/repo-tokens/internal/counter"
	"github.com/ehmo/repo-tokens/internal/detect"
)

func printDetected(projects []detect.Project, root string, asJSON bool) {
	if asJSON {
		type dp struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Type string `json:"type"`
		}
		out := make([]dp, len(projects))
		for i, p := range projects {
			out[i] = dp{Name: p.Name, Path: p.Path, Type: p.Preset.Name}
		}
		writeJSON(out)
		return
	}

	fmt.Printf("\n  %s — %d project(s) detected\n\n", filepath.Base(root), len(projects))
	for _, p := range projects {
		exts := strings.Join(p.Preset.Extensions, ", ")
		if len(exts) > 50 {
			exts = exts[:47] + "..."
		}
		fmt.Printf("  %-20s %-12s %s\n", p.Path, p.Preset.Name, exts)
	}
	fmt.Println()
}

func printTable(s ScanResult, root string) {
	rootName := filepath.Base(root)
	fmt.Printf("\n  %s", rootName)
	if len(s.Results) > 1 {
		fmt.Printf(" — %d projects", len(s.Results))
	}
	fmt.Printf(" (context: %s)\n\n", badge.FormatTokens(s.ContextWindow))

	if len(s.Results) == 1 {
		r := s.Results[0]
		pct := badge.Percentage(r.Tokens, s.ContextWindow)
		fmt.Printf("  %s (%s)\n", r.Project.Name, r.Project.Preset.Name)
		fmt.Printf("  %s files · %s tokens · %s%% of context\n", badge.FormatNumber(r.Files), badge.FormatTokens(r.Tokens), badge.FormatNumber(pct))
		if r.SkippedBin > 0 {
			fmt.Printf("  (%d binary files skipped)\n", r.SkippedBin)
		}
	} else {
		fmt.Printf("  %-20s %10s %10s %12s   %s\n", "PROJECT", "TYPE", "FILES", "TOKENS", "CTX")
		fmt.Printf("  %s\n", strings.Repeat("─", 70))

		for _, r := range s.Results {
			pct := badge.Percentage(r.Tokens, s.ContextWindow)
			fmt.Printf("  %-20s %10s %10s %12s %5s%%  %s\n",
				r.Project.Path, r.Project.Preset.Name, badge.FormatNumber(r.Files), badge.FormatTokens(r.Tokens), badge.FormatNumber(pct), bar(pct))
		}

		fmt.Printf("  %s\n", strings.Repeat("─", 70))
		pct := badge.Percentage(s.TotalTokens, s.ContextWindow)
		fmt.Printf("  %-20s            %10s %12s %5s%%\n", "Total", badge.FormatNumber(s.TotalFiles), badge.FormatTokens(s.TotalTokens), badge.FormatNumber(pct))
	}
	fmt.Println()
}

func printTopFiles(results []counter.Result, n int) {
	type entry struct {
		project string
		file    counter.FileTokens
	}
	var all []entry
	for _, r := range results {
		for _, f := range r.TopFiles {
			all = append(all, entry{r.Project.Path, f})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].file.Tokens > all[j].file.Tokens })

	if n > len(all) {
		n = len(all)
	}
	fmt.Printf("  Top %d files by token count:\n\n", n)
	for i := range n {
		e := all[i]
		path := e.file.Path
		if len(results) > 1 {
			path = filepath.Join(e.project, e.file.Path)
		}
		fmt.Printf("  %8s  %s\n", badge.FormatTokens(e.file.Tokens), path)
	}
	fmt.Println()
}

type jsonOutputData struct {
	Projects      []jsonProject `json:"projects"`
	TotalFiles    int           `json:"total_files"`
	TotalTokens   int           `json:"total_tokens"`
	ContextWindow int           `json:"context_window"`
	Percentage    int           `json:"percentage"`
}

type jsonProject struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	Files   int    `json:"files"`
	Tokens  int    `json:"tokens"`
	Percent int    `json:"percentage"`
}

func printJSONResults(s ScanResult) {
	out := jsonOutputData{
		TotalFiles:    s.TotalFiles,
		TotalTokens:   s.TotalTokens,
		ContextWindow: s.ContextWindow,
		Percentage:    badge.Percentage(s.TotalTokens, s.ContextWindow),
	}
	for _, r := range s.Results {
		out.Projects = append(out.Projects, jsonProject{
			Name:    r.Project.Name,
			Path:    r.Project.Path,
			Type:    r.Project.Preset.Name,
			Files:   r.Files,
			Tokens:  r.Tokens,
			Percent: badge.Percentage(r.Tokens, s.ContextWindow),
		})
	}
	writeJSON(out)
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func bar(pct int) string {
	filled := min(max(pct*20/100, 0), 20)
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", 20-filled) + "]"
}
