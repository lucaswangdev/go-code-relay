package tools

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type GlobTool struct{}

func (t *GlobTool) Name() string        { return "glob" }
func (t *GlobTool) Description() string { return "Find files matching a glob pattern." }
func (t *GlobTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern, e.g. '**/*.py' or 'src/**/*.ts'",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in (default: cwd)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Execute(args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	path := "."
	if p, ok := args["path"].(string); ok {
		path = p
	}

	base, err := filepath.Abs(path)
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	info, err := os.Stat(base)
	if err != nil {
		return "Error: " + path + " not found", nil
	}
	if !info.IsDir() {
		return "Error: " + path + " is not a directory", nil
	}

	matches, err := filepath.Glob(filepath.Join(base, pattern))
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	sort.Slice(matches, func(i, j int) bool {
		iInfo, _ := os.Stat(matches[i])
		jInfo, _ := os.Stat(matches[j])
		if iInfo == nil || jInfo == nil {
			return false
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	total := len(matches)
	shown := matches
	if total > 100 {
		shown = matches[:100]
	}

	var result strings.Builder
	for _, m := range shown {
		result.WriteString(m + "\n")
	}

	if total > 100 {
		result.WriteString("... (" + strconv.Itoa(total) + " matches, showing first 100)")
	}

	if result.Len() == 0 {
		return "No files matched.", nil
	}
	return result.String(), nil
}

func (t *GlobTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
