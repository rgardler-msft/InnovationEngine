# Innovation Engine MCP Server - Executable Documentation

## Introduction

This document provides step-by-step instructions for setting up and using the Innovation Engine Model Context Protocol (MCP) server. The MCP server enables GitHub Copilot to execute markdown files using the Innovation Engine CLI directly from VS Code, providing seamless integration between AI assistance and executable documentation.

The Innovation Engine MCP server exposes two primary tools to Copilot: `execute-file` for running markdown files and `test-file` for testing them. The server supports both stdio mode (for Copilot integration) and HTTP mode (for testing and debugging).

## Prerequisites

This section outlines the requirements and dependencies needed before setting up the Innovation Engine MCP server.

- Linux operating system with bash shell
- Node.js (version 14 or higher) and npm installed
- TypeScript compiler (`tsc`) available globally or locally
- Innovation Engine binary built and available at `./bin/ie`
- VS Code with GitHub Copilot extension installed
- curl command-line tool for HTTP testing
- Basic understanding of terminal commands and file permissions

Ensure that the Innovation Engine project is already built and the `./bin/ie` executable exists before proceeding with the MCP server setup.

## Setting up the environment

This section defines all environment variables used throughout the documentation. These variables provide sensible defaults and ensure consistent execution across different environments.

```bash
# Current timestamp hash for unique identifiers
HASH=$(date +%y%m%d%H%M)

# MCP server directory path
MCP_DIR="/home/rogardle/projects/InnovationEngine/sandbox/ie-copilot-mcp"

# Innovation Engine workspace root directory
IE_WORKSPACE_ROOT="/home/rogardle/projects/InnovationEngine"

# Innovation Engine binary path
IE_BINARY_PATH="${IE_WORKSPACE_ROOT}/bin/ie"

# MCP server port for HTTP mode testing
MCP_HTTP_PORT="3333"

# Node.js environment setting
NODE_ENV="production"

# Test file for validation
TEST_FILE="hello-world.md"

# MCP server name for VS Code integration
MCP_SERVER_NAME="ie-copilot-server_${HASH}"

# Build output directory
BUILD_DIR="${MCP_DIR}/build"

# TypeScript source directory
SRC_DIR="${MCP_DIR}/src"
```

Each variable serves a specific purpose in the setup and operation of the MCP server, with the `HASH` variable ensuring unique identifiers when multiple instances or deployments are needed.

## Steps

### Step 1: Navigate to MCP Directory

Move to the Innovation Engine MCP server directory to begin the setup process.

```bash
cd "${MCP_DIR}"
```

This step ensures all subsequent commands are executed from the correct working directory where the MCP server files are located.

### Step 2: Install Dependencies

Install the required Node.js dependencies for the MCP server using npm.

```bash
npm install
```

The installation process downloads and configures all necessary packages including TypeScript, the MCP SDK, and other runtime dependencies required for the server to function properly.

### Step 3: Build the MCP Server

Compile the TypeScript source code into JavaScript for execution.

```bash
npm run build
```

This command uses the TypeScript compiler to transform the source files in the `src` directory into executable JavaScript files in the `build` directory, ensuring all type checking and optimizations are applied.

### Step 4: Verify Innovation Engine Binary

Confirm that the Innovation Engine binary exists and is executable.

```bash
ls -la "${IE_BINARY_PATH}"
```

This verification step ensures the MCP server will be able to execute Innovation Engine commands when requested by Copilot, preventing runtime errors due to missing dependencies.

### Step 5: Test MCP Server Functionality

Run a comprehensive test of the MCP server including both stdio and HTTP modes.

```bash
./setup-mcp.sh --with-http-test
```

The setup script performs automated testing of the server's core functionality, validating that it can properly execute markdown files and respond to HTTP requests when configured in testing mode.

### Step 6: Configure VS Code Integration

Verify that VS Code settings include the MCP server configuration for Copilot integration.

```bash
cat "${IE_WORKSPACE_ROOT}/.vscode/settings.json"
```

This configuration enables VS Code to automatically start the MCP server when the workspace is opened, making it available to GitHub Copilot for executing markdown files seamlessly.

### Step 7: Start MCP Server

Launch the MCP server in stdio mode for Copilot integration.

```bash
./run-mcp.sh
```

The server runs in stdio mode by default, which is the required communication protocol for GitHub Copilot integration, allowing real-time interaction between the AI assistant and the Innovation Engine.

## Summary

The Innovation Engine MCP server setup provides a robust integration between GitHub Copilot and the Innovation Engine CLI. The server exposes two primary tools: `execute-file` for running markdown files and `test-file` for validating them. The setup includes comprehensive testing, error handling, and both stdio and HTTP communication modes.

Key components include automated build scripts, VS Code integration settings, and comprehensive documentation. The server automatically detects the Innovation Engine binary location and workspace root, ensuring portability across different environments and installations.

The system is designed for ease of use, with single-command setup and automated testing that validates all components are working correctly before deployment.

## Next Steps

After completing the setup, users can leverage the following capabilities and enhancements:

1. **Use Copilot Commands**: In VS Code, use `/execute this file` or `/test this file` in Copilot Chat to run markdown files
2. **HTTP Testing**: Use `./run-mcp.sh --http` to start the server in HTTP mode for debugging and testing
3. **Add Custom Tools**: Extend the server by adding new tools to the `tools` array in `src/server.ts`
4. **Monitor Logs**: Check console output for execution details and error messages during development
5. **Customize Settings**: Modify VS Code settings in `.vscode/settings.json` for environment-specific configurations

For advanced usage, consider implementing additional MCP tools for other Innovation Engine commands, setting up automated testing pipelines, or creating custom VS Code tasks for specific workflow automation needs.
