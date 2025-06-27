import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { exec } from "child_process";
import { promisify } from "util";
import { promises as fs } from "fs";
import * as fsSync from "fs";
import path from "path";
import http from "http";

const execAsync = promisify(exec);

// Configuration
const DEFAULT_PORT = 3001;
const IE_BINARY_PATH = "./bin/ie";  // Always relative to IE root directory

class IECopilotMCPServer {
    private server: McpServer;

    constructor() {
        this.server = new McpServer({
            name: "ie-copilot-server",
            version: "1.0.0",
            capabilities: {
                tools: {},
                resources: {}
            }
        });

        this.setupTools();
    }

    private setupTools() {
        // Main tool for executing files with Innovation Engine
        this.server.tool(
            "execute-file",
            "Execute a markdown file using Innovation Engine. This runs the './bin/ie execute' command on the specified file.",
            {
                filePath: z.string().describe("Path to the markdown file to execute with Innovation Engine"),
                workingDirectory: z.string().optional().describe("Working directory for execution (defaults to Innovation Engine root)")
            },
            async (args: { filePath: string; workingDirectory?: string }) => {
                try {
                    const { filePath, workingDirectory } = args;
                    
                    // Validate file exists and is a markdown file
                    const fileExists = await this.fileExists(filePath);
                    if (!fileExists) {
                        return {
                            content: [
                                { 
                                    type: "text", 
                                    text: `❌ Error: File '${filePath}' not found or is not accessible.` 
                                }
                            ]
                        };
                    }

                    if (!filePath.endsWith('.md')) {
                        return {
                            content: [
                                { 
                                    type: "text", 
                                    text: `⚠️  Warning: File '${filePath}' doesn't appear to be a markdown file. Innovation Engine works best with .md files containing executable code blocks.` 
                                }
                            ]
                        };
                    }

                    // Determine working directory
                    const cwd = workingDirectory || this.findInnovationEngineRoot(filePath);
                    
                    // Execute the Innovation Engine command
                    const command = `${IE_BINARY_PATH} execute "${filePath}"`;
                    console.error(`🚀 Executing: ${command}`);
                    console.error(`📁 Working directory: ${cwd}`);

                    const { stdout, stderr } = await execAsync(command, { 
                        cwd,
                        timeout: 30000 // 30 second timeout
                    });

                    const output = stdout || stderr || "Command executed successfully with no output.";
                    
                    return {
                        content: [
                            { 
                                type: "text", 
                                text: `✅ Innovation Engine execution completed:\n\n${output}` 
                            }
                        ]
                    };

                } catch (error: any) {
                    console.error("❌ Execution error:", error);
                    
                    let errorMessage = "Unknown error occurred";
                    if (error.code === 'ENOENT') {
                        errorMessage = `Innovation Engine binary not found. Please ensure '${IE_BINARY_PATH}' exists and is executable.`;
                    } else if (error.code === 'ETIMEDOUT') {
                        errorMessage = "Execution timed out after 30 seconds.";
                    } else if (error.stderr) {
                        errorMessage = error.stderr;
                    } else if (error.message) {
                        errorMessage = error.message;
                    }

                    return {
                        content: [
                            { 
                                type: "text", 
                                text: `❌ Error executing file: ${errorMessage}` 
                            }
                        ]
                    };
                }
            }
        );

        // Additional tool for testing files
        this.server.tool(
            "test-file",
            "Test a markdown file using Innovation Engine. This runs the './bin/ie test' command on the specified file.",
            {
                filePath: z.string().describe("Path to the markdown file to test with Innovation Engine"),
                workingDirectory: z.string().optional().describe("Working directory for testing (defaults to Innovation Engine root)")
            },
            async (args: { filePath: string; workingDirectory?: string }) => {
                try {
                    const { filePath, workingDirectory } = args;
                    
                    // Validate file exists
                    const fileExists = await this.fileExists(filePath);
                    if (!fileExists) {
                        return {
                            content: [
                                { 
                                    type: "text", 
                                    text: `❌ Error: File '${filePath}' not found or is not accessible.` 
                                }
                            ]
                        };
                    }

                    // Determine working directory
                    const cwd = workingDirectory || this.findInnovationEngineRoot(filePath);
                    
                    // Execute the Innovation Engine test command
                    const command = `${IE_BINARY_PATH} test "${filePath}"`;
                    console.error(`🧪 Testing: ${command}`);
                    console.error(`📁 Working directory: ${cwd}`);

                    const { stdout, stderr } = await execAsync(command, { 
                        cwd,
                        timeout: 30000
                    });

                    const output = stdout || stderr || "Test completed with no output.";
                    
                    return {
                        content: [
                            { 
                                type: "text", 
                                text: `🧪 Innovation Engine test completed:\n\n${output}` 
                            }
                        ]
                    };

                } catch (error: any) {
                    console.error("❌ Test error:", error);
                    
                    let errorMessage = "Unknown error occurred";
                    if (error.code === 'ENOENT') {
                        errorMessage = `Innovation Engine binary not found. Please ensure '${IE_BINARY_PATH}' exists and is executable.`;
                    } else if (error.code === 'ETIMEDOUT') {
                        errorMessage = "Test timed out after 30 seconds.";
                    } else if (error.stderr) {
                        errorMessage = error.stderr;
                    } else if (error.message) {
                        errorMessage = error.message;
                    }

                    return {
                        content: [
                            { 
                                type: "text", 
                                text: `❌ Error testing file: ${errorMessage}` 
                            }
                        ]
                    };
                }
            }
        );
    }

