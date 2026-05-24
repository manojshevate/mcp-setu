package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/manojshevate/mcp-setu/internal/auth"
	"github.com/manojshevate/mcp-setu/internal/config"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
		Long:  "Manage OAuth 2.1 credentials for MCP servers",
	}

	authCmd.AddCommand(newAuthLoginCmd())
	authCmd.AddCommand(newAuthLogoutCmd())
	authCmd.AddCommand(newAuthStatusCmd())

	return authCmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "login <server>",
		Short:   "Login to an MCP server",
		Example: "  mcp-setu auth login github\n  mcp-setu auth login github --config my-config.json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogin(cmd.Context(), args[0])
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "logout <server>",
		Short:   "Logout from an MCP server",
		Example: "  mcp-setu auth logout github",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout(cmd.Context(), args[0])
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show authentication status for all servers",
		Example: "  mcp-setu auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus(cmd.Context())
		},
	}
}

func runAuthLogin(ctx context.Context, serverName string) error {
	printer := ui.NewPrinter(verbose)

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("Config error: %v", err))
		return err
	}

	// Find server config
	serverCfg, ok := cfg.MCPServers[serverName]
	if !ok {
		printer.PrintError(fmt.Sprintf("Server %q not found in config", serverName))
		return fmt.Errorf("server not found")
	}

	// Check if server supports OAuth2
	if serverCfg.Auth == nil || serverCfg.Auth.Type != "oauth2" {
		printer.PrintError(fmt.Sprintf("Server %q does not support OAuth2 authentication", serverName))
		return fmt.Errorf("server does not support OAuth2")
	}

	if serverCfg.Auth.AuthorizationServerURL == "" {
		printer.PrintError(fmt.Sprintf("Authorization server URL not configured for server %q", serverName))
		return fmt.Errorf("missing authorization server URL")
	}

	// Create auth manager
	mgr, err := auth.NewManager()
	if err != nil {
		printer.PrintError(fmt.Sprintf("Failed to initialize auth: %v", err))
		return err
	}

	// Perform login
	fmt.Printf("Authenticating to server %q...\n", serverName)
	if err := mgr.Login(ctx, serverName, serverCfg.Auth); err != nil {
		printer.PrintError(fmt.Sprintf("Authentication failed: %v", err))
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Successfully authenticated to server %q", serverName))
	return nil
}

func runAuthLogout(ctx context.Context, serverName string) error {
	printer := ui.NewPrinter(verbose)

	// Load config to verify server exists
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("Config error: %v", err))
		return err
	}

	if _, ok := cfg.MCPServers[serverName]; !ok {
		printer.PrintError(fmt.Sprintf("Server %q not found in config", serverName))
		return fmt.Errorf("server not found")
	}

	// Create auth manager and logout
	mgr, err := auth.NewManager()
	if err != nil {
		printer.PrintError(fmt.Sprintf("Failed to initialize auth: %v", err))
		return err
	}

	if err := mgr.Logout(ctx, serverName); err != nil {
		printer.PrintError(fmt.Sprintf("Logout failed: %v", err))
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Logged out from server %q", serverName))
	return nil
}

func runAuthStatus(ctx context.Context) error {
	printer := ui.NewPrinter(verbose)

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		printer.PrintError(fmt.Sprintf("Config error: %v", err))
		return err
	}

	// Create auth manager
	mgr, err := auth.NewManager()
	if err != nil {
		printer.PrintError(fmt.Sprintf("Failed to initialize auth: %v", err))
		return err
	}

	fmt.Println("Authentication Status")
	fmt.Println("====================")

	for serverName, serverCfg := range cfg.MCPServers {
		if serverCfg.Auth == nil {
			fmt.Printf("  %s: No authentication required\n", serverName)
			continue
		}

		authType := serverCfg.Auth.Type
		if authType == "" || authType == "none" {
			fmt.Printf("  %s: No authentication required\n", serverName)
			continue
		}

		if authType == "bearer-token" || authType == "env" {
			// Check if token is available
			if serverCfg.Auth.TokenEnvVar != "" && os.Getenv(serverCfg.Auth.TokenEnvVar) != "" {
				fmt.Printf("  %s: ✓ Token available (from env var)\n", serverName)
			} else if serverCfg.Auth.Token != "" {
				fmt.Printf("  %s: ✓ Token configured\n", serverName)
			} else {
				fmt.Printf("  %s: ✗ Token not available\n", serverName)
			}
			continue
		}

		if authType == "oauth2" {
			// Check if token is stored
			storedToken, _ := mgr.GetToken(ctx, serverName, serverCfg.Auth)
			if storedToken != "" {
				fmt.Printf("  %s: ✓ Authenticated\n", serverName)
			} else {
				fmt.Printf("  %s: ✗ Not authenticated (run 'mcp-setu auth login %s')\n", serverName, serverName)
			}
		}
	}

	return nil
}
