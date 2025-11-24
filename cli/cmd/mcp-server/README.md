# Hasura MCP Server

A Model Context Protocol (MCP) server that enables AI assistants to interact with Hasura GraphQL Engine instances. This server exposes Hasura operations as MCP tools, resources, and prompts, allowing seamless integration with AI assistants like Claude Desktop and ChatGPT.

## Features

- **Metadata Management**: Export, apply, reload, and manage Hasura metadata
- **Database Operations**: Execute SQL queries on PostgreSQL sources
- **Schema Introspection**: Access GraphQL schema and server information
- **Resources**: Direct access to metadata, version, and schema via URIs
- **Interactive Prompts**: Guided workflows for common tasks
- **Read-Only Mode**: Safe mode for exploring without making changes
- **Docker Support**: Lightweight Alpine-based container (~41MB)

## Quick Start

### Using Docker (Recommended)

```bash
# Pull the image
docker pull sanjay3290/hasura-mcp:latest

# Run with local Hasura
docker run -i --rm \
  --network host \
  sanjay3290/hasura-mcp:latest \
  --endpoint http://localhost:8080 \
  --admin-secret your-secret

# Run with remote Hasura
docker run -i --rm \
  sanjay3290/hasura-mcp:latest \
  --endpoint https://your-hasura.com \
  --admin-secret your-secret
```

### Build from Source

```bash
cd cli/cmd/mcp-server
go build -o hasura-mcp-server
./hasura-mcp-server --endpoint http://localhost:8080 --admin-secret your-secret
```

### Build Docker Image Locally

```bash
# From the graphql-engine repository root
docker build -f cli/cmd/mcp-server/Dockerfile -t hasura-mcp-server:latest cli/

# Run the local build
docker run -i --rm \
  --network host \
  hasura-mcp-server:latest \
  --endpoint http://localhost:8080 \
  --admin-secret your-secret
```

## Configuration

### Command-Line Flags

- `--endpoint`, `-e`: Hasura GraphQL Engine endpoint (required)
- `--admin-secret`, `-s`: Hasura admin secret
- `--read-only`, `-r`: Enable read-only mode (default: false)
- `--name`, `-n`: MCP server name (default: "hasura-mcp-server")
- `--version`, `-v`: Print version information

### Environment Variables

- `HASURA_GRAPHQL_ENDPOINT`: Hasura endpoint URL
- `HASURA_GRAPHQL_ADMIN_SECRET`: Admin secret for authentication

### Docker Environment Variables

```bash
docker run -i --rm \
  -e HASURA_GRAPHQL_ENDPOINT=http://localhost:8080 \
  -e HASURA_GRAPHQL_ADMIN_SECRET=your-secret \
  hasura-mcp-server:latest
```

## MCP Client Integration

### Claude Desktop

Add to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

#### Using Docker Image

```json
{
  "mcpServers": {
    "hasura": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--network",
        "host",
        "sanjay3290/hasura-mcp:latest",
        "--endpoint",
        "http://localhost:8080",
        "--admin-secret",
        "your-secret"
      ]
    }
  }
}
```

**Note**: Use `--network host` for localhost connections. Remove it for remote Hasura instances.

#### Using Binary

```json
{
  "mcpServers": {
    "hasura": {
      "command": "/path/to/hasura-mcp-server",
      "args": [
        "--endpoint",
        "http://localhost:8080",
        "--admin-secret",
        "your-secret"
      ]
    }
  }
}
```

#### Read-Only Mode (Production)

```json
{
  "mcpServers": {
    "hasura-prod": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "sanjay3290/hasura-mcp:latest",
        "--endpoint",
        "https://prod.hasura.app",
        "--admin-secret",
        "prod-secret",
        "--read-only"
      ]
    }
  }
}
```

### Other MCP Clients

The server uses stdio transport and is compatible with any MCP client that supports the Model Context Protocol specification (2024-11-05 or later).

## Available Capabilities

### Tools (9)

#### Metadata Management
- **export_metadata**: Export complete Hasura metadata (JSON/YAML)
- **apply_metadata**: Apply metadata to Hasura server
- **reload_metadata**: Reload Hasura metadata cache
- **clear_metadata**: Clear all metadata (destructive)
- **get_inconsistent_metadata**: Check for inconsistent metadata
- **drop_inconsistent_metadata**: Remove inconsistent metadata objects

#### Database Operations
- **run_sql**: Execute SQL on PostgreSQL source
- **get_version**: Get Hasura server version
- **get_schema**: Get GraphQL schema introspection

### Resources (4)

