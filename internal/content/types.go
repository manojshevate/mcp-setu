package content

import (
	"encoding/json"
	"strings"
)

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
