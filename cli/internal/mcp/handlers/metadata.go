package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hasura/graphql-engine/cli/v2/internal/errors"
	"github.com/hasura/graphql-engine/cli/v2/internal/hasura"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MetadataHandler handles metadata-related MCP operations
type MetadataHandler struct {
	client   *hasura.Client
	readOnly bool
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(client *hasura.Client, readOnly bool) *MetadataHandler {
	return &MetadataHandler{
		client:   client,
		readOnly: readOnly,
	}
}

// RegisterTools registers all metadata-related tools
func (h *MetadataHandler) RegisterTools(s *server.MCPServer) error {
	// Export metadata tool
	exportTool := mcpsdk.NewTool("export_metadata",
		mcpsdk.WithDescription("Export complete Hasura metadata from the server"),
		mcpsdk.WithString("format",
			mcpsdk.Description("Output format for metadata"),
			mcpsdk.Enum("json", "yaml"),
		),
	)
	s.AddTool(exportTool, h.HandleExportMetadata)

	// Reload metadata tool
	reloadTool := mcpsdk.NewTool("reload_metadata",
		mcpsdk.WithDescription("Reload Hasura metadata cache"),
	)
	s.AddTool(reloadTool, h.HandleReloadMetadata)

	// Get inconsistent metadata tool
	inconsistentTool := mcpsdk.NewTool("get_inconsistent_metadata",
		mcpsdk.WithDescription("Check for inconsistent metadata objects"),
	)
	s.AddTool(inconsistentTool, h.HandleGetInconsistentMetadata)

	// Only register write operations if not in read-only mode
	if !h.readOnly {
		// Apply metadata tool
		applyTool := mcpsdk.NewTool("apply_metadata",
			mcpsdk.WithDescription("Apply metadata to Hasura server"),
			mcpsdk.WithObject("metadata",
				mcpsdk.Required(),
				mcpsdk.Description("Metadata object to apply"),
			),
			mcpsdk.WithBoolean("allow_inconsistent",
				mcpsdk.Description("Allow inconsistent metadata"),
			),
		)
		s.AddTool(applyTool, h.HandleApplyMetadata)

		// Clear metadata tool
		clearTool := mcpsdk.NewTool("clear_metadata",
			mcpsdk.WithDescription("Clear all metadata from Hasura server (WARNING: destructive operation)"),
		)
		s.AddTool(clearTool, h.HandleClearMetadata)

		// Drop inconsistent metadata tool
		dropTool := mcpsdk.NewTool("drop_inconsistent_metadata",
			mcpsdk.WithDescription("Drop inconsistent metadata objects"),
		)
		s.AddTool(dropTool, h.HandleDropInconsistentMetadata)
	}

	return nil
}

// HandleExportMetadata handles the export_metadata tool call
func (h *MetadataHandler) HandleExportMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var op errors.Op = "handlers.MetadataHandler.HandleExportMetadata"

	args := request.GetArguments()
	format := "json"
	if val, ok := args["format"].(string); ok {
		format = val
	}

	// Export metadata from Hasura
	metadataReader, err := h.client.V1Metadata.ExportMetadata()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to export metadata: %v", err)), nil
	}

	// Read metadata
	metadataBytes, err := io.ReadAll(metadataReader)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to read metadata: %w", err))
	}

	// Format output
	var resultText string
	if format == "json" {
		// Pretty print JSON
		var jsonData interface{}
		if err := json.Unmarshal(metadataBytes, &jsonData); err != nil {
			return nil, errors.E(op, fmt.Errorf("failed to parse metadata: %w", err))
		}
		prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return nil, errors.E(op, fmt.Errorf("failed to format metadata: %w", err))
		}
		resultText = string(prettyJSON)
	} else {
		resultText = string(metadataBytes)
	}

	return mcpsdk.NewToolResultText(fmt.Sprintf("Successfully exported metadata:\n\n%s", resultText)), nil
}

// HandleApplyMetadata handles the apply_metadata tool call
func (h *MetadataHandler) HandleApplyMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	if h.readOnly {
		return mcpsdk.NewToolResultError("Server is in read-only mode"), nil
	}

	args := request.GetArguments()
	metadata, ok := args["metadata"]
	if !ok {
		return mcpsdk.NewToolResultError("metadata parameter is required"), nil
	}

	allowInconsistent := false
	if val, ok := args["allow_inconsistent"].(bool); ok {
		allowInconsistent = val
	}

	// Apply metadata
	replaceArgs := hasura.V2ReplaceMetadataArgs{
		AllowInconsistentMetadata: allowInconsistent,
		Metadata:                  metadata,
	}

	response, err := h.client.V1Metadata.V2ReplaceMetadata(replaceArgs)
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to apply metadata: %v", err)), nil
	}

	if !response.IsConsistent {
		inconsistentJSON, _ := json.MarshalIndent(response.InconsistentObjects, "", "  ")
		return mcpsdk.NewToolResultText(fmt.Sprintf("Metadata applied but is inconsistent:\n%s", string(inconsistentJSON))), nil
	}

	return mcpsdk.NewToolResultText("Metadata applied successfully and is consistent"), nil
}

// HandleReloadMetadata handles the reload_metadata tool call
func (h *MetadataHandler) HandleReloadMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	_, err := h.client.V1Metadata.ReloadMetadata()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to reload metadata: %v", err)), nil
	}

	return mcpsdk.NewToolResultText("Metadata cache reloaded successfully"), nil
}

// HandleClearMetadata handles the clear_metadata tool call
func (h *MetadataHandler) HandleClearMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	if h.readOnly {
		return mcpsdk.NewToolResultError("Server is in read-only mode"), nil
	}

	_, err := h.client.V1Metadata.ClearMetadata()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to clear metadata: %v", err)), nil
	}

	return mcpsdk.NewToolResultText("All metadata cleared successfully"), nil
}

// HandleGetInconsistentMetadata handles the get_inconsistent_metadata tool call
func (h *MetadataHandler) HandleGetInconsistentMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var op errors.Op = "handlers.MetadataHandler.HandleGetInconsistentMetadata"

	response, err := h.client.V1Metadata.GetInconsistentMetadata()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to get inconsistent metadata: %v", err)), nil
	}

	if response.IsConsistent {
		return mcpsdk.NewToolResultText("Metadata is consistent - no issues found"), nil
	}

	// Format inconsistent objects
	inconsistentJSON, err := json.MarshalIndent(response.InconsistentObjects, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format response: %w", err))
	}

	return mcpsdk.NewToolResultText(fmt.Sprintf("Found %d inconsistent objects:\n%s",
		len(response.InconsistentObjects), string(inconsistentJSON))), nil
}

// HandleDropInconsistentMetadata handles the drop_inconsistent_metadata tool call
func (h *MetadataHandler) HandleDropInconsistentMetadata(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	if h.readOnly {
		return mcpsdk.NewToolResultError("Server is in read-only mode"), nil
	}

	_, err := h.client.V1Metadata.DropInconsistentMetadata()
	if err != nil {
		return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to drop inconsistent metadata: %v", err)), nil
	}

	return mcpsdk.NewToolResultText("Inconsistent metadata objects dropped successfully"), nil
}
