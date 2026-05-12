package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

type SessionModel struct {
	width        int
	height       int
	input        InputModel
	ctx          context.Context
	br           *bridge.Bridge
	mcpClient    *mcp.MultiClient
	ollamaClient *ollama.Client
	printer      *ui.Printer
	history      []ollama.Message
	model        string
	processing   bool
	lastResponse string
}

func NewSessionModel(
	ctx context.Context,
	br *bridge.Bridge,
	mcpClient *mcp.MultiClient,
	ollamaClient *ollama.Client,
	printer *ui.Printer,
	model string,
	systemPrompt string,
) SessionModel {
	history := []ollama.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	return SessionModel{
		ctx:          ctx,
		br:           br,
		mcpClient:    mcpClient,
		ollamaClient: ollamaClient,
		printer:      printer,
		history:      history,
		model:        model,
		input:        NewInputModel(),
	}
}

func (m SessionModel) Init() tea.Cmd {
	return nil
}

func (m SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width - 4)
		return m, nil

	case tea.KeyMsg:
		// Ctrl+C or Ctrl+D to quit
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlD {
			return m, tea.Quit
		}

		// Enter to submit
		if msg.Type == tea.KeyEnter && !m.processing {
			input := m.input.GetValue()
			if input != "" {
				m.input.AddToHistory(input)
				m.input.Clear()
				return m, m.submitInput(input)
			}
			return m, nil
		}

		// Pass other keys to input
		updatedInput, cmd := m.input.Update(msg)
		m.input = updatedInput.(InputModel)
		return m, cmd
	}

	return m, nil
}

func (m *SessionModel) submitInput(input string) tea.Cmd {
	return func() tea.Msg {
		// Handle special commands
		if handled, _ := m.handleSpecialCommand(input); handled {
			return nil
		}

		m.processing = true
		defer func() { m.processing = false }()

		// Add user message
		m.history = append(m.history, ollama.Message{
			Role:    "user",
			Content: input,
		})

		// Process message (printer handles output to stderr)
		response, err := m.br.ProcessMessage(m.ctx, m.history)
		if err != nil {
			m.printer.PrintError(err.Error())
			m.history = m.history[:len(m.history)-1] // Remove user message on error
			return nil
		}

		// Add response to history
		m.history = append(m.history, ollama.Message{
			Role:    "assistant",
			Content: response,
		})

		m.lastResponse = response
		return nil
	}
}

func (m *SessionModel) handleSpecialCommand(input string) (bool, error) {
	if input == "exit" || input == "quit" {
		return true, nil
	}

	if input == "/tools" {
		tools := ui.GetToolsTableFromMCP(m.mcpClient)
		m.printer.PrintToolsTable(tools)
		return true, nil
	}

	if input == "/clear" {
		m.history = []ollama.Message{
			{
				Role:    "system",
				Content: m.history[0].Content,
			},
		}
		m.printer.PrintSuccess("Conversation cleared.")
		return true, nil
	}

	if input == "/servers" {
		servers := ui.GetServersTableInfo(m.mcpClient)
		m.printer.PrintServerTable(servers)
		return true, nil
	}

	if input == "/stats" {
		stats := m.br.GetStats()
		statsInfo := ui.StatsInfo{
			MessageCount:     stats.MessageCount,
			ToolCallCount:    stats.ToolCallCount,
			TotalDuration:    stats.TotalDuration,
			IterationCount:   stats.IterationCount,
			LastResponseTime: stats.LastResponseTime,
			AverageLoopTime:  stats.AverageLoopTime,
		}
		m.printer.PrintStats(statsInfo)
		return true, nil
	}

	if input == "/help" {
		m.printer.PrintHelp()
		return true, nil
	}

	if input == "/model" || strings.HasPrefix(input, "/model ") {
		if input == "/model" {
			models, err := listModelInfos(m.ctx, m.ollamaClient)
			if err != nil {
				m.printer.PrintError(err.Error())
			} else {
				m.printer.PrintModelSuggestions(m.model, models)
			}
		} else {
			parts := strings.Fields(input)
			if len(parts) == 2 {
				newModel := parts[1]
				if err := m.br.SetModel(m.ctx, newModel); err != nil {
					models, _ := listModelInfos(m.ctx, m.ollamaClient)
					m.printer.PrintModelAutocompleteHints(newModel, models)
					return true, err
				}
				m.model = newModel
				m.printer.PrintSuccess("Model switched to: " + newModel)
			}
		}
		return true, nil
	}

	return false, nil
}

func (m SessionModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Simple layout: input at bottom with styling
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	statusText := ""
	if m.processing {
		statusText = "⟳ Processing..."
	}

	inputSection := borderStyle.Render(
		promptStyle.Render("❯ ") + m.input.View(),
	)

	// Show status if processing
	if m.processing {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true).
			Padding(0, 2)
		return statusStyle.Render(statusText) + "\n" + inputSection
	}

	return inputSection
}

func listModelInfos(ctx context.Context, client *ollama.Client) ([]ui.ModelInfo, error) {
	models, err := client.ListLocalModels(ctx)
	if err != nil {
		return nil, err
	}
	var infos []ui.ModelInfo
	for _, m := range models {
		infos = append(infos, ui.ModelInfo{
			Name: m.Name,
			Size: m.Size,
		})
	}
	return infos, nil
}
