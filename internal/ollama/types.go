package ollama

// ChatRequest represents a request to the Ollama /api/chat endpoint.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Options  *Options  `json:"options,omitempty"`
	Stream   bool      `json:"stream"`
}

// Options represents sampling options for Ollama /api/chat endpoint.
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content,omitempty"`
	Thinking  string     `json:"thinking,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Function  *FunctionCall  `json:"function,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// FunctionCall represents a function call within a tool call (for newer Ollama format).
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Index     int            `json:"index,omitempty"`
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
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// ModelInfo represents metadata about a local Ollama model.
type ModelInfo struct {
	Name string
	Size string
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

// NormalizeToolCall extracts the actual tool name and arguments from the ToolCall,
// handling both the old format (name + arguments directly) and newer format
// (nested in function object).
func (tc *ToolCall) NormalizeToolCall() (name string, args map[string]any) {
	// If using newer format with nested function object
	if tc.Function != nil {
		return tc.Function.Name, tc.Function.Arguments
	}
	// Fall back to older format with direct name and arguments
	return tc.Name, tc.Arguments
}
