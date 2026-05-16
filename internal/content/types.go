package content

import (
	"encoding/json"
	"strings"
)

// FormatGuidance returns the structured content format guidance to be injected into system prompts.
// This is automatically prepended to the user's system prompt to ensure the LLM uses structured formats.
func FormatGuidance() string {
	return `## IMPORTANT: Response Format Instructions

YOU MUST use structured JSON format for the following cases:

1. TABLES - When user asks for data in table format, lists, comparisons, or any tabular data:
   ALWAYS respond with: {"type":"table","content":{"headers":["Name","Value","Status"],"rows":[["Item1","Data1","Active"],["Item2","Data2","Inactive"]]}}
   NEVER use markdown tables or ASCII tables.

2. LISTS - When presenting steps, items, or nested information:
   ALWAYS respond with: {"type":"list","content":{"type":"ordered","items":[{"text":"First step"},{"text":"Second step with details","children":[{"text":"Sub-step"}]}]}}
   For unordered use "type":"unordered" instead.

3. CODE - When showing code examples, commands, or configuration:
   ALWAYS respond with: {"type":"code","content":{"language":"bash","code":"your code here"}}
   NEVER use markdown code blocks. ALWAYS specify the language.

4. LINKS - When providing URLs or resources:
   ALWAYS respond with: {"type":"link","content":{"text":"Link description","url":"https://example.com"}}

5. DEFAULT - For conversational text, questions, and explanations:
   Use plain text normally. You can mix plain text explanations with structured content.

CRITICAL:
- Respond ONLY with valid JSON for structured content (nothing before or after).
- When user asks for "table", "list", "steps", "show me", "display", etc. - use appropriate structured format.
- Do NOT wrap JSON in markdown code blocks or backticks.
- Each response should be either plain text OR pure JSON, not both mixed on same line.`
}

// ContentType represents the type of structured content
type ContentType string

const (
	TypeTable     ContentType = "table"
	TypeList      ContentType = "list"
	TypeLink      ContentType = "link"
	TypeCodeBlock ContentType = "code"
)

// StructuredContent represents a piece of structured content
type StructuredContent struct {
	Type    ContentType `json:"type"`
	Content interface{} `json:"content"`
}

// Table represents a table with headers and rows
type Table struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// List represents a list with optional nesting
type List struct {
	Items []ListItem `json:"items"`
	Type  string     `json:"type"` // "ordered" or "unordered"
}

// ListItem represents a single list item with optional children
type ListItem struct {
	Text     string      `json:"text"`
	Children []ListItem  `json:"children,omitempty"`
}

// Link represents a hyperlink
type Link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// CodeBlock represents a code snippet
type CodeBlock struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// ParseStructuredContent attempts to parse JSON-wrapped structured content
func ParseStructuredContent(data []byte) (*StructuredContent, error) {
	var sc StructuredContent
	err := json.Unmarshal(data, &sc)
	return &sc, err
}

// IsStructuredContent checks if a string looks like structured content JSON
func IsStructuredContent(text string) bool {
	if len(text) < 10 {
		return false
	}
	// Trim whitespace before checking
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	var m map[string]interface{}
	err := json.Unmarshal([]byte(trimmed), &m)
	if err != nil {
		return false
	}
	_, hasType := m["type"]
	_, hasContent := m["content"]
	return hasType && hasContent
}
