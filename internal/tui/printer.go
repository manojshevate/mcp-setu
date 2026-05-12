package tui

import (
	"fmt"
	"sync"

	"github.com/manojshevate/mcp-setu/internal/ui"
)

// TUIPrinter wraps the standard printer and routes output through the TUI
type TUIPrinter struct {
	original *ui.Printer
	msgChan  chan string
	mu       sync.Mutex
}

func NewTUIPrinter(original *ui.Printer, msgChan chan string) *TUIPrinter {
	return &TUIPrinter{
		original: original,
		msgChan:  msgChan,
	}
}

func (tp *TUIPrinter) PrintError(msg string) {
	tp.send("❌ Error: " + msg)
	tp.original.PrintError(msg) // Also print to original for logging
}

func (tp *TUIPrinter) PrintSuccess(msg string) {
	tp.send("✓ " + msg)
}

func (tp *TUIPrinter) PrintWarning(msg string) {
	tp.send("⚠ " + msg)
}

func (tp *TUIPrinter) PrintLLMProcessing(iterationNum int) {
	if iterationNum == 1 {
		tp.send("💭 Processing message...")
	} else {
		tp.send(fmt.Sprintf("💭 Processing again (iteration %d)...", iterationNum))
	}
}

func (tp *TUIPrinter) PrintToolCall(name string, args map[string]any) {
	tp.send(fmt.Sprintf("⚙ Tool: %s", name))
}

func (tp *TUIPrinter) PrintToolResult(name string, result string, truncated bool) {
	display := result
	if len(display) > 100 {
		display = display[:100] + "..."
	}
	tp.send(fmt.Sprintf("↳ %s: %s", name, display))
}

func (tp *TUIPrinter) PrintResponseStart() {
	// No-op for TUI
}

func (tp *TUIPrinter) PrintResponseChunk(chunk string) {
	// Send chunks to TUI
	tp.send(chunk)
}

func (tp *TUIPrinter) PrintResponseEnd() {
	// No-op for TUI
}

func (tp *TUIPrinter) PrintBanner(model, configPath string, serverCount, toolCount int) {
	// Handled in runner before TUI starts
}

func (tp *TUIPrinter) PrintServerTable(servers []ui.ServerInfo) {
	var lines []string
	lines = append(lines, "Connected Servers:")
	for _, s := range servers {
		lines = append(lines, fmt.Sprintf("  • %s (%d tools)", s.Name, s.Tools))
	}
	for _, line := range lines {
		tp.send(line)
	}
}

func (tp *TUIPrinter) PrintToolsTable(tools []ui.ToolInfo) {
	var lines []string
	lines = append(lines, "Available Tools:")
	for _, t := range tools {
		lines = append(lines, fmt.Sprintf("  • %s (%s): %s", t.Name, t.Server, t.Description))
	}
	for _, line := range lines {
		tp.send(line)
	}
}

func (tp *TUIPrinter) PrintModelsTable(models []ui.ModelInfo) {
	var lines []string
	lines = append(lines, "Available Models:")
	for _, m := range models {
		lines = append(lines, fmt.Sprintf("  • %s (%s)", m.Name, m.Size))
	}
	for _, line := range lines {
		tp.send(line)
	}
}

func (tp *TUIPrinter) PrintModelSuggestions(currentModel string, models []ui.ModelInfo) {
	var lines []string
	lines = append(lines, "Available Models:")
	for _, m := range models {
		current := ""
		if m.Name == currentModel {
			current = " (current)"
		}
		lines = append(lines, fmt.Sprintf("  • %s (%s)%s", m.Name, m.Size, current))
	}
	for _, line := range lines {
		tp.send(line)
	}
}

func (tp *TUIPrinter) PrintModelAutocompleteHints(input string, models []ui.ModelInfo) {
	var lines []string
	lines = append(lines, "Matching Models:")
	count := 0
	for _, m := range models {
		if count >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf("  • %s", m.Name))
		count++
	}
	for _, line := range lines {
		tp.send(line)
	}
}

func (tp *TUIPrinter) PrintHelp() {
	help := `Available Commands:
  /tools        - List available tools
  /servers      - Show connected servers
  /model [name] - Switch model
  /stats        - View performance stats
  /clear        - Clear conversation
  /help         - Show this help
  exit / quit   - Exit`
	tp.send(help)
}

func (tp *TUIPrinter) PrintStats(stats ui.StatsInfo) {
	info := fmt.Sprintf(
		"Stats: %d messages | %d tool calls | %d iterations | %v total",
		stats.MessageCount, stats.ToolCallCount, stats.IterationCount, stats.TotalDuration)
	tp.send(info)
}

func (tp *TUIPrinter) PrintExitSummary(stats ui.StatsInfo) {
	// Not needed in TUI
}

func (tp *TUIPrinter) send(msg string) {
	select {
	case tp.msgChan <- msg:
	default:
		// Channel full, skip
	}
}
