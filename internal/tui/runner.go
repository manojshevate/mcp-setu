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
