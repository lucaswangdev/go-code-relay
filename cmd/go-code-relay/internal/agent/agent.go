package agent

import (
	"sync"

	"github.com/corecoder/go-code/internal/context"
	"github.com/corecoder/go-code/internal/llm"
	"github.com/corecoder/go-code/internal/prompt"
	"github.com/corecoder/go-code/internal/tools"
)

type Agent struct {
	LLM          *llm.LLM
	Tools        []tools.Tool
	Messages     []map[string]interface{}
	Context      *context.ContextManager
	MaxRounds    int
	SystemPrompt string
}

func NewAgent(llmClient *llm.LLM, maxContextTokens int) *Agent {
	tools.RegisterTools()
	toolNames := make([]string, len(tools.AllTools))
	for i, t := range tools.AllTools {
		toolNames[i] = t.Name()
	}

	return &Agent{
		LLM:          llmClient,
		Tools:        tools.AllTools,
		Messages:     make([]map[string]interface{}, 0),
		Context:      context.NewContextManager(maxContextTokens),
		MaxRounds:    50,
		SystemPrompt: prompt.SystemPrompt(toolNames),
	}
}

func (a *Agent) Chat(userInput string, onToken func(string), onTool func(string, map[string]interface{})) string {
	a.Messages = append(a.Messages, map[string]interface{}{
		"role":    "user",
		"content": userInput,
	})

	a.Context.MaybeCompress(a.Messages, a.LLM)

	for i := 0; i < a.MaxRounds; i++ {
		toolSchemas := make([]map[string]interface{}, len(a.Tools))
		for i, t := range a.Tools {
			toolSchemas[i] = t.Schema()
		}

		fullMessages := append([]map[string]interface{}{
			{"role": "system", "content": a.SystemPrompt},
		}, a.Messages...)

		resp, err := a.LLM.Chat(fullMessages, toolSchemas, onToken)
		if err != nil {
			return "Error: " + err.Error()
		}

		if len(resp.ToolCalls) == 0 {
			a.Messages = append(a.Messages, map[string]interface{}{
				"role":    "assistant",
				"content": resp.Content,
			})
			return resp.Content
		}

		a.Messages = append(a.Messages, map[string]interface{}{
			"role":       "assistant",
			"content":    resp.Content,
			"tool_calls": resp.ToolCalls,
		})

		if len(resp.ToolCalls) == 1 {
			tc := resp.ToolCalls[0]
			if onTool != nil {
				onTool(tc.Name, tc.Arguments)
			}
			result := a.execTool(tc)
			a.Messages = append(a.Messages, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": tc.ID,
				"content":      result,
			})
		} else {
			results := a.execToolsParallel(resp.ToolCalls, onTool)
			for j, tc := range resp.ToolCalls {
				a.Messages = append(a.Messages, map[string]interface{}{
					"role":         "tool",
					"tool_call_id": tc.ID,
					"content":      results[j],
				})
			}
		}

		a.Context.MaybeCompress(a.Messages, a.LLM)
	}

	return "(reached maximum tool-call rounds)"
}

func (a *Agent) execTool(tc llm.ToolCall) string {
	tool := tools.GetTool(tc.Name)
	if tool == nil {
		return "Error: unknown tool '" + tc.Name + "'"
	}
	result, err := tool.Execute(tc.Arguments)
	if err != nil {
		return "Error executing " + tc.Name + ": " + err.Error()
	}
	return result
}

func (a *Agent) execToolsParallel(toolCalls []llm.ToolCall, onTool func(string, map[string]interface{})) []string {
	results := make([]string, len(toolCalls))
	var wg sync.WaitGroup

	for i, tc := range toolCalls {
		wg.Add(1)
		go func(idx int, call llm.ToolCall) {
			defer wg.Done()
			if onTool != nil {
				onTool(call.Name, call.Arguments)
			}
			results[idx] = a.execTool(call)
		}(i, tc)
	}

	wg.Wait()
	return results
}

func (a *Agent) Reset() {
	a.Messages = make([]map[string]interface{}, 0)
}
