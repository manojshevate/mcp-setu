package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/manojshevate/mcpgo/internal/mcp"
)

// Printer handles all terminal output formatting.
type Printer struct {
	verbose bool
}

// NewPrinter creates a new Printer.
func NewPrinter(verbose bool) *Printer {
	return &Printer{verbose: verbose}
}

// Color palette.
var (
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorError     = lipgloss.Color("#EF4444")
	colorMuted     = lipgloss.Color("#6B7280")
	colorHighlight = lipgloss.Color("#F3F4F6")
)

// PrintBanner prints the startup banner with info table.
func (p *Printer) PrintBanner(model, configPath string, serverCount, toolCount int) {
	// Title banner
	titleStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		MarginBottom(1)

	// Status section with better visual hierarchy
	statusStyle := lipgloss.NewStyle().Padding(0, 2)
	status := fmt.Sprintf(
		"Model       %s\nConfig      %s\nServers     %d connected\nTools       %d available",
		model, configPath, serverCount, toolCount,
	)

	// Build visual banner with dynamic divider width
	dividerStyle := lipgloss.NewStyle().Foreground(colorMuted)
	title := "mcpgo — MCP Bridge for Ollama"
	dividerWidth := lipgloss.Width(title)
	divider := strings.Repeat("━", dividerWidth)

	fmt.Fprintf(os.Stdout, "%s\n", titleStyle.Render(title))
	fmt.Fprintf(os.Stdout, "%s\n\n", dividerStyle.Render(divider))
	fmt.Fprintf(os.Stdout, "%s\n\n", statusStyle.Render(status))
}

// ServerInfo holds info about a server.
type ServerInfo struct {
	Name  string
	Tools int
}

