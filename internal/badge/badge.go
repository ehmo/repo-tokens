package badge

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const defaultLinkURL = "https://github.com/ehmo/repo-tokens"

// Percentage returns the context window fill percentage.
func Percentage(tokens, contextWindow int) int {
	return int(math.Round(float64(tokens) / float64(contextWindow) * 100))
}

// FormatNumber returns an integer formatted with comma separators.
func FormatNumber(n int) string {
	if n < 0 {
		return "-" + FormatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var buf strings.Builder
	r := len(s) % 3
	if r > 0 {
		buf.WriteString(s[:r])
	}
	for i := r; i < len(s); i += 3 {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(s[i : i+3])
	}
	return buf.String()
}

// FormatTokens returns a human-friendly token count string.
func FormatTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 100_000:
		return fmt.Sprintf("%dk", int(math.Round(float64(n)/1000)))
	case n >= 1000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// Text returns formatted badge text like "12.3k tokens · 6% of context window".
func Text(tokens, contextWindow int) string {
	return fmt.Sprintf("%s tokens · %s%% of context window",
		FormatTokens(tokens), FormatNumber(Percentage(tokens, contextWindow)))
}

// Color returns a hex color based on context window fill percentage.
func Color(pct int) string {
	switch {
	case pct < 30:
		return "#4c1"
	case pct < 50:
		return "#97ca00"
	case pct < 70:
		return "#dfb317"
	default:
		return "#e05d44"
	}
}

// SVG generates a shields.io-style badge.
func SVG(tokens, contextWindow int, repoURL string) string {
	formatted := FormatTokens(tokens)
	pct := Percentage(tokens, contextWindow)
	display := fmt.Sprintf("%s · %s%%", formatted, FormatNumber(pct))
	color := Color(pct)
	desc := fmt.Sprintf("%s tokens, %s%% of context window", formatted, FormatNumber(pct))
	url := repoURL
	if url == "" {
		url = defaultLinkURL
	}
	return renderSVG("tokens", display, desc, color, url)
}

// Write generates and writes a badge SVG to a file.
func Write(path string, tokens, contextWindow int, repoURL string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(SVG(tokens, contextWindow, repoURL)), 0o644)
}

func renderSVG(label, value, desc, color, url string) string {
	const cw = 7.0
	lw := int(float64(len(label))*cw) + 10
	vw := int(float64(len(value))*cw) + 10
	tw := lw + vw
	lx, vx := lw/2, lw+vw/2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="20" role="img" aria-label="%s">`, tw, desc)
	fmt.Fprintf(&b, "\n  <title>%s</title>", desc)
	b.WriteString("\n  <linearGradient id=\"s\" x2=\"0\" y2=\"100%\">")
	b.WriteString("\n    <stop offset=\"0\" stop-color=\"#bbb\" stop-opacity=\".1\"/>")
	b.WriteString("\n    <stop offset=\"1\" stop-opacity=\".1\"/>")
	b.WriteString("\n  </linearGradient>")
	fmt.Fprintf(&b, "\n  <clipPath id=\"r\"><rect width=\"%d\" height=\"20\" rx=\"3\" fill=\"#fff\"/></clipPath>", tw)
	fmt.Fprintf(&b, "\n  <a xlink:href=\"%s\">", url)
	b.WriteString("\n    <g clip-path=\"url(#r)\">")
	fmt.Fprintf(&b, "\n      <rect width=\"%d\" height=\"20\" fill=\"#555\"/>", lw)
	fmt.Fprintf(&b, "\n      <rect x=\"%d\" width=\"%d\" height=\"20\" fill=\"%s\"/>", lw, vw, color)
	fmt.Fprintf(&b, "\n      <rect width=\"%d\" height=\"20\" fill=\"url(#s)\"/>", tw)
	b.WriteString("\n      <g fill=\"#fff\" text-anchor=\"middle\" font-family=\"Verdana,Geneva,DejaVu Sans,sans-serif\" font-size=\"11\">")
	fmt.Fprintf(&b, "\n        <text aria-hidden=\"true\" x=\"%d\" y=\"15\" fill=\"#010101\" fill-opacity=\".3\">%s</text>", lx, label)
	fmt.Fprintf(&b, "\n        <text x=\"%d\" y=\"14\">%s</text>", lx, label)
	fmt.Fprintf(&b, "\n        <text aria-hidden=\"true\" x=\"%d\" y=\"15\" fill=\"#010101\" fill-opacity=\".3\">%s</text>", vx, value)
	fmt.Fprintf(&b, "\n        <text x=\"%d\" y=\"14\">%s</text>", vx, value)
	b.WriteString("\n      </g>")
	b.WriteString("\n    </g>")
	b.WriteString("\n  </a>")
	b.WriteString("\n</svg>")
	return b.String()
}
