# Structured Content System

The mcp-setu CLI supports formatted structured content (tables, lists, links, code blocks) that renders beautifully in the terminal. This document explains how to integrate structured content support.

## Overview

When the LLM needs to present structured data, it can wrap responses in JSON format. The client automatically detects, parses, and renders these in a user-friendly format.

## Supported Formats

See `internal/content/FORMATS.md` for complete format specifications.

- **Table**: Tabular data with headers and rows
- **List**: Ordered or unordered lists with optional nesting
- **Link**: URLs with descriptive text
- **Code Block**: Code snippets with language hint

## Integration with System Prompt

To encourage the LLM to use structured content, add this guidance to the system prompt:

### Example System Prompt Additions

Add this to the Ollama configuration system prompt:

```
## Structured Content

When presenting data that would benefit from formatting, use JSON-wrapped structured content.

For tables (lists with multiple attributes):
{
  "type": "table",
  "content": {
    "headers": ["Column 1", "Column 2"],
    "rows": [["value1", "value2"]]
  }
}

For lists (multi-item information):
{
  "type": "list",
  "content": {
    "type": "unordered",
    "items": [
      {"text": "Item 1", "children": [{"text": "Sub-item"}]},
      {"text": "Item 2"}
    ]
  }
}

For links/resources:
{
  "type": "link",
  "content": {"text": "Link text", "url": "https://example.com"}
}

For code examples:
{
  "type": "code",
  "content": {"language": "bash", "code": "command here"}
}

Guidelines:
- Use plain text for conversational responses
- Use structured content for data, lists, examples, or documentation
- Be descriptive with table headers and list items
- Always provide language hints in code blocks
```

## Client-Side Implementation

### Detecting Structured Content

The bridge provides helper methods:

```go
// Check if response contains structured content
if bridge.IsStructuredContent(response) {
    sc, err := bridge.ParseStructuredContent(response)
    if err == nil {
        printer.PrintStructuredContent(sc)
        return
    }
}

// Fall back to plain text
printer.PrintResponseChunk(response)
```

### Rendering in UI/TUI

The printer implementation should handle the `PrintStructuredContent` method:

```go
func (p *Printer) PrintStructuredContent(sc *content.StructuredContent) {
    renderer := content.NewRenderer(terminalWidth)
    formatted := renderer.Render(sc)
    fmt.Print(formatted)
}
```

## Examples

### Table Example

User asks: "List all local Ollama models with their sizes"

LLM responds:
```json
{
  "type": "table",
  "content": {
    "headers": ["Model", "Size", "Quantization"],
    "rows": [
      ["llama2:7b", "3.8GB", "Q4"],
      ["mistral:7b", "4.2GB", "Q4"],
      ["neural-chat:7b", "4.0GB", "Q4"]
    ]
  }
}
```

Renders as:
```
┌────────────────┬────────┬──────────────┐
│ Model          │ Size   │ Quantization │
├────────────────┼────────┼──────────────┤
│ llama2:7b      │ 3.8GB  │ Q4           │
│ mistral:7b     │ 4.2GB  │ Q4           │
│ neural-chat:7b │ 4.0GB  │ Q4           │
└────────────────┴────────┴──────────────┘
```

### List Example

User asks: "What are the steps to use a custom model?"

LLM responds:
```json
{
  "type": "list",
  "content": {
    "type": "ordered",
    "items": [
      {"text": "Create mcp.json in your project"},
      {
        "text": "Configure your model",
        "children": [
          {"text": "Set baseURL for Ollama"},
          {"text": "Specify model name"}
        ]
      },
      {"text": "Start mcp-setu chat"},
      {"text": "Type your message and press Enter"}
    ]
  }
}
```

Renders as:
```
1. Create mcp.json in your project
2. Configure your model
   2.1. Set baseURL for Ollama
   2.2. Specify model name
3. Start mcp-setu chat
4. Type your message and press Enter
```

### Code Block Example

LLM responds:
```json
{
  "type": "code",
  "content": {
    "language": "bash",
    "code": "mcp-setu chat --model llama2:7b --debug"
  }
}
```

Renders as:
```
┌─ bash ──────────────────────────────────┐
│ mcp-setu chat --model llama2:7b --debug │
└─────────────────────────────────────────┘
```

## Testing Structured Content

To test structured content rendering:

```go
import "github.com/manojshevate/mcp-setu/internal/content"

renderer := content.NewRenderer(80)

// Test table
table := &content.StructuredContent{
    Type: content.TypeTable,
    Content: map[string]interface{}{
        "headers": []string{"Name", "Value"},
        "rows": [][]string{{"Item1", "Value1"}},
    },
}
fmt.Println(renderer.Render(table))
```

## Future Enhancements

- Color highlighting for syntax-highlighted code blocks
- Custom styling for different content types
- Support for more complex nested structures
- Client-side content caching
- Theme support for different terminal color schemes
