package tools

var AllTools []Tool

func RegisterTools() {
	AllTools = []Tool{
		&BashTool{},
		&ReadFileTool{},
		&WriteFileTool{},
		&EditFileTool{},
		&GlobTool{},
		&GrepTool{},
		&AgentTool{},
	}
}

func GetTool(name string) Tool {
	for _, t := range AllTools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}
