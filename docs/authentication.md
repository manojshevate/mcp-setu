# Authentication Guide

`mcp-setu` provides secure credential management for MCP servers using OAuth 2.1 and environment-based credentials.

## Overview

Three authentication patterns are supported:

1. **No Auth** — Unauthenticated servers (default)
2. **Bearer Token** — Static API tokens via config or environment
3. **OAuth 2.1** — Interactive OAuth with secure token storage

## Authentication Methods

### No Authentication

Servers that don't require credentials:

```json
{
  "mcpServers": {
    "local-memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

No `auth` field needed.

### Bearer Token (Static API Key)

For servers requiring a static API token or key.

#### Option 1: Environment Variable (Recommended)

Safest approach — secrets never touch config files.

**Config:**
```json
{
  "mcpServers": {
    "github-api": {
      "type": "http-streamable",
      "url": "https://api.github.com/mcp",
      "auth": {
        "type": "bearer-token",
        "tokenEnvVar": "GITHUB_MCP_TOKEN"
      }
    }
  }
}
```

**Shell:**
```bash
export GITHUB_MCP_TOKEN="ghp_xxx..."
mcp-setu chat
```

**CI/CD (GitHub Actions):**
```yaml
- name: Run mcp-setu
  env:
    GITHUB_MCP_TOKEN: ${{ secrets.GITHUB_MCP_TOKEN }}
  run: mcp-setu chat
