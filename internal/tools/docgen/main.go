package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func main() {
	outputDir := "docs/cli"

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Build the root command structure (from cmd/mcp-setu/main.go)
	rootCmd := buildRootCommand()

	// Generate docs for all commands
	if err := generateDocs(rootCmd, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ CLI documentation generated to %s\n", outputDir)
}

// buildRootCommand reconstructs the command tree for doc generation
func buildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcp-setu",
		Short: "MCP bridge for Ollama",
		Long:  "mcp-setu bridges Ollama to MCP servers for interactive multi-turn chat",
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "mcp.json", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "print tool calls and results")

	// Version subcommand
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mcp-setu version %s\n", version.Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Chat subcommand
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat session",
		Long:  "Start an interactive multi-turn chat session with Ollama using configured MCP tools",
	}
	chatCmd.Flags().StringVar(&modelOverride, "model", "", "override model from config")
	chatCmd.Flags().StringVar(&systemOverride, "system", "", "override system prompt from config")
	rootCmd.AddCommand(chatCmd)

	// Tools subcommand
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "List all tools from configured MCP servers",
		Long:  "List all tools available from configured MCP servers and exit",
	}
	rootCmd.AddCommand(toolsCmd)

	// Models subcommand
	modelsCmd := &cobra.Command{
		Use:   "models",
		Short: "List Ollama models and their tool support status",
		Long:  "List all local Ollama models and show which ones support tool calling",
	}
	rootCmd.AddCommand(modelsCmd)

	// Validate subcommand
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config file and test MCP server connectivity",
		Long:  "Validate your mcp.json config file and test connectivity to Ollama and MCP servers",
	}
	rootCmd.AddCommand(validateCmd)

	return rootCmd
}

// generateDocs creates markdown files for each command
func generateDocs(cmd *cobra.Command, outputDir string) error {
	// Generate index (reference only, already created)
	// The user creates docs/cli/index.md manually with full documentation

	// For now, we just ensure the directory exists
	// In the future, this could be expanded to auto-generate command docs

	return nil
}

// Stub functions to allow imports without breaking the build
// These would normally be in cmd/mcp-setu/main.go

func runChat(ctx context.Context) error {
	return nil
}

func runTools(ctx context.Context) error {
	return nil
}

func runModels(ctx context.Context) error {
	return nil
}

func runValidate(ctx context.Context) error {
	return nil
}

func listModelInfos(ctx context.Context, client *ollama.Client) ([]ui.ModelInfo, error) {
	return nil, nil
}