// PrintServerTable prints a table of connected MCP servers.
func (p *Printer) PrintServerTable(servers []ServerInfo) {
	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)

	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("┌────────────────────┬──────────┬─────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │ %s │\n",
		headerStyle.Render("Server              "),
		headerStyle.Render("Status    "),
		headerStyle.Render("Tools          "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├────────────────────┼──────────┼─────────────────┤"))

	for _, s := range servers {
		checkmark := lipgloss.NewStyle().Foreground(colorSuccess).Render("✓")
		toolStr := fmt.Sprintf("%d tools", s.Tools)
		fmt.Fprintf(os.Stdout, "│ %-18s │ %s ready │ %-15s │\n", s.Name, checkmark, toolStr)
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("└────────────────────┴──────────┴─────────────────┘"))
}

// ToolInfo holds info about a tool.
type ToolInfo struct {
	Name        string
	Server      string
	Description string
}

// PrintToolsTable prints a table of all available tools.
func (p *Printer) PrintToolsTable(tools []ToolInfo) {
	if len(tools) == 0 {
		fmt.Fprintf(os.Stdout, "No tools available\n\n")
		return
	}

	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)

	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("┌─────────────────────┬────────────────┬──────────────────────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │ %s │\n",
		headerStyle.Render("Tool                "),
		headerStyle.Render("Server         "),
		headerStyle.Render("Description          "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├─────────────────────┼────────────────┼──────────────────────────────────┤"))

	for _, t := range tools {
		desc := t.Description
		if len(desc) > 32 {
			desc = desc[:29] + "..."
		}
		fmt.Fprintf(os.Stdout, "│ %-19s │ %-14s │ %-32s │\n", t.Name, t.Server, desc)
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("└─────────────────────┴────────────────┴──────────────────────────────────┘"))
}

// ModelInfo holds info about a model.
type ModelInfo struct {
	Name          string
	Size          string
	ToolSupported bool
}

// PrintModelsTable prints a table of available Ollama models.
func (p *Printer) PrintModelsTable(models []ModelInfo) {
	if len(models) == 0 {
		fmt.Fprintf(os.Stdout, "No models found locally\n\n")
		return
	}

	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)

	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("┌─────────────────────┬──────────┬──────────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │ %s │\n",
		headerStyle.Render("Model               "),
		headerStyle.Render("Size     "),
		headerStyle.Render("Tool Support        "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├─────────────────────┼──────────┼──────────────────────┤"))

	for _, m := range models {
		toolSupport := "✗ no tool calling"
		if m.ToolSupported {
			toolSupport = lipgloss.NewStyle().Foreground(colorSuccess).Render("✓ supported")
		} else {
			toolSupport = lipgloss.NewStyle().Foreground(colorError).Render(toolSupport)
		}
		fmt.Fprintf(os.Stdout, "│ %-19s │ %-8s │ %-20s │\n", m.Name, m.Size, toolSupport)
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("└─────────────────────┴──────────┴──────────────────────┘"))
}

// PrintUserPrompt prints the user input prompt.
func (p *Printer) PrintUserPrompt() {
	promptStyle := lipgloss.NewStyle().Foreground(colorPrimary)
	fmt.Fprint(os.Stdout, promptStyle.Render("❯ "))
}

// PrintAssistantResponse prints the assistant's response.
func (p *Printer) PrintAssistantResponse(content string) {
	labelStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorPrimary)

	fmt.Fprintf(os.Stdout, "\n%s\n", borderStyle.Render("┃"))
	fmt.Fprintf(os.Stdout, "%s %s\n", labelStyle.Render("assistant"), content)
	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("┃"))
}

// PrintLLMProcessing prints that the LLM is processing a message in verbose mode.
func (p *Printer) PrintLLMProcessing(iterationNum int) {
	if !p.verbose {
		return
	}
	if iterationNum == 1 {
		fmt.Fprintf(os.Stderr, "%s  %s  Processing message with LLM...\n", "💭", "llm")
	} else {
		fmt.Fprintf(os.Stderr, "%s  %s  Processing again (iteration %d) with LLM...\n", "💭", "llm", iterationNum)
	}
}

// PrintToolCall prints a tool call in verbose mode.
func (p *Printer) PrintToolCall(name string, args map[string]any) {
	if !p.verbose {
		return
	}
	toolStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	argsJSON, _ := formatJSON(args)
	fmt.Fprintf(os.Stderr, "%s  %s › %s  %s\n", "⚙", "mcp", toolStyle.Render(name), argsJSON)
}

// PrintToolResult prints a tool result in verbose mode.
func (p *Printer) PrintToolResult(name string, result string, truncated bool) {
	if !p.verbose {
		return
	}
	if len(result) > 120 {
		result = result[:120]
		truncated = true
	}
	truncStr := ""
	if truncated {
		truncStr = "  … (truncated)"
	}
	fmt.Fprintf(os.Stderr, "%s  %s  %s%s\n", "↳", name, escapeNewlines(result), truncStr)
}

// PrintError prints an error message with helpful hints.
func (p *Printer) PrintError(msg string) {
	errStyle := lipgloss.NewStyle().Foreground(colorError).Bold(true)
	fmt.Fprintf(os.Stderr, "%s %s\n", errStyle.Render("✗ Error"), msg)

	// Smart hints.
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "ollama") || strings.Contains(lower, "connection refused") {
		fmt.Fprintf(os.Stderr, "→ Is Ollama running? Try: %s\n", lipgloss.NewStyle().Foreground(colorMuted).Render("ollama serve"))
	}
	if strings.Contains(lower, "not found") && strings.Contains(msg, "model") {
		// Try to extract model name.
		if idx := strings.Index(msg, `"`); idx >= 0 {
			if endIdx := strings.Index(msg[idx+1:], `"`); endIdx >= 0 {
				modelName := msg[idx+1 : idx+1+endIdx]
				fmt.Fprintf(os.Stderr, "→ Run: %s\n", lipgloss.NewStyle().Foreground(colorMuted).Render(fmt.Sprintf("ollama pull %s", modelName)))
			}
		}
	}
	if strings.Contains(lower, "tool") || strings.Contains(lower, "support") {
		fmt.Fprintf(os.Stderr, "→ See supported models: %s\n", lipgloss.NewStyle().Foreground(colorMuted).Render("mcpgo models"))
	}
	if strings.Contains(lower, "config") || strings.Contains(lower, "mcp.json") {
		fmt.Fprintf(os.Stderr, "→ Run: %s to check your config\n", lipgloss.NewStyle().Foreground(colorMuted).Render("mcpgo validate"))
	}
}

