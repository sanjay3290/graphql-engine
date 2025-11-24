package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hasura/graphql-engine/cli/v2/internal/errors"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// QueryHandler handles database query-related MCP operations
type QueryHandler struct {
	client   *hasura.Client
	readOnly bool
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(client *hasura.Client, readOnly bool) *QueryHandler {
	return &QueryHandler{
		client:   client,
		readOnly: readOnly,
	}
}

// RegisterTools registers all query-related tools
func (h *QueryHandler) RegisterTools(s *server.MCPServer) error {
	// Run SQL tool (PostgreSQL)
	runSQLTool := mcpsdk.NewTool("run_sql",
		mcpsdk.WithDescription("Execute SQL query on a PostgreSQL source"),
		mcpsdk.WithString("sql",
			mcpsdk.Required(),
			mcpsdk.Description("SQL query to execute"),
		),
		mcpsdk.WithString("source",
			mcpsdk.Description("Source name (default: 'default')"),
		),
		mcpsdk.WithBoolean("read_only",
			mcpsdk.Description("Execute in read-only mode"),
		),
		mcpsdk.WithBoolean("cascade",
			mcpsdk.Description("Cascade operations"),
		),
	)
	s.AddTool(runSQLTool, h.HandleRunSQL)

	// Get version tool
	versionTool := mcpsdk.NewTool("get_version",
		mcpsdk.WithDescription("Get Hasura GraphQL Engine version information"),
	)
	s.AddTool(versionTool, h.HandleGetVersion)

	// Get GraphQL schema tool
	schemaTool := mcpsdk.NewTool("get_schema",
		mcpsdk.WithDescription("Get GraphQL schema introspection"),
	)
	s.AddTool(schemaTool, h.HandleGetSchema)

	return nil
}

// HandleRunSQL handles the run_sql tool call
func (h *QueryHandler) HandleRunSQL(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var op errors.Op = "handlers.QueryHandler.HandleRunSQL"

	args := request.GetArguments()

	// Extract parameters
	sql, err := request.RequireString("sql")
	if err != nil {
		return mcpsdk.NewToolResultError("sql parameter is required"), nil
	}

	source := "default"
	if val, ok := args["source"].(string); ok {
		source = val
	}

	readOnly := true
	if val, ok := args["read_only"].(bool); ok {
		readOnly = val
	}

	cascade := false
	if val, ok := args["cascade"].(bool); ok {
		cascade = val
	}

	// Enforce server-level read-only mode
	if h.readOnly && !readOnly {
		return mcpsdk.NewToolResultError("Server is in read-only mode. Cannot execute write operations."), nil
	}

	// Prepare input
	input := hasura.PGRunSQLInput{
		SQL:      sql,
		Source:   source,
		ReadOnly: readOnly,
		Cascade:  cascade,
	}

	// Execute SQL
	output, err := h.client.V2Query.PGRunSQL(input)
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to execute SQL: %v", err)), nil
	}

	// Format result
	resultJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format result: %w", err))
	}

	return mcpsdk.NewToolResultText(fmt.Sprintf("SQL executed successfully:\n\nResult Type: %s\n\nResult:\n%s",
		output.ResultType, string(resultJSON))), nil
}

// HandleGetVersion handles the get_version tool call
func (h *QueryHandler) HandleGetVersion(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var op errors.Op = "handlers.QueryHandler.HandleGetVersion"

	version, err := h.client.V1Version.GetVersion()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to get version: %v", err)), nil
	}

	versionJSON, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format version: %w", err))
	}

	return mcpsdk.NewToolResultText(fmt.Sprintf("Hasura version information:\n%s", string(versionJSON))), nil
}

// HandleGetSchema handles the get_schema tool call
func (h *QueryHandler) HandleGetSchema(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var op errors.Op = "handlers.QueryHandler.HandleGetSchema"

	schema, err := h.client.V1Graphql.GetIntrospectionSchema()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to get schema: %v", err)), nil
	}

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format schema: %w", err))
	}

	return mcpsdk.NewToolResultText(fmt.Sprintf("GraphQL Schema:\n%s", string(schemaJSON))), nil
}
