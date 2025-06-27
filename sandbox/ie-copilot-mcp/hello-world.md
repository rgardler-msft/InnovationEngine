# Hello World Example

This is a simple executable documentation example for testing the Innovation Engine MCP integration.

## Introduction

This document demonstrates basic Innovation Engine functionality with executable code blocks.

## Prerequisites

- Innovation Engine CLI installed
- Bash shell environment

## Setting up the environment

```bash
# Basic greeting configuration
export GREETING_MESSAGE="Hello from Innovation Engine!"
export USER_NAME="Developer"

# Generate a unique hash for this session
export HASH=$(date +"%y%m%d%H%M")

# Create a unique session identifier
export SESSION_ID="hello_${HASH}"
```

## Steps

### Simple Hello World

Let's start with a basic greeting:

```bash
echo "Hello World!"
```

### Personalized Greeting

Now let's use our environment variables:

```bash
echo "${GREETING_MESSAGE}"
echo "Welcome, ${USER_NAME}!"
```

### Session Information

Display our session details:

```bash
echo "Session ID: ${SESSION_ID}"
echo "Generated at: $(date)"
```

### System Information

Show some basic system information:

```bash
echo "Current directory: $(pwd)"
echo "Current user: $(whoami)"
echo "Current date: $(date)"
```

## Summary

This example demonstrates basic Innovation Engine functionality including:

- Simple command execution
- Environment variable usage
- Dynamic content generation

## Next Steps

- Try more complex examples
- Explore Innovation Engine testing features
- Create your own executable documentation
