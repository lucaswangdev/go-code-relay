package tools

import (
	"os"
	"strconv"
	"strings"
)

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string        { return "read_file" }
func (t *ReadFileTool) Description() string { return "Read a file's contents with line numbers." }
func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Start line (1-based). Default 1.",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Max lines to read. Default 2000.",
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *ReadFileTool) Execute(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	offset := 1
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	limit := 2000
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	lines := strings.Split(string(content), "\n")
	total := len(lines)

	start := offset - 1
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > total {
		end = total
	}

	var builder strings.Builder
	for i := start; i < end; i++ {
		builder.WriteString(strconv.Itoa(i+1) + "\t" + lines[i] + "\n")
	}

	if end < total {
		builder.WriteString("... (" + strconv.Itoa(total) + " lines total, showing " +
			strconv.Itoa(start+1) + "-" + strconv.Itoa(end) + ")")
	}

	return builder.String(), nil
}

func (t *ReadFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
