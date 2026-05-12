package tui

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

func RunChat(
	ctx context.Context,
	br *bridge.Bridge,
	mcpClient *mcp.MultiClient,
	ollamaClient *ollama.Client,
	printer *ui.Printer,
	model string,
	systemPrompt string,
) error {
	// Print welcome info before starting TUI
	serverCount := len(mcpClient.GetAllServerNames())
	toolCount := len(mcpClient.GetAllTools())
	printer.PrintBanner(model, "mcp.json", serverCount, toolCount)
	serverInfos := ui.GetServersTableInfo(mcpClient)
	printer.PrintServerTable(serverInfos)

	// Add spacing
	fmt.Fprintf(os.Stdout, "\n")

	// Create session (which creates TUI printer channel)
	sessionModel := NewSessionModel(
		ctx,
		br,
		mcpClient,
		ollamaClient,
		printer, // Original printer (ignored, TUI creates its own)
		model,
		systemPrompt,
	)

	// The TUI printer will capture output via the message channel
	// The original printer is still used but TUI printer wraps it

	p := tea.NewProgram(sessionModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
