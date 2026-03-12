package readme

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UpdateMarkers replaces content between <!-- marker --> and <!-- /marker -->.
// The replacement string is inserted as-is between the markers.
// Returns true if the file was modified.
func UpdateMarkers(path, marker, replacement string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(data)
	esc := regexp.QuoteMeta(marker)
	re, err := regexp.Compile(fmt.Sprintf(`(?s)(<!--\s*%s\s*-->).*?(<!--\s*/%s\s*-->)`, esc, esc))
	if err != nil {
		return false, err
	}

	updated := re.ReplaceAllString(content, "${1}"+replacement+"${2}")
	if updated == content {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(updated), 0o644)
}

// InsertMarkers adds token-count markers to a README if they don't exist.
// Inserts after the first heading line.
func InsertMarkers(path, marker string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(data)
	if strings.Contains(content, fmt.Sprintf("<!-- %s -->", marker)) {
		return false, nil
	}

	tag := fmt.Sprintf("\n<!-- %s --><!-- /%s -->\n", marker, marker)
	lines := strings.Split(content, "\n")

	insertAt := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			insertAt = i + 1
			break
		}
	}

	if insertAt == -1 {
		content = tag + "\n" + content
	} else {
		content = strings.Join(lines[:insertAt], "\n") + "\n" + tag + strings.Join(lines[insertAt:], "\n")
	}
	return true, os.WriteFile(path, []byte(content), 0o644)
}
