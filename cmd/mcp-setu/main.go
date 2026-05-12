package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/config"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
	"github.com/manojshevate/mcp-setu/internal/version"
)

var (
	configPath     string
	verbose        bool
	modelOverride  string
	systemOverride string
)

// listModelInfos converts Ollama models to UI format.
func listModelInfos(ctx context.Context, client *ollama.Client) ([]ui.ModelInfo, error) {
	models, err := client.ListLocalModels(ctx)
	if err != nil {
		return nil, err
	}
	var infos []ui.ModelInfo
	for _, m := range models {
		infos = append(infos, ui.ModelInfo{
			Name:          m.Name,
			Size:          m.Size,
			ToolSupported: m.ToolSupported,
		})
	}
	return infos, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcp-setu",
		Short: "MCP bridge for Ollama",
		Long:  "mcp-setu bridges Ollama to MCP servers for interactive multi-turn chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "mcp.json", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "print tool calls and results")

	// Version subcommand.
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mcp-setu version %s\n", version.Version)
			if version.Commit != "unknown" {
				fmt.Printf("commit: %s\n", version.Commit)
			}
			if version.BuildDate != "unknown" {
				fmt.Printf("build date: %s\n", version.BuildDate)
			}
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Chat subcommand (also default).
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context())
		},
	}
	chatCmd.Flags().StringVar(&modelOverride, "model", "", "override model from config")
	chatCmd.Flags().StringVar(&systemOverride, "system", "", "override system prompt from config")
	rootCmd.AddCommand(chatCmd)

	// Tools subcommand.
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "List all tools from configured MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTools(cmd.Context())
		},
	}
	rootCmd.AddCommand(toolsCmd)

	// Models subcommand.
	modelsCmd := &cobra.Command{
		Use:   "models",
		Short: "List Ollama models and their tool support status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModels(cmd.Context())
		},
	}
	rootCmd.AddCommand(modelsCmd)

	// Validate subcommand.
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config file and test MCP server connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context())
		},
	}
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runChat(ctx context.Context) error {
	printer := ui.NewPrinter(verbose)

	// Load config.
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("%v\n\nExample config:\n%s", err, config.ExampleConfig()))
		return err
	}

	// Override model if provided.
	model := cfg.Ollama.Model
	if modelOverride != "" {
		model = modelOverride
	}

	// Override system prompt if provided.
	systemPrompt := cfg.Ollama.SystemPrompt
	if systemOverride != "" {
		systemPrompt = systemOverride
	}

	// Create Ollama client.
	ollamaClient := ollama.NewClient(cfg.Ollama.BaseURL)

	// Check tool support.
	if err := ollamaClient.CheckToolSupport(ctx, model); err != nil {
		printer.PrintError(err.Error())
		return err
	}

	// Spawn MCP servers.
	mcpClient := mcp.NewMultiClient()
	for name, serverCfg := range cfg.MCPServers {
		client, err := mcp.NewClient(ctx, name, serverCfg)
		if err != nil {
			printer.PrintError(fmt.Sprintf("Failed to start MCP server %q: %v", name, err))
			_ = mcpClient.CloseAll()
			return err
		}
		mcpClient.Add(name, client)
	}

	defer func() {
		_ = mcpClient.CloseAll()
	}()

	// Print startup info.
	serverCount := len(cfg.MCPServers)
	toolCount := len(mcpClient.GetAllTools())
	printer.PrintBanner(model, configPath, serverCount, toolCount)
	serverInfos := ui.GetServersTableInfo(mcpClient)
	printer.PrintServerTable(serverInfos)

	// Create bridge.
	br := bridge.NewBridge(ollamaClient, mcpClient, model, cfg.Ollama.Temperature, printer)

	// Setup signal handling with context cancellation (inherit parent context).
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Signal handler goroutine.
	go func() {
		<-sigCh
		printer.PrintSuccess("Shutting down...")
		cancel()
	}()

	// REPL loop.
	scanner := bufio.NewScanner(os.Stdin)
	history := []ollama.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Channel to receive input from scanner goroutine.
	inputCh := make(chan string)
	scanDoneCh := make(chan struct{})

	// Goroutine to read from scanner non-blockingly, respecting context cancellation.
	go func() {
		defer close(scanDoneCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if scanner.Scan() {
				select {
				case inputCh <- scanner.Text():
				case <-ctx.Done():
					return
				}
			} else {
				return
			}
		}
	}()

	for {
		printer.PrintUserPrompt()

		// Check if context was cancelled (signal received) or input received.
		select {
		case <-ctx.Done():
			stats := br.GetStats()
			statsInfo := ui.StatsInfo{
				MessageCount:     stats.MessageCount,
				ToolCallCount:    stats.ToolCallCount,
				TotalDuration:    stats.TotalDuration,
				IterationCount:   stats.IterationCount,
				LastResponseTime: stats.LastResponseTime,
				AverageLoopTime:  stats.AverageLoopTime,
			}
			printer.PrintExitSummary(statsInfo)
			return nil
		case input, ok := <-inputCh:
			if !ok {
				// Input goroutine closed (EOF or error).
				stats := br.GetStats()
				statsInfo := ui.StatsInfo{
					MessageCount:     stats.MessageCount,
					ToolCallCount:    stats.ToolCallCount,
					TotalDuration:    stats.TotalDuration,
					IterationCount:   stats.IterationCount,
					LastResponseTime: stats.LastResponseTime,
					AverageLoopTime:  stats.AverageLoopTime,
				}
				printer.PrintExitSummary(statsInfo)
				return nil
			}
			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			// Handle special commands.
			if input == "exit" || input == "quit" {
				printer.PrintSuccess("Goodbye!")
				stats := br.GetStats()
				statsInfo := ui.StatsInfo{
					MessageCount:     stats.MessageCount,
					ToolCallCount:    stats.ToolCallCount,
					TotalDuration:    stats.TotalDuration,
					IterationCount:   stats.IterationCount,
					LastResponseTime: stats.LastResponseTime,
					AverageLoopTime:  stats.AverageLoopTime,
				}
				printer.PrintExitSummary(statsInfo)
				return nil
			}

			if input == "/tools" {
				tools := ui.GetToolsTableFromMCP(mcpClient)
				printer.PrintToolsTable(tools)
				continue
			}

			if input == "/clear" {
				history = []ollama.Message{
					{
						Role:    "system",
						Content: systemPrompt,
					},
				}
				printer.PrintSuccess("Conversation cleared.")
				continue
			}

			// Handle /model command with optional model name.
			if input == "/model" || strings.HasPrefix(input, "/model ") {
				if input == "/model" {
					fmt.Fprintf(os.Stdout, "Current model: %s\n", model)
					// Show available models for switching.
					modelInfos, err := listModelInfos(ctx, ollamaClient)
					if err != nil {
						printer.PrintWarning(fmt.Sprintf("Failed to list models: %v", err))
					} else {
						printer.PrintModelSuggestions(model, modelInfos)
					}
				} else {
					// Extract model name from "/model <name>".
					parts := strings.Fields(input)
					if len(parts) == 2 {
						newModel := parts[1]

						if err := br.SetModel(ctx, newModel); err != nil {
							printer.PrintError(fmt.Sprintf("Failed to switch to model %q: %v", newModel, err))
							// Show suggestions on error.
							modelInfos, err := listModelInfos(ctx, ollamaClient)
							if err == nil {
								printer.PrintModelAutocompleteHints(newModel, modelInfos)
							}
						} else {
							model = newModel
							fmt.Fprintf(os.Stdout, "✓ Switched to model: %s\n\n", model)
						}
					} else {
						printer.PrintError("Usage: /model <name>")
					}
				}
				continue
			}

			if input == "/stats" {
				stats := br.GetStats()
				statsInfo := ui.StatsInfo{
					MessageCount:     stats.MessageCount,
					ToolCallCount:    stats.ToolCallCount,
					TotalDuration:    stats.TotalDuration,
					IterationCount:   stats.IterationCount,
					LastResponseTime: stats.LastResponseTime,
					AverageLoopTime:  stats.AverageLoopTime,
				}
				printer.PrintStats(statsInfo)
				continue
			}

			if input == "/servers" {
				serverInfos := ui.GetServersTableInfo(mcpClient)
				printer.PrintServerTable(serverInfos)
				continue
			}

			if input == "/help" {
				printer.PrintHelp()
				continue
			}

			// Add user message to history.
			history = append(history, ollama.Message{
				Role:    "user",
				Content: input,
			})

			// Run bridge.
			response, err := br.ProcessMessage(ctx, history)
			if err != nil {
				printer.PrintError(err.Error())
				continue
			}

			// Add response to history.
			history = append(history, ollama.Message{
				Role:    "assistant",
				Content: response,
			})

			// Note: Response is already printed by the bridge during streaming.
			// The PrintAssistantResponse call is skipped since the response
			// is displayed in real-time as it streams from Ollama.
		}
	}
}

