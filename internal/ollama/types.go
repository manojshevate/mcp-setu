package ollama

// KnownToolSupportedModels is the list of Ollama model prefixes known to
// support tool calling. Update this list as new models gain support.
// Prefix matching is used so "gemma4:e4b" matches "gemma4".
var KnownToolSupportedModels = []string{
	"gemma4", "gemma3",
	"qwen2.5", "qwen3",
	"llama3.1", "llama3.2", "llama3.3",
	"mistral-nemo",
	"command-r",
	"phi4",
	"deepseek-r1",
}

// ChatRequest represents a request to the Ollama /api/chat endpoint.
type ChatRequest struct {
	Model       string       `json:"model"`
	Messages    []Message    `json:"messages"`
	Tools       []Tool       `json:"tools,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	Stream      bool         `json:"stream"`
}

// Message represents a single message in the conversation.
type Message struct {
	Role      string         `json:"role"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	ToolCall  *ToolCall      `json:"tool_call,omitempty"`
}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Tool represents a tool definition for Ollama.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents the function definition of a tool.
type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ChatResponse represents a response from the Ollama /api/chat endpoint.
type ChatResponse struct {
	Model     string    `json:"model"`
	CreatedAt string    `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
}

// ModelInfo represents metadata about a local Ollama model.
type ModelInfo struct {
	Name           string
	Size           string
	ToolSupported  bool
}

// ModelsResponse represents the response from /api/tags.
type ModelsResponse struct {
	Models []ModelDetail `json:"models"`
}

// ModelDetail represents a single model from /api/tags.
type ModelDetail struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}

// ShowResponse represents the response from /api/show.
type ShowResponse struct {
	License    string `json:"license"`
	ModelFile  string `json:"modelfile"`
	Parameters string `json:"parameters"`
	Template   string `json:"template"`
}
