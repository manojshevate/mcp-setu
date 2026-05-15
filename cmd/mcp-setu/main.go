package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/config"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/tui"
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
			Name: m.Name,
			Size: m.Size,
		})
	}
	return infos, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "mcp-setu",
		Short:   "MCP bridge for Ollama",
		Long:    "mcp-setu bridges Ollama to MCP servers for interactive multi-turn chat",
		Example: "  mcp-setu\n  mcp-setu chat --model qwen2.5:7b",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context())
		},
	}

	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate("mcp-setu version {{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "mcp.json", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "print tool calls and results")
	rootCmd.PersistentFlags().StringVar(&modelOverride, "model", "", "override model from config")
	rootCmd.PersistentFlags().StringVar(&systemOverride, "system", "", "override system prompt from config")

	// Version subcommand.
	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Show version information",
		Example: "  mcp-setu version",
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
		Use:     "chat",
		Short:   "Start interactive chat session",
		Example: "  mcp-setu chat\n  mcp-setu chat --model qwen2.5:7b\n  mcp-setu chat --system \"You are a helpful assistant\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChat(cmd.Context())
		},
	}
	rootCmd.AddCommand(chatCmd)

	// Tools subcommand.
	toolsCmd := &cobra.Command{
		Use:     "tools",
		Short:   "List all tools from configured MCP servers",
		Example: "  mcp-setu tools\n  mcp-setu tools --config my-config.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTools(cmd.Context())
		},
	}
	rootCmd.AddCommand(toolsCmd)

	// Models subcommand.
	modelsCmd := &cobra.Command{
		Use:     "models",
		Short:   "List Ollama models and their tool support status",
		Example: "  mcp-setu models",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModels(cmd.Context())
		},
	}
	rootCmd.AddCommand(modelsCmd)

	// Validate subcommand.
	validateCmd := &cobra.Command{
		Use:     "validate",
		Short:   "Validate config file and test MCP server connectivity",
		Example: "  mcp-setu validate\n  mcp-setu validate --config my-config.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context())
		},
	}
	rootCmd.AddCommand(validateCmd)

	// Init subcommand.
	initCmd := &cobra.Command{
		Use:     "init",
		Short:   "Create a starter mcp.json config file",
		Example: "  mcp-setu init",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
	rootCmd.AddCommand(initCmd)

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

	// Check model exists.
	if err := ollamaClient.EnsureModelExists(ctx, model); err != nil {
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

	// Create bridge. Banner is rendered inside the TUI, not before it starts,
	// so it doesn't get scrolled off-screen by AltScreen initialization.
	br := bridge.NewBridge(ollamaClient, mcpClient, model, *cfg.Ollama.Temperature, *cfg.Ollama.ContextLength, printer)

	// Setup signal handling with context cancellation (inherit parent context).
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sigCh := make(chan os.Signal, 1)

	// Build signals list based on OS — syscall.SIGTERM is not available on Windows.
	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	signal.Notify(sigCh, sigs...)
	defer signal.Stop(sigCh)

	// Signal handler goroutine.
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	// Detect TTY: fall back to non-interactive mode when stdout is not a terminal
	// (e.g., piped output, CI, scripts).
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return tui.RunNonInteractive(ctx, br, systemPrompt, verbose)
	}

	// Run the TUI chat.
	return tui.RunChat(ctx, br, mcpClient, ollamaClient, printer, model, systemPrompt, verbose)
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
	var startErrs []string
	for name, serverCfg := range cfg.MCPServers {
		client, err := mcp.NewClient(ctx, name, serverCfg)
		if err != nil {
			printer.PrintWarning(fmt.Sprintf("Failed to start MCP server %q: %v", name, err))
			startErrs = append(startErrs, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		mcpClient.Add(name, client)
	}

	defer func() {
		_ = mcpClient.CloseAll()
	}()

	// List tools.
	tools := ui.GetToolsTableFromMCP(mcpClient)
	printer.PrintToolsTable(tools)

	if len(startErrs) > 0 {
		return fmt.Errorf("some servers failed to start: %s", strings.Join(startErrs, "; "))
	}
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

	// Pre-flight: check required env vars are set
	for name, serverCfg := range cfg.MCPServers {
		if serverCfg.Auth != nil && serverCfg.Auth.TokenEnvVar != "" {
			if os.Getenv(serverCfg.Auth.TokenEnvVar) == "" {
				printer.PrintError(fmt.Sprintf("Server %q requires env var %q but it is not set", name, serverCfg.Auth.TokenEnvVar))
				return fmt.Errorf("missing required env var %q for server %q", serverCfg.Auth.TokenEnvVar, name)
			}
		}
	}

	// Test Ollama connectivity.
	ollamaClient := ollama.NewClient(cfg.Ollama.BaseURL)
	if err := ollamaClient.EnsureModelExists(ctx, cfg.Ollama.Model); err != nil {
		printer.PrintError(fmt.Sprintf("Ollama check failed: %v", err))
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Model %q found", cfg.Ollama.Model))

	// Test MCP servers.
	var serverErrs []string
	for name, serverCfg := range cfg.MCPServers {
		client, err := mcp.NewClient(ctx, name, serverCfg)
		if err != nil {
			printer.PrintError(fmt.Sprintf("MCP server %q failed: %v", name, err))
			serverErrs = append(serverErrs, name)
			continue
		}
		tools := client.GetTools()
		printer.PrintSuccess(fmt.Sprintf("MCP server %q connected with %d tools", name, len(tools)))
		_ = client.Close()
	}

	if len(serverErrs) > 0 {
		return fmt.Errorf("validation failed for servers: %s", strings.Join(serverErrs, ", "))
	}
	printer.PrintSuccess("All validations passed!")
	return nil
}

func runInit() error {
	printer := ui.NewPrinter(false)
	const configFile = "mcp.json"
	if _, err := os.Stat(configFile); err == nil {
		printer.PrintError(fmt.Sprintf("%s already exists", configFile))
		return fmt.Errorf("%s already exists", configFile)
	}
	content := config.ExampleConfig()
	if err := os.WriteFile(configFile, []byte(content+"\n"), 0644); err != nil {
		printer.PrintError(fmt.Sprintf("Failed to create %s: %v", configFile, err))
		return err
	}
	printer.PrintSuccess(fmt.Sprintf("Created %s — edit it then run: mcp-setu chat", configFile))
	return nil
}
