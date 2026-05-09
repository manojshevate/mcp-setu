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

	"github.com/manojshevate/mcpgo/internal/bridge"
	"github.com/manojshevate/mcpgo/internal/config"
	"github.com/manojshevate/mcpgo/internal/mcp"
	"github.com/manojshevate/mcpgo/internal/ollama"
	"github.com/manojshevate/mcpgo/internal/ui"
)

var (
	configPath     string
	verbose        bool
	modelOverride  string
	systemOverride string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcpgo",
		Short: "MCP bridge for Ollama",
		Long:  "mcpgo bridges Ollama to MCP servers for interactive multi-turn chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "mcp.json", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "print tool calls and results")

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

	// Setup signal handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		printer.PrintSuccess("Shutting down...")
		os.Exit(0)
	}()

	// REPL loop.
	scanner := bufio.NewScanner(os.Stdin)
	history := []ollama.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	for {
		printer.PrintUserPrompt()

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands.
		switch input {
		case "exit", "quit":
			printer.PrintSuccess("Goodbye!")
			return nil

		case "/tools":
			tools := ui.GetToolsTableFromMCP(mcpClient)
			printer.PrintToolsTable(tools)
			continue

		case "/clear":
			history = []ollama.Message{
				{
					Role:    "system",
					Content: systemPrompt,
				},
			}
			printer.PrintSuccess("Conversation cleared.")
			continue

		case "/model":
			fmt.Fprintf(os.Stdout, "Current model: %s\n\n", model)
			continue

		case "/servers":
			serverInfos := ui.GetServersTableInfo(mcpClient)
			printer.PrintServerTable(serverInfos)
			continue

		case "/help":
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

		// Print response.
		printer.PrintAssistantResponse(response)
	}

	return nil
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
	models, err := ollamaClient.ListLocalModels(ctx)
	if err != nil {
		printer.PrintError(fmt.Sprintf("Failed to list models: %v", err))
		return err
	}

	// Convert to UI format.
	var modelInfos []ui.ModelInfo
	for _, m := range models {
		modelInfos = append(modelInfos, ui.ModelInfo{
			Name:          m.Name,
			Size:          m.Size,
			ToolSupported: m.ToolSupported,
		})
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