    private async fileExists(filePath: string): Promise<boolean> {
        try {
            await fs.access(filePath);
            return true;
        } catch {
            return false;
        }
    }

    private findInnovationEngineRoot(filePath: string): string {
        // Try to find the Innovation Engine root directory
        // Look for the directory containing the 'bin/ie' binary
        let currentDir = path.dirname(path.resolve(filePath));
        
        while (currentDir !== path.dirname(currentDir)) {
            const binPath = path.join(currentDir, 'bin', 'ie');
            try {
                // Check if bin/ie exists in this directory
                if (fsSync.existsSync(binPath)) {
                    return currentDir;
                }
            } catch {
                // Continue searching
            }
            currentDir = path.dirname(currentDir);
        }
        
        // Also try some known relative paths from the current working directory
        const possibleRoots = [
            process.cwd(),                    // Current directory
            path.resolve(process.cwd(), "../.."),  // Two levels up from MCP dir
            path.resolve(process.cwd(), "../../.."), // Three levels up
            "/home/rogardle/projects/InnovationEngine"  // Absolute fallback
        ];
        
        for (const rootPath of possibleRoots) {
            const binPath = path.join(rootPath, 'bin', 'ie');
            try {
                if (fsSync.existsSync(binPath)) {
                    return rootPath;
                }
            } catch {
                continue;
            }
        }
        
        // Default fallback
        return '/home/rogardle/projects/InnovationEngine';
    }

    async startStdio() {
        const transport = new StdioServerTransport();
        await this.server.connect(transport);
        console.error("🎯 IE Copilot MCP Server started in stdio mode");
    }

    async startHttp(port: number = DEFAULT_PORT) {
        // Create HTTP endpoint for testing
        const httpServer = http.createServer(async (req: any, res: any) => {
            if (req.method === 'POST' && req.url === '/execute') {
                let body = '';
                req.on('data', (chunk: any) => body += chunk);
                req.on('end', async () => {
                    try {
                        const { filePath, workingDirectory } = JSON.parse(body);
                        
                        // Execute the IE command directly
                        const cwd = workingDirectory || this.findInnovationEngineRoot(filePath);
                        const command = `${IE_BINARY_PATH} execute "${filePath}"`;
                        
                        const { stdout, stderr } = await execAsync(command, { 
                            cwd,
                            timeout: 30000
                        });

                        const output = stdout || stderr || "Command executed successfully with no output.";
                        const result = {
                            content: [
                                { 
                                    type: "text", 
                                    text: `✅ Innovation Engine execution completed:\n\n${output}` 
                                }
                            ]
                        };
                        
                        res.writeHead(200, { 'Content-Type': 'application/json' });
                        res.end(JSON.stringify(result));
                    } catch (error: any) {
                        res.writeHead(500, { 'Content-Type': 'application/json' });
                        res.end(JSON.stringify({ error: error.message }));
                    }
                });
            } else {
                res.writeHead(404);
                res.end('Not found');
            }
        });

        httpServer.listen(port, () => {
            console.error(`🌐 IE Copilot MCP Server HTTP endpoint running on http://localhost:${port}/execute`);
            console.error(`📘 Example: curl -X POST -H "Content-Type: application/json" -d '{"filePath":"hello.md"}' http://localhost:${port}/execute`);
        });
    }
}

async function main() {
    const server = new IECopilotMCPServer();
    
    // Check command line arguments
    const args = process.argv.slice(2);
    const useHttp = args.includes('--http');
    const testMode = args.includes('--test');
    
    if (testMode) {
        console.error("🧪 Test mode - not starting server");
        process.exit(0);
    }
    
    try {
        if (useHttp) {
            // HTTP mode for testing
            const portArg = args.find((arg: string) => arg.startsWith('--port='));
            const port = portArg ? parseInt(portArg.split('=')[1], 10) : DEFAULT_PORT;
            await server.startHttp(port);
        } else {
            // Default stdio mode for VS Code integration
            await server.startStdio();
        }
    } catch (error) {
        console.error("❌ Fatal error starting server:", error);
        process.exit(1);
    }
}

// Handle graceful shutdown
process.on('SIGINT', () => {
    console.error("\n👋 IE Copilot MCP Server shutting down...");
    process.exit(0);
});

process.on('SIGTERM', () => {
    console.error("\n👋 IE Copilot MCP Server shutting down...");
    process.exit(0);
});

main().catch(error => {
    console.error("❌ Unhandled error:", error);
    process.exit(1);
});
