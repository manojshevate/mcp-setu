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
   ALWAYS respond ONLY with: {"type":"table","content":{"headers":["Name","Value","Status"],"rows":[["Item1","Data1","Active"],["Item2","Data2","Inactive"]]}}
   - NEVER add preamble text before the JSON
   - NEVER use markdown tables or ASCII tables
   - No explanations, just the JSON object

2. LISTS - When presenting steps, items, or nested information:
   ALWAYS respond ONLY with: {"type":"list","content":{"type":"ordered","items":[{"text":"First step"},{"text":"Second step with details","children":[{"text":"Sub-step"}]}]}}
   For unordered use "type":"unordered" instead.
   - No text before or after, just the JSON

3. CODE - When showing code examples, commands, or configuration:
   ALWAYS respond ONLY with: {"type":"code","content":{"language":"bash","code":"your code here"}}
   - NEVER use markdown code blocks
   - ALWAYS specify the language
   - No preamble or explanation, just the JSON

4. LINKS - When providing URLs or resources:
   ALWAYS respond ONLY with: {"type":"link","content":{"text":"Link description","url":"https://example.com"}}
   - Just the JSON, no additional text

5. DEFAULT - For conversational text, questions, and explanations:
   Use plain text normally. You can mix plain text with multiple structured objects if needed.
   Example: "Here is the analysis:" [plain text explanation] then {"type":"table",...} or {"type":"code",...}

CRITICAL RULES:
- For structured content (table, list, code, link): Respond ONLY with valid JSON.
- NEVER wrap JSON in markdown code blocks or backticks.
- NEVER add explanatory text before structured responses.
- When user explicitly asks for "table", "list", "steps", "code", "link", or similar - ONLY return the JSON object.
- Plain text can include structured objects; structured objects must NOT include plain text within them.`
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

// ParseStructuredContent attempts to parse JSON-wrapped structured content.
// It extracts JSON from mixed content if necessary (e.g., preamble text before JSON).
func ParseStructuredContent(data []byte) (*StructuredContent, error) {
	text := string(data)
	trimmed := strings.TrimSpace(text)

	// Try parsing directly first
	var sc StructuredContent
	err := json.Unmarshal([]byte(trimmed), &sc)
	if err == nil {
		return &sc, nil
	}

	// Look for JSON embedded in the content
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '{' {
			depth := 0
			for j := i; j < len(trimmed); j++ {
				if trimmed[j] == '{' {
					depth++
				} else if trimmed[j] == '}' {
					depth--
					if depth == 0 {
						jsonStr := trimmed[i : j+1]
						var sc2 StructuredContent
						err2 := json.Unmarshal([]byte(jsonStr), &sc2)
						if err2 == nil && sc2.Type != "" && sc2.Content != nil {
							return &sc2, nil
						}
						i = j
						break
					}
				}
			}
		}
	}

	return nil, err
}

// IsStructuredContent checks if a string contains valid structured content JSON.
// It looks for JSON that has both "type" and "content" fields, extracting it
// from mixed content if necessary (e.g., text before or after JSON).
func IsStructuredContent(text string) bool {
	if len(text) < 10 {
		return false
	}

	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return false
	}

	// If the trimmed content starts with {, try parsing it directly
	if trimmed[0] == '{' {
		if isValidStructuredJSON([]byte(trimmed)) {
			return true
		}
	}

	// Look for JSON in the content by finding { and } boundaries
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '{' {
			// Found a potential JSON start, find the matching closing brace
			depth := 0
			for j := i; j < len(trimmed); j++ {
				if trimmed[j] == '{' {
					depth++
				} else if trimmed[j] == '}' {
					depth--
					if depth == 0 {
						// Found complete JSON object
						jsonStr := trimmed[i : j+1]
						if isValidStructuredJSON([]byte(jsonStr)) {
							return true
						}
						i = j // Skip past this JSON
						break
					}
				}
			}
		}
	}

	return false
}

// isValidStructuredJSON checks if bytes are valid structured content JSON
func isValidStructuredJSON(data []byte) bool {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return false
	}
	_, hasType := m["type"]
	_, hasContent := m["content"]
	return hasType && hasContent
}
