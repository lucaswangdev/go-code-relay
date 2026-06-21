package tools

import (
	"os"
	"strconv"
	"strings"
)

type EditFileTool struct{}

func (t *EditFileTool) Name() string        { return "edit_file" }
func (t *EditFileTool) Description() string {
	return "Edit a file by replacing an exact string match."
}
func (t *EditFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to edit",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "Exact text to find (must be unique in file)",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *EditFileTool) Execute(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	oldString, _ := args["old_string"].(string)
	newString, _ := args["new_string"].(string)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Error: " + filePath + " not found", nil
	}

	text := string(content)
	count := strings.Count(text, oldString)
	if count == 0 {
		preview := text
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return "Error: old_string not found in " + filePath + ".\nFile starts with:\n" + preview, nil
	}
	if count > 1 {
		return "Error: old_string appears " + strconv.Itoa(count) + " times. Include more context.", nil
	}

	newContent := strings.Replace(text, oldString, newString, 1)
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return "Error: " + err.Error(), nil
	}

	changedFiles[filePath] = true

	diff := generateDiff(text, newContent, filePath)
	return "Edited " + filePath + "\n" + diff, nil
}

func generateDiff(old, new, filename string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var builder strings.Builder
	builder.WriteString("--- a/" + filename + "\n")
	builder.WriteString("+++ b/" + filename + "\n")

	minLen := len(oldLines)
	if len(newLines) < minLen {
		minLen = len(newLines)
	}

	for i := 0; i < minLen; i++ {
		if oldLines[i] != newLines[i] {
			builder.WriteString("@@ -" + strconv.Itoa(i) + " +" + strconv.Itoa(i) + " @@\n")
			builder.WriteString("-" + oldLines[i] + "\n")
			builder.WriteString("+" + newLines[i] + "\n")
		}
	}

	if len(newLines) > minLen {
		for i := minLen; i < len(newLines); i++ {
			builder.WriteString("+" + newLines[i] + "\n")
		}
	}

	result := builder.String()
	if len(result) > 3000 {
		result = result[:2500] + "\n... (diff truncated)\n"
	}
	return result
}

func (t *EditFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
