#!/bin/bash

# MCP Server Auto-Runner
# This script automatically ensures the MCP server is built and runs it

set -e

# Get the directory where this script is located (should be the MCP directory)
MCP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📁 MCP Directory: $MCP_DIR"

# Change to MCP directory
cd "$MCP_DIR"

# Auto-build if needed
if [ ! -f "build/server.js" ] || [ "src/server.ts" -nt "build/server.js" ]; then
    echo "🔨 Auto-building MCP server..."
    npm run build
else
    echo "✅ MCP server build is up to date"
fi

# Run the server
echo "🚀 Starting MCP server..."
exec node build/server.js "$@"
