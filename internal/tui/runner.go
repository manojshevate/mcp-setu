package tui

import (
	"context"

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
	// Print welcome banner before starting TUI (it gets cleared by AltScreen)
	serverCount := len(mcpClient.GetAllServerNames())
	toolCount := len(mcpClient.GetAllTools())
	printer.PrintBanner(model, "mcp.json", serverCount, toolCount)

	sessionModel := NewSessionModel(
		ctx,
		br,
		mcpClient,
		ollamaClient,
		printer,
		model,
		systemPrompt,
	)

	// Run without AltScreen to preserve banner and allow proper rendering
	p := tea.NewProgram(sessionModel)
	_, err := p.Run()
	return err
}
