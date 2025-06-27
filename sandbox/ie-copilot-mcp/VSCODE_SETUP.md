# VS Code MCP Configuration

To integrate this MCP server with GitHub Copilot in VS Code, you need to configure the MCP settings.

## Global Configuration

Add this configuration to your VS Code user settings (`settings.json`):

```json
{
  "mcp": {
    "servers": {
      "ie-copilot-server": {
        "command": "node",
        "args": ["build/server.js"],
        "cwd": "${workspaceFolder}",
        "env": {
          "NODE_ENV": "production"
        }
      }
    }
  }
}
```

## Workspace Configuration

For workspace-specific configuration, the settings are already included in `.vscode/settings.json`.

## Available Commands

Once configured, you can use these commands in GitHub Copilot Chat:

1. **Execute File**: `/execute this file` - Executes the current markdown file with Innovation Engine
2. **Test File**: `/test this file` - Tests the current markdown file with Innovation Engine

## Tools Available

The MCP server exposes these tools:

- `execute-file`: Runs `./bin/ie execute` on a specified file
- `test-file`: Runs `./bin/ie test` on a specified file

## Troubleshooting

1. **Server not starting**: Check that Node.js dependencies are installed (`npm install`)
2. **Build errors**: Run `npm run build` to compile TypeScript
3. **Innovation Engine not found**: Ensure `./bin/ie` exists in the Innovation Engine root directory
4. **MCP not enabled**: Verify that `github.copilot.chat.mcp.enabled` is set to `true`