```

#### Option 2: Convention-Based Env Var

If your config doesn't specify `tokenEnvVar`, `mcp-setu` automatically checks a convention-based variable:

`MCPSETU_<SERVERNAME>_TOKEN`

For a server named `github-api`, set:
```bash
export MCPSETU_GITHUB_API_TOKEN="ghp_xxx..."
mcp-setu chat
```

No config `auth` field needed in this case.

#### Option 3: Config File (Not Recommended)

⚠️ **Avoid storing secrets in config files.** Only use for development/testing.

```json
{
  "auth": {
    "type": "bearer-token",
    "token": "ghp_xxx..."
  }
}
```

**Risk:** Config files can be accidentally committed or copied.

### OAuth 2.1 (Interactive Authentication)

For servers supporting OAuth 2.1 (such as Anthropic's MCP servers).

#### Configuration

```json
{
  "mcpServers": {
    "anthropic-mcp": {
      "type": "http-streamable",
      "url": "https://api.anthropic.com/mcp",
      "auth": {
        "type": "oauth2",
        "authorizationServerUrl": "https://auth.anthropic.com",
        "clientId": "mcp-setu-default",
        "scopes": ["mcp:read", "mcp:write"]
      }
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `type` | Must be `"oauth2"` |
| `authorizationServerUrl` | OAuth authorization server endpoint |
| `clientId` | Client ID (mcp-setu uses public client flow, no secret embedded) |
| `scopes` | OAuth scopes to request |

#### Interactive Login

```bash
mcp-setu auth login anthropic-mcp
```

This opens your browser for authorization. After consent, the token is securely stored in your OS keyring:
- **macOS** — Keychain
- **Linux** — Secret Service (libsecret)
- **Windows** — Credential Manager

#### Check Authentication Status

```bash
mcp-setu auth status
```

Output:
```
Authentication Status
====================
  anthropic-mcp: ✓ Authenticated
  github-api: ✓ Token available (from env var)
```

#### Logout

```bash
mcp-setu auth logout anthropic-mcp
```

Removes credentials from OS keyring.

## How Token Lookup Works

When connecting to an MCP server, `mcp-setu` looks for credentials in this order (three-tier lookup):

### Tier 1: Environment Variables (Highest Priority)

If `tokenEnvVar` is configured:
```bash
export MCP_GITHUB_TOKEN="ghp_xxx..."
mcp-setu chat  # Uses env var
```

Convention-based env var (server name → uppercase):
```bash
export MCPSETU_GITHUB_API_TOKEN="ghp_xxx..."
mcp-setu chat  # Automatic lookup
```

**Use case:** CI/CD pipelines, headless servers.

### Tier 2: OS Keyring (Recommended for Interactive Use)

After `mcp-setu auth login`, tokens are stored in OS keyring:
- Encrypted at rest (OS-managed)
- Not visible in `mcp.json` or shell history
- Automatically refreshed when expired

**Use case:** Desktop/laptop development.

### Tier 3: Interactive OAuth Flow

If no token is found and the server supports OAuth 2.1:
- Browser opens for authorization
- Token exchanged via PKCE (RFC 7636)
- Token stored in keyring

**Use case:** First-time setup, or after `mcp-setu auth logout`.

## Security Best Practices

### ✅ Do

- **Use environment variables** for CI/CD and production deployments
- **Use OS keyring** for interactive development (run `mcp-setu auth login` once)
- **Use OAuth 2.1** when available (most secure for interactive use)
- **Never commit tokens** — add `mcp.json` to `.gitignore` if it contains inline tokens
- **Use HTTPS** for auth servers and MCP endpoints

### ❌ Don't

- **Don't store tokens in config files** — use `tokenEnvVar` instead
- **Don't share `mcp.json`** if it contains inline tokens
- **Don't log in via `mcp.json`** — use `mcp-setu auth login` instead
- **Don't hardcode tokens** in shell scripts (use env vars or keyring)
- **Don't use HTTP** for OAuth endpoints (always HTTPS)

## Common Scenarios

### Scenario 1: Using a Public MCP Server

Server requires an API key:

**Config (mcp.json):**
```json
{
  "mcpServers": {
    "my-api": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "bearer-token",
        "tokenEnvVar": "MY_API_KEY"
      }
    }
  }
}
```

**Setup:**
```bash
export MY_API_KEY="sk_..."
mcp-setu chat
```

### Scenario 2: OAuth-Protected MCP Server

Server uses OAuth 2.1:

**Config (mcp.json):**
```json
{
  "mcpServers": {
    "protected-api": {
      "type": "http-streamable",
      "url": "https://protected.example.com/mcp",
      "auth": {
        "type": "oauth2",
        "authorizationServerUrl": "https://auth.example.com"
      }
    }
  }
}
```

**Setup:**
```bash
mcp-setu auth login protected-api  # Opens browser, stores token in keyring
mcp-setu chat                       # Automatically uses stored token
```

### Scenario 3: Local Development + CI/CD

Develop locally with keyring, deploy with env vars:

**Local (dev):**
```bash
mcp-setu auth login my-server
mcp-setu chat  # Uses keyring token
```

**CI/CD (GitHub Actions):**
```yaml
env:
  MCPSETU_MY_SERVER_TOKEN: ${{ secrets.MY_SERVER_TOKEN }}
run: mcp-setu chat
```

## Troubleshooting

### Token Not Found

```
error: no valid token for server 'github-api'
```

**Solutions:**
1. For bearer token: Set the env var
   ```bash
   export MCPSETU_GITHUB_API_TOKEN="..."
   ```
2. For OAuth: Run login
   ```bash
   mcp-setu auth login github-api
   ```

### Keyring Not Available (Linux)

On Linux without D-Bus/libsecret, tokens fall back to encrypted file storage:
- Location: `~/.config/mcp-setu/server-name.json`
- Permissions: `0600` (read/write for user only)

### OAuth Flow Fails

Check that:
1. Authorization server URL is correct
2. Internet connection is available
3. Browser can open (not in headless environment)
4. Server supports the `code` response type

## Environment Variable Reference

Convention-based environment variables checked automatically:

| Server Name | Env Var |
|-------------|---------|
| `github` | `MCPSETU_GITHUB_TOKEN` |
| `anthropic-api` | `MCPSETU_ANTHROPIC_API_TOKEN` |
| `my-server` | `MCPSETU_MY_SERVER_TOKEN` |

Plus any custom env var specified in config via `tokenEnvVar`.

## Next Steps

- **[Configuration Guide](./configuration.md)** — Full config reference
- **[Examples](./examples.md)** — Real-world setups
- **[Troubleshooting](./troubleshooting.md)** — Common issues
