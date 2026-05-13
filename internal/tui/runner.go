package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

// RunChat starts the interactive TUI chat. The banner and server list are
// rendered as the initial content of the output area — nothing is written
// to stdout before tea.Run, so AltScreen has a clean slate.
func RunChat(
	ctx context.Context,
	br *bridge.Bridge,
	mcpClient *mcp.MultiClient,
	ollamaClient *ollama.Client,
	printer *ui.Printer, // unused; kept for call-site compatibility
	model string,
	systemPrompt string,
	verbose bool,
) error {
	_ = printer
	sessionModel := NewSessionModel(ctx, br, mcpClient, ollamaClient, model, systemPrompt, verbose)
	p := tea.NewProgram(&sessionModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
