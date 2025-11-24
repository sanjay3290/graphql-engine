package mcp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hasura/graphql-engine/cli/v2"
	"github.com/hasura/graphql-engine/cli/v2/internal/errors"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/pgdump"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/v1graphql"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/v1metadata"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/v1query"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/v1version"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura/v2query"
	"github.com/hasura/graphql-engine/cli/v2/internal/httpc"
	"github.com/hasura/graphql-engine/cli/v2/internal/mcp/handlers"
	"github.com/mark3labs/mcp-go/server"
)

// HasuraMCPServer wraps the MCP server with Hasura client integration
type HasuraMCPServer struct {
	server       *server.MCPServer
	hasuraClient *hasura.Client
	ec           *cli.ExecutionContext
	config       *ServerConfig
}

// ServerConfig holds configuration for the MCP server
type ServerConfig struct {
	Endpoint    string
	AdminSecret string
	ReadOnly    bool
	ServerName  string
	Version     string
}

// NewHasuraMCPServer creates a new Hasura MCP server instance
func NewHasuraMCPServer(config *ServerConfig) (*HasuraMCPServer, error) {
	var op errors.Op = "mcp.NewHasuraMCPServer"

	if config == nil {
		return nil, errors.E(op, "config cannot be nil")
	}

	if config.Endpoint == "" {
		return nil, errors.E(op, "endpoint is required")
	}

	if config.ServerName == "" {
		config.ServerName = "hasura-mcp-server"
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(config.ServerName, config.Version)

	// Initialize Hasura client through ExecutionContext
	ec, err := initializeExecutionContext(config)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to initialize execution context: %w", err))
	}

	hasuraServer := &HasuraMCPServer{
		server:       mcpServer,
		hasuraClient: ec.APIClient,
		ec:           ec,
		config:       config,
	}

	// Register all capabilities
	if err := hasuraServer.registerCapabilities(); err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to register capabilities: %w", err))
	}

	return hasuraServer, nil
}

// registerCapabilities registers all tools, resources, and prompts
func (s *HasuraMCPServer) registerCapabilities() error {
	var op errors.Op = "mcp.HasuraMCPServer.registerCapabilities"

	// Initialize handlers
	metadataHandler := handlers.NewMetadataHandler(s.hasuraClient, s.config.ReadOnly)
	queryHandler := handlers.NewQueryHandler(s.hasuraClient, s.config.ReadOnly)
	resourceHandler := handlers.NewResourceHandler(s.hasuraClient)
	promptHandler := handlers.NewPromptHandler()

	// Register metadata tools
	if err := metadataHandler.RegisterTools(s.server); err != nil {
		return errors.E(op, fmt.Errorf("failed to register metadata tools: %w", err))
	}

	// Register query tools
	if err := queryHandler.RegisterTools(s.server); err != nil {
		return errors.E(op, fmt.Errorf("failed to register query tools: %w", err))
	}

	// Register resources
	if err := resourceHandler.RegisterResources(s.server); err != nil {
		return errors.E(op, fmt.Errorf("failed to register resources: %w", err))
	}

	// Register prompts
	if err := promptHandler.RegisterPrompts(s.server); err != nil {
		return errors.E(op, fmt.Errorf("failed to register prompts: %w", err))
	}

	return nil
}

// Serve starts the MCP server using stdio transport
func (s *HasuraMCPServer) Serve() error {
	var op errors.Op = "mcp.HasuraMCPServer.Serve"

	if err := server.ServeStdio(s.server); err != nil {
		return errors.E(op, fmt.Errorf("failed to serve: %w", err))
	}

	return nil
}

// GetServerInfo returns information about the MCP server
func (s *HasuraMCPServer) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	var op errors.Op = "mcp.HasuraMCPServer.GetServerInfo"

	// Get Hasura version
	versionResp, err := s.hasuraClient.V1Version.GetVersion()
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to get Hasura version: %w", err))
	}

	return &ServerInfo{
		MCPServerName:    s.config.ServerName,
		MCPServerVersion: s.config.Version,
		HasuraEndpoint:   s.config.Endpoint,
		HasuraVersion:    versionResp.Version,
		ReadOnlyMode:     s.config.ReadOnly,
	}, nil
}

// ServerInfo contains information about the MCP server and Hasura instance
type ServerInfo struct {
	MCPServerName    string `json:"mcp_server_name"`
	MCPServerVersion string `json:"mcp_server_version"`
	HasuraEndpoint   string `json:"hasura_endpoint"`
	HasuraVersion    string `json:"hasura_version"`
	ReadOnlyMode     bool   `json:"read_only_mode"`
}

// initializeExecutionContext creates a CLI execution context for Hasura operations
func initializeExecutionContext(config *ServerConfig) (*cli.ExecutionContext, error) {
	var op errors.Op = "mcp.initializeExecutionContext"

	ec := &cli.ExecutionContext{}

	// Set basic configuration
	parsedURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to parse endpoint: %w", err))
	}

	apiPaths := &cli.ServerAPIPaths{
		V1Query:    "/v1/query",
		V2Query:    "/v2/query",
		V1Metadata: "/v1/metadata",
		GraphQL:    "/v1/graphql",
		Config:     "/v1alpha1/config",
		PGDump:     "/v1alpha1/pg_dump",
		Version:    "/v1/version",
	}

	ec.Config = &cli.Config{
		Version: cli.V3, // Use latest config version
		ServerConfig: cli.ServerConfig{
			Endpoint:       config.Endpoint,
			AdminSecret:    config.AdminSecret,
			ParsedEndpoint: parsedURL,
			APIPaths:       apiPaths,
		},
	}

	// Initialize HTTP client
	httpClient, err := initializeHTTPClient(config)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to initialize HTTP client: %w", err))
	}

	ec.Config.HTTPClient = httpClient

	// Initialize API client directly
	ec.APIClient = initializeAPIClient(httpClient, ec.Config)

	return ec, nil
}

// initializeHTTPClient creates an HTTP client for Hasura API communication
func initializeHTTPClient(config *ServerConfig) (*httpc.Client, error) {
	var op errors.Op = "mcp.initializeHTTPClient"

	// Create standard HTTP client
	standardHTTPClient, err := httpc.NewHttpClientWithTLSConfig(nil)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to create HTTP client: %w", err))
	}

	// Set up headers
	headers := make(map[string]string)
	if config.AdminSecret != "" {
		headers[cli.XHasuraAdminSecret] = config.AdminSecret
	}

	// Create httpc.Client
	httpClient, err := httpc.New(standardHTTPClient, config.Endpoint, headers)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to create HTTP client wrapper: %w", err))
	}

	return httpClient, nil
}

// initializeAPIClient creates the Hasura API client with all endpoints
func initializeAPIClient(httpClient *httpc.Client, config *cli.Config) *hasura.Client {
	return &hasura.Client{
		V1Metadata: v1metadata.New(httpClient, config.GetV1MetadataEndpoint()),
		V1Query:    v1query.New(httpClient, config.GetV1QueryEndpoint()),
		V2Query:    v2query.New(httpClient, config.GetV2QueryEndpoint()),
		PGDump:     pgdump.New(httpClient, config.GetPGDumpEndpoint()),
		V1Graphql:  v1graphql.New(httpClient, config.GetV1GraphqlEndpoint()),
		V1Version:  v1version.New(httpClient, config.GetVersionEndpoint()),
	}
}
