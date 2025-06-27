# Innovation Engine Copilot MCP Server

## Introduction

This MCP (Model Context Protocol) server integrates Innovation Engine with GitHub Copilot to enable seamless execution of executable documentation. When working with markdown files containing executable code blocks, you can simply type `/execute this file` in Copilot Chat and the server will automatically run the Innovation Engine command `./bin/ie execute` on the current file.

The server provides a bridge between GitHub Copilot's natural language interface and Innovation Engine's powerful executable documentation capabilities, making it easier than ever to run and test your documentation.

## Prerequisites

Before setting up this MCP server, ensure you have:

- Node.js (version 18 or higher)
- npm (Node Package Manager)
- Innovation Engine CLI installed and accessible via `./bin/ie`
- GitHub Copilot installed in VS Code
- VS Code with MCP support enabled

## Setting up the environment

This section defines all environment variables used throughout this document:

```bash
# Project configuration
export PROJECT_NAME="ie-copilot-mcp"
export PROJECT_DIR="$(pwd)"
export IE_PATH="./bin/ie"

# Server configuration
export MCP_SERVER_PORT="3001"
export MCP_SERVER_NAME="ie-copilot-server"

# Generate a unique hash for this session
export HASH=$(date +"%y%m%d%H%M")

# Unique identifiers (if needed)
export SESSION_ID="mcp_${HASH}"
```

## Steps

### Quick Setup

Use the automated setup script:

```bash
cd $PROJECT_DIR
./setup-mcp.sh
```

### Manual Setup

#### Install Dependencies

Navigate to the project directory and install the required Node.js packages:

```bash
cd $PROJECT_DIR
npm install
```

#### Build the Server

Compile the TypeScript source code:

```bash
npm run build
```

#### Test the Server

Verify the server works correctly:

```bash
./setup-mcp.sh
```

Run comprehensive tests including HTTP endpoint:

```bash
./setup-mcp.sh --with-http-test
```

Or test individual components:

```bash
./run-mcp.sh --test
./run-mcp.sh --http --port=3005
```

### Configure VS Code Integration

#### Option 1: Global Configuration

Add the MCP server configuration to your VS Code user settings:

```bash
# Open VS Code settings
code ~/.config/Code/User/settings.json
```

Add this configuration:

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
  },
  "github.copilot.chat.mcp.enabled": true
}
```

#### Option 2: Workspace Configuration

The workspace configuration is already included in `.vscode/settings.json` for this project.

### Start the MCP Server

#### For VS Code Integration (stdio mode):

```bash
./run-mcp.sh
```

#### For Testing (HTTP mode):

```bash
./run-mcp.sh --http
```

### Test the Integration

1. Open a markdown file with executable code blocks (e.g., `hello.md`)
2. Open GitHub Copilot Chat in VS Code
3. Type `/execute this file`
4. The server will automatically run Innovation Engine on the current file

## Summary

This MCP server successfully bridges GitHub Copilot and Innovation Engine, enabling natural language commands to execute documentation. The server provides two main tools:

- **execute-file**: Runs `./bin/ie execute` on a specified markdown file
- **test-file**: Runs `./bin/ie test` on a specified markdown file

Key features:

- Automatic file path detection and validation
- Intelligent working directory resolution
- Comprehensive error handling and logging
- Support for both stdio (VS Code) and HTTP (testing) modes
- Proper timeout handling for long-running executions

## Next Steps

- **Enhanced Commands**: Add support for additional Innovation Engine commands (extract, learn mode, etc.)
- **Parameter Support**: Implement custom execution parameters and flags
- **File Monitoring**: Add support for watching files and auto-execution
- **Multi-file Support**: Enable batch processing of multiple markdown files
- **Integration Testing**: Create comprehensive test suites for various scenarios
- **Documentation**: Expand documentation with more examples and use cases

## Available Commands

Once the MCP server is configured and running, you can use these commands in GitHub Copilot Chat:

- `/execute this file` - Execute the current markdown file with Innovation Engine
- `/test this file` - Test the current markdown file with Innovation Engine

## Troubleshooting

### Common Issues

1. **Server not starting**:

   - Check that Node.js dependencies are installed: `npm install`
   - Verify the build completed successfully: `npm run build`

2. **Build errors**:

   - Ensure TypeScript is properly installed: `npm install typescript`
   - Check for syntax errors in the source code

3. **Innovation Engine not found**:

   - Verify `./bin/ie` exists in the Innovation Engine root directory
   - Check file permissions: `ls -la ./bin/ie`

4. **MCP not enabled**:

   - Ensure `github.copilot.chat.mcp.enabled` is set to `true` in VS Code settings
   - Restart VS Code after configuration changes

5. **Port conflicts**:
   - Use a different port: `npm run start:http -- --port=3002`
   - Check for other processes using the port: `lsof -i :3001`

### Testing

Run the test script to verify everything is working:

```bash
./test.sh
```

This will start the server, test it with the sample `hello.md` file, and report the results.

## Project Structure

```
ie-copilot-mcp/
├── README.md              # This file
├── package.json           # Node.js dependencies and scripts
├── tsconfig.json          # TypeScript configuration
├── setup-mcp.sh           # Automated setup and test script
├── run-mcp.sh             # Auto-runner script (builds and starts server)
├── hello.md               # Sample executable documentation
├── VSCODE_SETUP.md        # VS Code integration guide
├── src/
│   └── server.ts          # Main MCP server implementation
├── build/                 # Compiled JavaScript (generated)
├── .vscode/
│   ├── settings.json      # VS Code workspace settings
│   └── tasks.json         # VS Code tasks
└── .gitignore            # Git ignore rules
```
