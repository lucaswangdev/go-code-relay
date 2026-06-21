package context

import (
	"strconv"
	"strings"
)

func approxTokens(text string) int {
	return len(text) / 3
}

func EstimateTokens(messages []map[string]interface{}) int {
	total := 0
	for _, m := range messages {
		if content, ok := m["content"].(string); ok {
			total += approxTokens(content)
		}
	}
	return total
}

type ContextManager struct {
	maxTokens    int
	snipAt       int
	summarizeAt  int
	collapseAt   int
}

func NewContextManager(maxTokens int) *ContextManager {
	return &ContextManager{
		maxTokens:   maxTokens,
		snipAt:      int(float64(maxTokens) * 0.5),
		summarizeAt: int(float64(maxTokens) * 0.7),
		collapseAt:  int(float64(maxTokens) * 0.9),
	}
}

func (cm *ContextManager) MaybeCompress(messages []map[string]interface{}, llm interface{}) bool {
	current := EstimateTokens(messages)

	if current > cm.snipAt {
		if cm.snipToolOutputs(messages) {
			return true
		}
	}

	return false
}

func (cm *ContextManager) snipToolOutputs(messages []map[string]interface{}) bool {
	changed := false
	for _, m := range messages {
		if role, _ := m["role"].(string); role != "tool" {
			continue
		}
		content, _ := m["content"].(string)
		if len(content) <= 1500 {
			continue
		}
		lines := strings.Split(content, "\n")
		if len(lines) <= 6 {
			continue
		}
		snipped := strings.Join(lines[:3], "\n") +
			"\n... (" + strconv.Itoa(len(lines)) + " lines, snipped) ...\n" +
			strings.Join(lines[len(lines)-3:], "\n")
		m["content"] = snipped
		changed = true
	}
	return changed
}
