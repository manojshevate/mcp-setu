package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	printer      *TUIPrinter
	history      []ollama.Message
	model        string

	// UI
	input       InputModel
	output      []string // All messages (user, assistant, verbose output)
	msgChan     chan string
	status      string
	quitting    bool
	currentResp string // Buffer for streaming response
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

	msgChan := make(chan string, 100)
	tuiPrinter := NewTUIPrinter(printer, msgChan)

	return SessionModel{
		ctx:          ctx,
		br:           br,
		mcpClient:    mcpClient,
		ollamaClient: ollamaClient,
		printer:      tuiPrinter,
		history:      history,
		model:        model,
		input:        NewInputModel(),
		output:       []string{},
		msgChan:      msgChan,
	}
}

func (m SessionModel) Init() tea.Cmd {
	return m.pollMessages()
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
		m.currentResp += msg.Content
		return m, nil

	case ResponseEndMsg:
		if m.currentResp != "" {
			m.output = append(m.output, "Assistant: "+m.currentResp)
			m.currentResp = ""
		}
		m.status = ""
		return m, m.pollMessages()

	case ErrorMsg:
		m.output = append(m.output, "Error: "+msg.Error)
		m.status = ""
		return m, m.pollMessages()

	case MessageMsg:
		m.output = append(m.output, msg.Text)
		return m, m.pollMessages()
	}

	return m, nil
}

func (m SessionModel) pollMessages() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.msgChan:
			return MessageMsg{Text: msg}
		case <-time.After(50 * time.Millisecond):
			return MessageMsg{Text: ""}
		}
	}
}

type MessageMsg struct {
	Text string
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

		// Process message - bridge will use the TUI printer which sends to msgChan
		m.history = append(m.history, ollama.Message{Role: "user", Content: input})
		response, err := m.br.ProcessMessage(m.ctx, m.history)
		if err != nil {
			m.status = ""
			return ErrorMsg{Error: err.Error()}
		}

		m.history = append(m.history, ollama.Message{Role: "assistant", Content: response})
		m.status = ""
		return ResponseEndMsg{}
	}
}

func (m SessionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Input section (fixed at bottom)
	inputLine := "❯ " + m.input.View()
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", m.width))

	// Status line
	statusLine := ""
	if m.status != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true)
		statusLine = statusStyle.Render(m.status) + "\n"
	}

	// Bottom panel (input + status)
	bottomHeight := 3
	if m.status != "" {
		bottomHeight = 4
	}

	// Output area (everything above input)
	outputHeight := m.height - bottomHeight - 1

	// Show latest messages that fit
	startIdx := 0
	if len(m.output) > outputHeight {
		startIdx = len(m.output) - outputHeight
	}

	var outputLines []string
	for i := startIdx; i < len(m.output); i++ {
		outputLines = append(outputLines, m.output[i])
	}

	// Add streaming response if any
	if m.currentResp != "" {
		outputLines = append(outputLines, "Assistant: "+m.currentResp)
	}

	// Fill remaining space
	output := strings.Join(outputLines, "\n")
	fillerLines := outputHeight - len(outputLines)
	if fillerLines > 0 {
		output = strings.Repeat("\n", fillerLines) + output
	}

	// Combine: messages on top, separator, status, input at bottom
	return output + "\n" + separator + "\n" + statusLine + inputLine
}

type ResponseMsg struct {
	Content string
}

type ResponseEndMsg struct{}

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