- `hasura://metadata` - Current server metadata
- `hasura://version` - Server version information
- `hasura://schema` - GraphQL schema introspection
- `hasura://inconsistent-metadata` - List of inconsistent metadata objects

### Prompts (4)

- **setup_table**: Guide for tracking and configuring a table
- **debug_metadata**: Diagnostic guide for metadata issues
- **create_migration**: Template for creating migrations
- **setup_permissions**: Guide for configuring permissions

## Example Usage with AI Assistants

Once configured with Claude Desktop, you can interact naturally:

```
"Export the metadata from my Hasura instance"
"Check if there are any inconsistent metadata objects"
"Show me all tables in the public schema"
"What version of Hasura am I running?"
"Help me set up the users table"
```

## Read-Only Mode

When running in read-only mode (`--read-only`), the following tools are disabled:
- `apply_metadata`
- `clear_metadata`
- `drop_inconsistent_metadata`

The `run_sql` tool enforces `read_only: true` for all queries, preventing data modifications.

This mode is useful for:
- Exploring production systems safely
- Providing read-only access to team members
- Debugging without risk of changes

## Docker Details

### Image Specifications
- **Base**: Alpine Linux 3.19
- **Size**: ~41 MB
- **User**: Non-root (hasura:1000)
- **Architecture**: linux/amd64

### Docker Networking

**Use `--network host` when**:
- Connecting to `localhost` or `127.0.0.1`
- Hasura is in Docker Compose on same machine

**Don't use `--network host` when**:
- Connecting to external/cloud Hasura instances
- Using public HTTPS URLs

### Available Tags
- `latest` - Latest stable release
- `1.0.0` - Specific version
- `1.0` - Latest 1.0.x version
- `1` - Latest 1.x.x version

## Security Considerations

1. **Admin Secret**: Always use a strong admin secret and never commit it to version control
2. **Read-Only Mode**: Use read-only mode when exploring production systems
3. **Network Access**: Ensure the MCP server can only be accessed by authorized MCP clients
4. **Docker Isolation**: Container runs as non-root user (hasura:1000)

## Troubleshooting

### Connection Issues

**Problem**: Server can't connect to Hasura

```bash
# Test connection manually
curl -H "X-Hasura-Admin-Secret: your-secret" http://localhost:8080/v1/version
```

### Docker: "Cannot connect to localhost"

**Problem**: `dial tcp: connect: connection refused`

**Solution**: Add `--network host` when connecting to localhost:

```json
{
  "args": ["run", "-i", "--rm", "--network", "host", ...]
}
```

### MCP Client Not Finding Server

**Problem**: MCP client doesn't see the server

**Solution**: Ensure paths are absolute and Claude Desktop was restarted

### Read-Only Errors

**Problem**: "Server is in read-only mode" errors

**Solution**: Remove the `--read-only` flag if you need write access

## Architecture

```
AI Assistant (Claude/ChatGPT)
         ↓
    MCP Client
         ↓ (JSON-RPC over stdio)
    Hasura MCP Server
         ↓
cli/internal/hasura/* (API Clients)
         ↓ (HTTP + JSON)
Hasura GraphQL Engine
```

## Development

### Project Structure

```
cli/
├── cmd/
│   └── mcp-server/          # Main entry point
│       ├── main.go
│       ├── Dockerfile
│       └── README.md
└── internal/
    └── mcp/                 # MCP server implementation
        ├── server.go
        └── handlers/
```

### Running Tests

```bash
cd cli/internal/mcp
go test ./...
```

## Version History

- **v1.0.0** (2025-11-24)
  - Initial release
  - Core metadata management tools
  - Database query operations
  - Resources for metadata, version, and schema
  - Interactive prompts
  - Read-only mode support
  - Docker image

## Resources

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Hasura Documentation](https://hasura.io/docs/)
- [Docker Hub](https://hub.docker.com/r/sanjay3290/hasura-mcp)
- [MCP Go SDK](https://github.com/mark3labs/mcp-go)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

### Attribution

This MCP server is built using the Hasura GraphQL Engine CLI internal libraries:
- **Hasura GraphQL Engine**: Copyright (c) Hasura Inc. - [Apache 2.0 License](https://github.com/hasura/graphql-engine/blob/master/LICENSE)
- **MCP-Go SDK**: Copyright (c) 2024 Anthropic, PBC - [MIT License](https://github.com/mark3labs/mcp-go/blob/main/LICENSE)

This is a derivative work based on Hasura's open source code. All trademarks and product names are the property of their respective owners.

## Support

- [GitHub Issues](https://github.com/hasura/graphql-engine/issues)
- [Hasura Discord](https://hasura.io/discord)
- [Hasura Documentation](https://hasura.io/docs/)
