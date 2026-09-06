package content

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer renders structured content for CLI display
type Renderer struct {
	terminalWidth int
}

// NewRenderer creates a new content renderer
func NewRenderer(terminalWidth int) *Renderer {
	if terminalWidth < 40 {
		terminalWidth = 40
	}
	return &Renderer{terminalWidth: terminalWidth}
}

// Render takes structured content and returns formatted CLI output
func (r *Renderer) Render(sc *StructuredContent) string {
	if sc == nil {
		return ""
	}

	switch sc.Type {
	case TypeTable:
		if table, ok := sc.Content.(map[string]interface{}); ok {
			return r.renderTable(table)
		}
		// Log failed type assertion for debugging
		return fmt.Sprintf("(warning: table content type assertion failed)\n")
	case TypeList:
		if list, ok := sc.Content.(map[string]interface{}); ok {
			return r.renderList(list)
		}
		return fmt.Sprintf("(warning: list content type assertion failed)\n")
	case TypeLink:
		if link, ok := sc.Content.(map[string]interface{}); ok {
			return r.renderLink(link)
		}
		return fmt.Sprintf("(warning: link content type assertion failed)\n")
	case TypeCodeBlock:
		if code, ok := sc.Content.(map[string]interface{}); ok {
			return r.renderCodeBlock(code)
		}
		return fmt.Sprintf("(warning: code block content type assertion failed)\n")
	}
	return fmt.Sprintf("(warning: unknown content type: %s)\n", sc.Type)
}

// renderTable renders a table with proper column alignment
func (r *Renderer) renderTable(tableData map[string]interface{}) string {
	headers, ok := tableData["headers"].([]interface{})
	if !ok {
		return ""
	}
	rows, ok := tableData["rows"].([]interface{})
	if !ok {
		return ""
	}

	// Convert to string slices
	headerStrs := make([]string, len(headers))
	for i, h := range headers {
		headerStrs[i] = fmt.Sprintf("%v", h)
	}

	rowStrs := make([][]string, len(rows))
	for i, row := range rows {
		if rowData, ok := row.([]interface{}); ok {
			rowStrs[i] = make([]string, len(rowData))
			for j, cell := range rowData {
				rowStrs[i][j] = fmt.Sprintf("%v", cell)
			}
		}
	}

	return formatTable(headerStrs, rowStrs)
}

// renderList renders a bulleted or numbered list
func (r *Renderer) renderList(listData map[string]interface{}) string {
	items, ok := listData["items"].([]interface{})
	if !ok {
		return ""
	}
	listType, _ := listData["type"].(string)

	var output strings.Builder
	r.renderListItems(&output, items, listType, 0)
	return output.String()
}

// renderListItems recursively renders list items with nesting
func (r *Renderer) renderListItems(output *strings.Builder, items []interface{}, listType string, depth int) {
	indent := strings.Repeat("  ", depth)
	for i, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			text, _ := itemMap["text"].(string)
			children, _ := itemMap["children"].([]interface{})

			bullet := "•"
			if listType == "ordered" {
				bullet = fmt.Sprintf("%d.", i+1)
			}

			output.WriteString(indent + bullet + " " + text + "\n")

			if len(children) > 0 {
				r.renderListItems(output, children, listType, depth+1)
			}
		}
	}
}

// renderLink renders a link
func (r *Renderer) renderLink(linkData map[string]interface{}) string {
	text, _ := linkData["text"].(string)
	url, _ := linkData["url"].(string)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#0087FF")).Underline(true)
	return style.Render(text) + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(fmt.Sprintf("(%s)", url))
}

// renderCodeBlock renders a code block with language highlighting indicator
func (r *Renderer) renderCodeBlock(codeData map[string]interface{}) string {
	language, _ := codeData["language"].(string)
	code, _ := codeData["code"].(string)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6B7280")).
		Padding(0, 1)

	var output strings.Builder
	if language != "" {
		langStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Italic(true)
		output.WriteString(langStyle.Render("# " + language) + "\n")
	}
	output.WriteString(style.Render(code))
	return output.String()
}

// formatTable formats a table with proper column alignment
func formatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths with max of 30 chars per cell
	const maxCellWidth = 30
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = minInt(len(h), maxCellWidth)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				width := minInt(len(cell), maxCellWidth)
				if width > widths[i] {
					widths[i] = width
				}
			}
		}
	}

	// Truncate cells that are longer than max width
	for i := range rows {
		for j := range rows[i] {
			if len(rows[i][j]) > maxCellWidth {
				rows[i][j] = rows[i][j][:maxCellWidth-3] + "…"
			}
		}
	}

	// Build separator
	separator := "├"
	for i, w := range widths {
		separator += strings.Repeat("─", w+2)
		if i < len(widths)-1 {
			separator += "┼"
		}
	}
	separator += "┤\n"

	// Build output
	var output strings.Builder
	topSep := "┌"
	for i, w := range widths {
		topSep += strings.Repeat("─", w+2)
		if i < len(widths)-1 {
			topSep += "┬"
		}
	}
	topSep += "┐\n"
	output.WriteString(topSep)

	// Headers
	headerRow := "│"
	for i, h := range headers {
		headerRow += fmt.Sprintf(" %-*s │", widths[i], h)
	}
	headerRow += "\n"
	output.WriteString(headerRow)
	output.WriteString(separator)

	// Rows
	for _, row := range rows {
		rowStr := "│"
		for i, cell := range row {
			if i < len(widths) {
				rowStr += fmt.Sprintf(" %-*s │", widths[i], cell)
			}
		}
		rowStr += "\n"
		output.WriteString(rowStr)
	}

	// Bottom separator
	bottomSep := "└"
	for i, w := range widths {
		bottomSep += strings.Repeat("─", w+2)
		if i < len(widths)-1 {
			bottomSep += "┴"
		}
	}
	bottomSep += "┘\n"
	output.WriteString(bottomSep)

	return output.String()
}

// RenderRawJSON renders the raw JSON for debugging
func (r *Renderer) RenderRawJSON(data []byte) string {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return ""
	}
	pretty, _ := json.MarshalIndent(obj, "", "  ")
	return string(pretty)
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