// PrintSuccess prints a success message.
func (p *Printer) PrintSuccess(msg string) {
	style := lipgloss.NewStyle().Foreground(colorSuccess)
	fmt.Fprintf(os.Stdout, "%s  %s\n", style.Render("✓"), msg)
}

// PrintWarning prints a warning message.
func (p *Printer) PrintWarning(msg string) {
	style := lipgloss.NewStyle().Foreground(colorWarning)
	fmt.Fprintf(os.Stderr, "%s  %s\n", style.Render("⚠"), msg)
}

// PrintHelp prints the REPL command help.
func (p *Printer) PrintHelp() {
	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)

	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("┌──────────────┬────────────────────────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │\n",
		headerStyle.Render("Command     "),
		headerStyle.Render("Description                    "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├──────────────┼────────────────────────────────────┤"))

	commands := []struct {
		cmd, desc string
	}{
		{"/tools", "List all available tools"},
		{"/clear", "Clear conversation history"},
		{"/model [name]", "Show/switch model"},
		{"/stats", "Show performance stats"},
		{"/servers", "Show connected MCP servers"},
		{"/help", "Show this help"},
		{"exit / quit", "Quit mcpgo"},
	}

	for _, c := range commands {
		fmt.Fprintf(os.Stdout, "│ %-12s │ %-34s │\n", c.cmd, c.desc)
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("└──────────────┴────────────────────────────────────┘"))
}

// StatsInfo holds performance metrics for display.
type StatsInfo struct {
	MessageCount    int
	ToolCallCount   int
	TotalDuration   time.Duration
	IterationCount  int
	LastResponseTime time.Duration
	AverageLoopTime  time.Duration
}

// PrintStats prints performance statistics.
func (p *Printer) PrintStats(stats StatsInfo) {
	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)

	fmt.Fprintf(os.Stdout, "\n%s\n", borderStyle.Render("┌─────────────────────────────────┬──────────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │\n",
		headerStyle.Render("Metric                          "),
		headerStyle.Render("Value              "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├─────────────────────────────────┼──────────────────────┤"))

	metricsData := []struct {
		label string
		value string
	}{
		{"Messages sent", fmt.Sprintf("%d", stats.MessageCount)},
		{"Tool calls made", fmt.Sprintf("%d", stats.ToolCallCount)},
		{"Total iterations", fmt.Sprintf("%d", stats.IterationCount)},
		{"Last response time", formatDuration(stats.LastResponseTime)},
		{"Average loop time", formatDuration(stats.AverageLoopTime)},
		{"Total session time", formatDuration(stats.TotalDuration)},
	}

	for _, m := range metricsData {
		fmt.Fprintf(os.Stdout, "│ %-31s │ %-20s │\n", m.label, m.value)
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", borderStyle.Render("└─────────────────────────────────┴──────────────────────┘"))
}

// PrintModelSuggestions prints available models for switching with autocomplete hints.
func (p *Printer) PrintModelSuggestions(currentModel string, models []ModelInfo) {
	if len(models) == 0 {
		fmt.Fprintf(os.Stdout, "No models found locally\n\n")
		return
	}

	headerStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(colorMuted)
	currentStyle := lipgloss.NewStyle().Foreground(colorSuccess)
	hintStyle := lipgloss.NewStyle().Foreground(colorMuted).Italic(true)

	fmt.Fprintf(os.Stdout, "\n%s\n", borderStyle.Render("┌─────────────────────┬──────────┬──────────────────────┐"))
	fmt.Fprintf(os.Stdout, "│ %s │ %s │ %s │\n",
		headerStyle.Render("Model               "),
		headerStyle.Render("Size     "),
		headerStyle.Render("Tool Support        "))
	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("├─────────────────────┼──────────┼──────────────────────┤"))

	for _, m := range models {
		toolSupport := "✗ no tool calling"
		modelName := m.Name
		if m.Name == currentModel {
			modelName = currentStyle.Render(modelName + " (current)")
		}
		if m.ToolSupported {
			toolSupport = lipgloss.NewStyle().Foreground(colorSuccess).Render("✓ supported")
		} else {
			toolSupport = lipgloss.NewStyle().Foreground(colorError).Render(toolSupport)
		}
		fmt.Fprintf(os.Stdout, "│ %-19s │ %-8s │ %-20s │\n", modelName, m.Size, toolSupport)
	}

	fmt.Fprintf(os.Stdout, "%s\n", borderStyle.Render("└─────────────────────┴──────────┴──────────────────────┘"))
	fmt.Fprintf(os.Stdout, "%s\n\n", hintStyle.Render("→ Use '/model <name>' to switch (e.g., /model gemma4:e4b)"))
}

// PrintExitSummary prints a summary on exit.
func (p *Printer) PrintExitSummary(stats StatsInfo) {
	if stats.MessageCount == 0 {
		return
	}

	summaryStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	fmt.Fprintf(os.Stdout, "\n%s\n", summaryStyle.Render("Session Summary"))
	fmt.Fprintf(os.Stdout, "  Messages: %d | Tools: %d | Iterations: %d | Duration: %s\n\n",
		stats.MessageCount, stats.ToolCallCount, stats.IterationCount, formatDuration(stats.TotalDuration))
}

// PrintModelAutocompleteHints prints autocomplete hints for model switching.
func (p *Printer) PrintModelAutocompleteHints(input string, models []ModelInfo) {
	if input == "" {
		return
	}

	var matches []ModelInfo
	lowerInput := strings.ToLower(input)
	for _, m := range models {
		if strings.HasPrefix(strings.ToLower(m.Name), lowerInput) {
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		return
	}

	// Show up to 3 matching models
	hintStyle := lipgloss.NewStyle().Foreground(colorMuted)
	var buf strings.Builder
	buf.WriteString("→ Did you mean: ")
	for i, m := range matches {
		if i >= 3 {
			break
		}
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(m.Name)
	}
	if len(matches) > 3 {
		buf.WriteString(", ...")
	}

	fmt.Fprintf(os.Stdout, "%s\n", hintStyle.Render(buf.String()))
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", d.Seconds()*1e6)
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// Helper functions.

func formatJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}

func escapeNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

// GetToolsTableFromMCP converts MCP tools to ToolInfo.
func GetToolsTableFromMCP(mcpClient *mcp.MultiClient) []ToolInfo {
	var tools []ToolInfo
	allTools := mcpClient.GetAllTools()

	for toolName, serverName := range allTools {
		server := mcpClient.GetServer(serverName)
		if server != nil {
			serverTools := server.GetTools()
			if tool, ok := serverTools[toolName]; ok {
				tools = append(tools, ToolInfo{
					Name:        toolName,
					Server:      serverName,
					Description: tool.Description,
				})
			}
		}
	}
	return tools
}

// GetServersTableInfo converts connected servers to ServerInfo.
func GetServersTableInfo(mcpClient *mcp.MultiClient) []ServerInfo {
	var servers []ServerInfo
	for _, name := range mcpClient.GetAllServerNames() {
		server := mcpClient.GetServer(name)
		if server != nil {
			tools := server.GetTools()
			servers = append(servers, ServerInfo{
				Name:  name,
				Tools: len(tools),
			})
		}
	}
	return servers
}