func runTools(ctx context.Context) error {
	printer := ui.NewPrinter(verbose)

	// Load config.
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("%v", err))
		return err
	}

	// Spawn MCP servers.
	mcpClient := mcp.NewMultiClient()
	for name, serverCfg := range cfg.MCPServers {
		client, err := mcp.NewClient(ctx, name, serverCfg)
		if err != nil {
			printer.PrintError(fmt.Sprintf("Failed to start MCP server %q: %v", name, err))
			_ = mcpClient.CloseAll()
			return err
		}
		mcpClient.Add(name, client)
	}

	defer func() {
		_ = mcpClient.CloseAll()
	}()

	// List tools.
	tools := ui.GetToolsTableFromMCP(mcpClient)
	printer.PrintToolsTable(tools)

	return nil
}

func runModels(ctx context.Context) error {
	printer := ui.NewPrinter(verbose)

	// Load config.
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("%v", err))
		return err
	}

	// Create Ollama client.
	ollamaClient := ollama.NewClient(cfg.Ollama.BaseURL)

	// List models.
	modelInfos, err := listModelInfos(ctx, ollamaClient)
	if err != nil {
		printer.PrintError(fmt.Sprintf("Failed to list models: %v", err))
		return err
	}

	printer.PrintModelsTable(modelInfos)

	return nil
}

func runValidate(ctx context.Context) error {
	printer := ui.NewPrinter(verbose)

	// Load config.
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("%v", err))
		return err
	}

	printer.PrintSuccess("Config file valid")

	// Test Ollama connectivity.
	ollamaClient := ollama.NewClient(cfg.Ollama.BaseURL)
	if err := ollamaClient.CheckToolSupport(ctx, cfg.Ollama.Model); err != nil {
		printer.PrintError(fmt.Sprintf("Ollama check failed: %v", err))
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Model %q found and supports tool calling", cfg.Ollama.Model))

	// Test MCP servers.
	for name, serverCfg := range cfg.MCPServers {
		client, err := mcp.NewClient(ctx, name, serverCfg)
		if err != nil {
			printer.PrintError(fmt.Sprintf("Failed to start MCP server %q: %v", name, err))
			return err
		}

		tools := client.GetTools()
		printer.PrintSuccess(fmt.Sprintf("MCP server %q connected with %d tools", name, len(tools)))

		_ = client.Close()
	}

	printer.PrintSuccess("All validations passed!")
	return nil
}
