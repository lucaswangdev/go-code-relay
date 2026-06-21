package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var cwd string

var dangerousPatterns = [](struct {
	pattern *regexp.Regexp
	reason  string
}){
	{regexp.MustCompile(`\brm\s+(-\w*)?-r\w*\s+(/|~|\$HOME)`), "recursive delete on home/root"},
	{regexp.MustCompile(`\brm\s+(-\w*)?-rf\s`), "force recursive delete"},
	{regexp.MustCompile(`\bmkfs\b`), "format filesystem"},
	{regexp.MustCompile(`\bdd\s+.*of=/dev/`), "raw disk write"},
	{regexp.MustCompile(`>\s*/dev/sd[a-z]`), "overwrite block device"},
	{regexp.MustCompile(`\bchmod\s+(-R\s+)?777\s+/`), "chmod 777 on root"},
	{regexp.MustCompile(`:\(\)\s*\{.*:\|:.*\}`), "fork bomb"},
	{regexp.MustCompile(`\bcurl\b.*\|\s*(sudo\s+)?bash`), "pipe curl to bash"},
	{regexp.MustCompile(`\bwget\b.*\|\s*(sudo\s+)?bash`), "pipe wget to bash"},
}

type BashTool struct{}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) Description() string {
	return "Execute a shell command. Returns stdout, stderr, and exit code."
}
func (t *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to run",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default 120)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Execute(args map[string]interface{}) (string, error) {
	command, _ := args["command"].(string)
	timeoutSec := 120
	if to, ok := args["timeout"].(float64); ok {
		timeoutSec = int(to)
	}

	if warning := checkDangerous(command); warning != "" {
		return "⚠ Blocked: " + warning + "\nCommand: " + command, nil
	}

	workDir := cwd
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := string(output)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "Error: timed out after " + string(rune(timeoutSec)) + "s", nil
		}
		result += "\n[exit code: " + err.Error() + "]"
	}

	if len(result) > 15000 {
		result = result[:6000] + "\n\n... truncated ...\n\n" + result[len(result)-3000:]
	}

	if strings.Contains(command, "cd ") && cmd.ProcessState != nil && cmd.ProcessState.Success() {
		updateCWD(command, workDir)
	}

	return result, nil
}

func checkDangerous(cmd string) string {
	for _, p := range dangerousPatterns {
		if p.pattern.MatchString(cmd) {
			return p.reason
		}
	}
	return ""
}

func updateCWD(command, currentCwd string) {
	parts := strings.Split(command, "&&")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "cd ") {
			target := strings.TrimPrefix(part, "cd ")
			target = strings.Trim(target, "'\"")
			if target != "" {
				newDir := target
				if !filepath.IsAbs(newDir) {
					newDir = currentCwd + "/" + newDir
				}
				newDir = os.ExpandEnv(newDir)
				if info, err := os.Stat(newDir); err == nil && info.IsDir() {
					cwd = newDir
				}
			}
		}
	}
}

func (t *BashTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        t.Name(),
			"description": t.Description(),
			"parameters":  t.Parameters(),
		},
	}
}
