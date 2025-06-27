#!/bin/bash

# Innovation Engine MCP Auto-Setup Script
# This script ensures the MCP server is built and ready for Copilot

set -e

echo "🚀 Auto-setting up Innovation Engine MCP Server..."

# Get the directory where this script is located (should be the MCP directory)
MCP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "📁 MCP directory: $MCP_DIR"

# Find the workspace root (where .vscode exists)
WORKSPACE_ROOT="$MCP_DIR"
while [ "$WORKSPACE_ROOT" != "/" ] && [ ! -d "$WORKSPACE_ROOT/.vscode" ]; do
    WORKSPACE_ROOT="$(dirname "$WORKSPACE_ROOT")"
done

if [ "$WORKSPACE_ROOT" = "/" ]; then
    echo "❌ Could not find workspace root (no .vscode directory found)"
    exit 1
fi

echo "📁 Workspace root: $WORKSPACE_ROOT"

# Change to MCP directory
cd "$MCP_DIR"

# Check if this is the right directory
if [ ! -f "package.json" ] || ! grep -q "ie-copilot-mcp" package.json 2>/dev/null; then
    echo "❌ This doesn't appear to be the IE MCP server directory"
    exit 1
fi

# Check if build exists and is recent
if [ ! -f "build/server.js" ] || [ "src/server.ts" -nt "build/server.js" ]; then
    echo "🔨 Building MCP server..."
    npm run build
else
    echo "✅ MCP server build is up to date"
fi

# Check if server can start
echo "🧪 Testing MCP server..."
if timeout 3 node build/server.js --test > /dev/null 2>&1; then
    echo "✅ MCP server is ready!"
else
    echo "❌ MCP server test failed"
    exit 1
fi

# Optional: Test HTTP endpoint
if [ "$1" = "--with-http-test" ]; then
    echo "🌐 Testing HTTP endpoint..."
    # Start server in background
    node build/server.js --http --port=3333 > /dev/null 2>&1 &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Test the server with hello-world.md (use full path)
    RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        -d "{\"filePath\":\"$MCP_DIR/hello-world.md\"}" \
        "http://localhost:3333/execute" 2>/dev/null || echo "CURL_FAILED")
    
    if echo "$RESPONSE" | grep -q "Innovation Engine execution completed"; then
        echo "✅ HTTP test passed!"
    else
        echo "❌ HTTP test failed - server may not be responding"
        echo "Response: ${RESPONSE:0:100}..."
    fi
    
    # Clean up
    kill $SERVER_PID 2>/dev/null || true
    sleep 1
fi

echo "🎯 Innovation Engine MCP Server is ready for Copilot!"
echo ""
echo "Available commands in Copilot Chat:"
echo "  /execute this file  - Execute current markdown file"
echo "  /test this file     - Test current markdown file"
echo ""
echo "To start the server manually:"
echo "  ./run-mcp.sh        - Start in stdio mode (for VS Code)"
echo "  ./run-mcp.sh --http - Start in HTTP mode (for testing)"
echo ""
echo "To run comprehensive tests:"
echo "  ./setup-mcp.sh --with-http-test - Include HTTP endpoint testing"
