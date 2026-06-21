package llm

// Pricing per million tokens: (input, output)
var pricing = map[string][2]float64{
	"gpt-5.4":            {2.5, 15},
	"gpt-5.4-mini":       {0.75, 4.5},
	"gpt-5.4-nano":       {0.2, 1.25},
	"o4-mini":            {1.1, 4.4},
	"gpt-4.1":            {2, 8},
	"gpt-4.1-mini":       {0.4, 1.6},
	"gpt-4.1-nano":       {0.1, 0.4},
	"gpt-4o":             {2.5, 10},
	"gpt-4o-mini":        {0.15, 0.6},
	"deepseek-chat":      {0.27, 1.10},
	"deepseek-reasoner":  {0.55, 2.19},
	"claude-opus-4-6":   {5, 25},
	"claude-sonnet-4-6":  {3, 15},
	"claude-haiku-4-5":   {1, 5},
	"qwen3-max":          {0.78, 3.9},
	"qwen3-plus":         {0.26, 0.78},
	"qwen-max":           {0.78, 3.9},
	"kimi-k2.5":          {0.6, 3},
}
