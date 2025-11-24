package handlers

import (
	"context"
	"fmt"

	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PromptHandler handles MCP prompt operations
type PromptHandler struct{}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler() *PromptHandler {
	return &PromptHandler{}
}

// RegisterPrompts registers all MCP prompts
func (h *PromptHandler) RegisterPrompts(s *server.MCPServer) error {
	// Setup table prompt
	setupTablePrompt := mcpsdk.NewPrompt("setup_table",
		mcpsdk.WithPromptDescription("Interactive guide for tracking and configuring a table in Hasura"),
		mcpsdk.WithArgument("table_name",
			mcpsdk.ArgumentDescription("Name of the table to track"),
			mcpsdk.RequiredArgument(),
		),
		mcpsdk.WithArgument("schema",
			mcpsdk.ArgumentDescription("Database schema name (e.g., 'public')"),
		),
		mcpsdk.WithArgument("source",
			mcpsdk.ArgumentDescription("Source name (default: 'default')"),
		),
	)
	s.AddPrompt(setupTablePrompt, h.HandleSetupTablePrompt)

	// Debug metadata prompt
	debugMetadataPrompt := mcpsdk.NewPrompt("debug_metadata",
		mcpsdk.WithPromptDescription("Diagnostic guide for troubleshooting metadata issues"),
	)
	s.AddPrompt(debugMetadataPrompt, h.HandleDebugMetadataPrompt)

	// Create migration prompt
	createMigrationPrompt := mcpsdk.NewPrompt("create_migration",
		mcpsdk.WithPromptDescription("Template for creating database migrations"),
		mcpsdk.WithArgument("operation",
			mcpsdk.ArgumentDescription("Type of migration (e.g., 'add_column', 'create_table')"),
			mcpsdk.RequiredArgument(),
		),
		mcpsdk.WithArgument("source",
			mcpsdk.ArgumentDescription("Source name (default: 'default')"),
		),
	)
	s.AddPrompt(createMigrationPrompt, h.HandleCreateMigrationPrompt)

	// Setup permissions prompt
	setupPermissionsPrompt := mcpsdk.NewPrompt("setup_permissions",
		mcpsdk.WithPromptDescription("Guide for configuring table permissions and roles"),
		mcpsdk.WithArgument("table",
			mcpsdk.ArgumentDescription("Table name"),
			mcpsdk.RequiredArgument(),
		),
		mcpsdk.WithArgument("role",
			mcpsdk.ArgumentDescription("Role name (e.g., 'user', 'admin')"),
			mcpsdk.RequiredArgument(),
		),
	)
	s.AddPrompt(setupPermissionsPrompt, h.HandleSetupPermissionsPrompt)

	return nil
}

// HandleSetupTablePrompt handles the setup_table prompt
func (h *PromptHandler) HandleSetupTablePrompt(ctx context.Context, request mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	args := request.Params.Arguments

	tableName := ""
	if val, ok := args["table_name"]; ok {
		tableName = val
	}

	schema := ""
	if val, ok := args["schema"]; ok {
		schema = val
	}

	source := ""
	if val, ok := args["source"]; ok {
		source = val
	}

	if schema == "" {
		schema = "public"
	}
	if source == "" {
		source = "default"
	}

	promptText := fmt.Sprintf(`I need to track and configure a table in Hasura with the following details:

Table: %s
Schema: %s
Source: %s

Please help me with the following steps:

1. **Track the table**: Generate the metadata configuration to track this table
2. **Set up relationships**: Identify potential foreign keys and create relationships
3. **Configure permissions**: Set up appropriate select/insert/update/delete permissions
4. **Add computed fields** (if needed): Create any computed fields based on the table structure

For each step, provide the necessary commands or metadata changes I need to make.`, tableName, schema, source)

	return &mcpsdk.GetPromptResult{
		Description: fmt.Sprintf("Setup guide for table '%s'", tableName),
		Messages: []mcpsdk.PromptMessage{
			{
				Role: "user",
				Content: mcpsdk.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// HandleDebugMetadataPrompt handles the debug_metadata prompt
func (h *PromptHandler) HandleDebugMetadataPrompt(ctx context.Context, request mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	promptText := `I'm experiencing issues with my Hasura metadata. Please help me diagnose and fix the problems.

Run through this diagnostic checklist:

1. **Check for inconsistent metadata**:
   - Use the 'get_inconsistent_metadata' tool to identify any inconsistent objects
   - Explain what each inconsistency means and how to fix it

2. **Verify metadata structure**:
   - Use 'export_metadata' to review the current metadata
   - Check for common issues like missing tables, broken relationships, or invalid permissions

3. **Check server connectivity**:
   - Use 'get_version' to verify the server is accessible
   - Confirm the admin secret is valid

4. **Recommend fixes**:
   - Provide specific commands or metadata changes to resolve issues
   - Suggest whether to use 'reload_metadata' or 'drop_inconsistent_metadata'

Please provide a detailed analysis and actionable steps to resolve any issues found.`

	return &mcpsdk.GetPromptResult{
		Description: "Metadata debugging guide",
		Messages: []mcpsdk.PromptMessage{
			{
				Role: "user",
				Content: mcpsdk.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// HandleCreateMigrationPrompt handles the create_migration prompt
func (h *PromptHandler) HandleCreateMigrationPrompt(ctx context.Context, request mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	args := request.Params.Arguments

	operation := ""
	if val, ok := args["operation"]; ok {
		operation = val
	}

	source := ""
	if val, ok := args["source"]; ok {
		source = val
	}

	if source == "" {
		source = "default"
	}

	promptText := fmt.Sprintf(`I need to create a database migration for the following operation:

Operation: %s
Source: %s

Please help me:

1. **Generate SQL for up migration**: Provide the SQL statements to perform this operation
2. **Generate SQL for down migration**: Provide the SQL statements to rollback this operation
3. **Explain the impact**: Describe what this migration will do and any potential risks
4. **Suggest best practices**: Recommend any additional steps (e.g., adding indexes, updating permissions)

For reference, I can execute SQL using the 'run_sql' tool with the generated statements.`, operation, source)

	return &mcpsdk.GetPromptResult{
		Description: fmt.Sprintf("Migration template for '%s'", operation),
		Messages: []mcpsdk.PromptMessage{
			{
				Role: "user",
				Content: mcpsdk.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// HandleSetupPermissionsPrompt handles the setup_permissions prompt
func (h *PromptHandler) HandleSetupPermissionsPrompt(ctx context.Context, request mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	args := request.Params.Arguments

	table := ""
	if val, ok := args["table"]; ok {
		table = val
	}

	role := ""
	if val, ok := args["role"]; ok {
		role = val
	}

	promptText := fmt.Sprintf(`I need to configure permissions for a Hasura table:

Table: %s
Role: %s

Please help me set up appropriate permissions:

1. **Select permissions**: Define which rows and columns this role can read
2. **Insert permissions**: Define what data this role can insert and any validation rules
3. **Update permissions**: Define which rows and columns this role can update
4. **Delete permissions**: Define which rows this role can delete

For each permission type:
- Provide the metadata configuration
- Explain the security implications
- Suggest appropriate filter conditions (e.g., {"user_id": {"_eq": "X-Hasura-User-Id"}})
- Recommend any column-level restrictions

Generate the complete metadata configuration for these permissions.`, table, role)

	return &mcpsdk.GetPromptResult{
		Description: fmt.Sprintf("Permission setup for role '%s' on table '%s'", role, table),
		Messages: []mcpsdk.PromptMessage{
			{
				Role: "user",
				Content: mcpsdk.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}
