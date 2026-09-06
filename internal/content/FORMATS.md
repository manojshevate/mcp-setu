# Structured Content Formats

When presenting structured data to users, wrap your response in JSON format as shown below. The client will automatically render these in a user-friendly CLI format.

## Format Overview

All structured content follows this envelope:
```json
{
  "type": "table|list|link|code",
  "content": { ... }
}
```

---

## Table Format

Use for tabular data with headers and rows.

```json
{
  "type": "table",
  "content": {
    "headers": ["Column 1", "Column 2", "Column 3"],
    "rows": [
      ["Row 1 Col 1", "Row 1 Col 2", "Row 1 Col 3"],
      ["Row 2 Col 1", "Row 2 Col 2", "Row 2 Col 3"]
    ]
  }
}
```

**When to use:** Lists of items with multiple attributes, comparison data, results from queries or commands.

**Example use case:**
- List of files with size and date
- Database query results
- Model information with parameters
- Comparison of options

---

## List Format

Use for ordered or unordered lists with optional nesting.

```json
{
  "type": "list",
  "content": {
    "type": "unordered",
    "items": [
      {
        "text": "First item",
        "children": [
          {"text": "Sub-item 1"},
          {"text": "Sub-item 2"}
        ]
      },
      {
        "text": "Second item"
      }
    ]
  }
}
```

**When to use:** Steps, features, options, or items that need hierarchy.

**Type values:** `"ordered"` for numbered lists, `"unordered"` for bullets.

**Example use case:**
- Steps to accomplish a task
- List of features or capabilities
- Nested categories of items
- Instructions with sub-steps

---

## Link Format

Use for presenting URLs with descriptive text.

```json
{
  "type": "link",
  "content": {
    "text": "Link Description",
    "url": "https://example.com"
  }
}
```

**When to use:** References, documentation links, or resources.

**CLI rendering:** Displays as styled text with URL in parentheses.

---

## Code Block Format

Use for code snippets, commands, or configuration examples.

```json
{
  "type": "code",
  "content": {
    "language": "bash",
    "code": "mcp-setu chat --debug --model llama2:7b"
  }
}
```

**Common languages:** `bash`, `python`, `json`, `yaml`, `go`, `sql`

**When to use:** Commands to run, code examples, configuration samples.

---

## Mixed Content Guidance

- **Pure text responses:** Send as plain text (not wrapped in JSON)
- **Structured data:** Use the formats above
- **Explanatory text + structured data:** Send explanation as plain text, then structured content

Example:
```
Here's the data you requested:

{
  "type": "table",
  "content": { ... }
}

To use this data, consider the following steps:

{
  "type": "list",
  "content": { ... }
}
```

---

## Format Detection

The client automatically detects structured content by checking for:
1. Valid JSON
2. Presence of `"type"` field
3. Presence of `"content"` field
4. Valid type value (`table`, `list`, `link`, or `code`)

Invalid or malformed JSON will be displayed as plain text.

---

## Best Practices

1. **Clarity first:** If structured format is unclear, use plain text
2. **One format per response:** Don't mix tables and lists in one JSON object
3. **Reasonable sizes:** Keep tables to ~10 rows and ~5 columns for CLI
4. **Descriptive text:** Always pair structured content with explanation
5. **Language hints:** Always specify language in code blocks
