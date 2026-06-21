package tools

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var changedFiles = make(map[string]bool)

type WriteFileTool struct{}

func (t *WriteFileTool) Name() string        { return "write_file" }
func (t *WriteFileTool) Description() string { return "Create a new file or completely overwrite an existing one." }
func (t *WriteFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path for the file",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Full file content to write",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *WriteFileTool) Execute(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)

	dir := filepath.Dir(filePath)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	changedFiles[filePath] = true

	lines := strings.Count(content, "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		lines++
	}

	return "Wrote " + strconv.Itoa(lines) + " lines to " + filePath, nil
}

func (t *WriteFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
