#!/bin/bash

# MCP Server Auto-Runner
# This script automatically ensures the MCP server is built and runs it

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_DIR="$SCRIPT_DIR/ie-copilot-mcp"

# Auto-build if needed
if [ ! -f "$MCP_DIR/build/server.js" ] || [ "$MCP_DIR/src/server.ts" -nt "$MCP_DIR/build/server.js" ]; then
    echo "🔨 Auto-building MCP server..."
    cd "$MCP_DIR"
    # Use npx to ensure we find the local typescript
    npx tsc
    cd "$SCRIPT_DIR"
fi

# Run the server
echo "🚀 Starting MCP server..."
cd "$MCP_DIR"
exec node build/server.js "$@"
