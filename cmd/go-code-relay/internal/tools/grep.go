package tools

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true,
	".venv": true, "venv": true, ".tox": true,
	"dist": true, "build": true,
}

type GrepTool struct{}

func (t *GrepTool) Name() string        { return "grep" }
func (t *GrepTool) Description() string {
	return "Search file contents with regex. Returns matching lines with file path and line number."
}
func (t *GrepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory to search (default: cwd)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "Only search files matching this glob (e.g. '*.py')",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Execute(args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	path := "."
	if p, ok := args["path"].(string); ok {
		path = p
	}
	include, _ := args["include"].(string)

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "Invalid regex: " + err.Error(), nil
	}

	base, err := filepath.Abs(path)
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	info, err := os.Stat(base)
	if err != nil {
		return "Error: " + path + " not found", nil
	}

	var files []string
	if info.IsDir() {
		files = t.walkDir(base, include)
	} else {
		files = []string{base}
	}

	var matches []string
	for _, fp := range files {
		content, err := os.ReadFile(fp)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for lineno, line := range lines {
			if regex.MatchString(line) {
				matches = append(matches, fp+":"+strconv.Itoa(lineno+1)+": "+strings.TrimRight(line, "\r"))
				if len(matches) >= 200 {
					matches = append(matches, "... (200 match limit reached)")
					return strings.Join(matches, "\n"), nil
				}
			}
		}
	}

	if len(matches) == 0 {
		return "No matches found.", nil
	}
	return strings.Join(matches, "\n"), nil
}

func (t *GrepTool) walkDir(root, include string) []string {
	var results []string
	count := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		parts := strings.Split(path, string(os.PathSeparator))
		for _, part := range parts {
			if skipDirs[part] {
				return filepath.SkipDir
			}
		}
		if info.IsDir() {
			return nil
		}
		if include != "" {
			matched, _ := filepath.Match(include, filepath.Base(path))
			if !matched {
				return nil
			}
		}
		results = append(results, path)
		count++
		if count >= 5000 {
			return filepath.SkipAll
		}
		return nil
	})
	return results
}

func (t *GrepTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
