package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

type SessionModel struct {
	// Terminal
	width  int
	height int

	// Data
	ctx          context.Context
	br           *bridge.Bridge
	mcpClient    *mcp.MultiClient
	ollamaClient *ollama.Client
	printer      *ui.Printer
	history      []ollama.Message
	model        string

	// UI
	input    InputModel
	output   []string // Chat messages to display
	status   string
	quitting bool
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
		{Role: "system", Content: systemPrompt},
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
		output:       []string{},
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
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlD {
			return m, tea.Quit
		}

		if msg.Type == tea.KeyEnter {
			input := m.input.GetValue()
			if input != "" {
				m.input.AddToHistory(input)
				m.output = append(m.output, "You: "+input)
				m.input.Clear()
				m.status = "⟳ Processing..."
				return m, m.processCmd(input)
			}
			return m, nil
		}

		updatedInput, _ := m.input.Update(msg)
		m.input = updatedInput.(InputModel)
		return m, nil

	case ResponseMsg:
		m.output = append(m.output, "Assistant: "+msg.Content)
		m.status = ""
		return m, nil

	case ErrorMsg:
		m.output = append(m.output, "Error: "+msg.Error)
		m.status = ""
		return m, nil
	}

	return m, nil
}

func (m *SessionModel) processCmd(input string) tea.Cmd {
	return func() tea.Msg {
		// Handle special commands - return as messages instead of printing
		if input == "exit" || input == "quit" {
			m.quitting = true
			return nil
		}

		if input == "/tools" {
			tools := ui.GetToolsTableFromMCP(m.mcpClient)
			var lines []string
			lines = append(lines, "Available Tools:")
			for _, t := range tools {
				lines = append(lines, "  • "+t.Name+" ("+t.Server+"): "+t.Description)
			}
			m.output = append(m.output, strings.Join(lines, "\n"))
			return nil
		}

		if input == "/clear" {
			m.history = []ollama.Message{{Role: "system", Content: m.history[0].Content}}
			m.output = []string{"Conversation cleared."}
			return nil
		}

		if input == "/servers" {
			servers := ui.GetServersTableInfo(m.mcpClient)
			var lines []string
			lines = append(lines, "Connected MCP Servers:")
			for _, s := range servers {
				lines = append(lines, "  • "+s.Name+" ("+string(rune(s.Tools))+" tools)")
			}
			m.output = append(m.output, strings.Join(lines, "\n"))
			return nil
		}

		if input == "/stats" {
			stats := m.br.GetStats()
			info := fmt.Sprintf(
				"Stats: %d messages, %d tool calls, %d iterations, %v total time",
				stats.MessageCount, stats.ToolCallCount, stats.IterationCount, stats.TotalDuration)
			m.output = append(m.output, info)
			return nil
		}

		if input == "/help" {
			help := `Commands:
  /tools        - List available tools
  /servers      - Show connected servers
  /model [name] - Switch model
  /stats        - View performance stats
  /clear        - Clear conversation
  /help         - Show this help
  exit/quit     - Exit`
			m.output = append(m.output, help)
			return nil
		}

		if input == "/model" || strings.HasPrefix(input, "/model ") {
			if input == "/model" {
				models, _ := listModelInfos(m.ctx, m.ollamaClient)
				var lines []string
				lines = append(lines, "Available Models:")
				for _, model := range models {
					marker := ""
					if model.Name == m.model {
						marker = " (current)"
					}
					lines = append(lines, "  • "+model.Name+" ("+model.Size+")"+marker)
				}
				m.output = append(m.output, strings.Join(lines, "\n"))
			} else {
				parts := strings.Fields(input)
				if len(parts) == 2 {
					if err := m.br.SetModel(m.ctx, parts[1]); err == nil {
						m.model = parts[1]
						m.output = append(m.output, "Switched to model: "+parts[1])
					} else {
						m.output = append(m.output, "Error switching model: "+err.Error())
					}
				}
			}
			return nil
		}

		// Process message
		m.history = append(m.history, ollama.Message{Role: "user", Content: input})
		response, err := m.br.ProcessMessage(m.ctx, m.history)
		if err != nil {
			return ErrorMsg{Error: err.Error()}
		}

		m.history = append(m.history, ollama.Message{Role: "assistant", Content: response})
		return ResponseMsg{Content: response}
	}
}

func (m SessionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Input area
	inputView := "❯ " + m.input.View()

	// Separator
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", m.width))

	// Status
	statusLine := ""
	if m.status != "" {
		statusLine = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(m.status) + "\n"
	}

	// Output area (scrollable)
	outputHeight := m.height - 4
	startIdx := 0
	if len(m.output) > outputHeight {
		startIdx = len(m.output) - outputHeight
	}

	var outputLines []string
	for i := startIdx; i < len(m.output); i++ {
		outputLines = append(outputLines, m.output[i])
	}

	output := strings.Join(outputLines, "\n")

	return output + "\n" + sep + "\n" + statusLine + inputView
}

type ResponseMsg struct {
	Content string
}

type ErrorMsg struct {
	Error string
}

func listModelInfos(ctx context.Context, client *ollama.Client) ([]ui.ModelInfo, error) {
	models, err := client.ListLocalModels(ctx)
	if err != nil {
		return nil, err
	}
	var infos []ui.ModelInfo
	for _, m := range models {
		infos = append(infos, ui.ModelInfo{Name: m.Name, Size: m.Size})
	}
	return infos, nil
}
