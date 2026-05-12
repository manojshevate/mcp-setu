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
	sessionModel := NewSessionModel(
		ctx,
		br,
		mcpClient,
		ollamaClient,
		printer,
		model,
		systemPrompt,
	)

	p := tea.NewProgram(sessionModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
