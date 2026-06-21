package llm

type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

type LLMResponse struct {
	Content          string
	ToolCalls        []ToolCall
	PromptTokens     int
	CompletionTokens int
}
