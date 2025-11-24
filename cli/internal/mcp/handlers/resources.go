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

// ResourceHandler handles MCP resource operations
type ResourceHandler struct {
	client *hasura.Client
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(client *hasura.Client) *ResourceHandler {
	return &ResourceHandler{
		client: client,
	}
}

// RegisterResources registers all MCP resources
func (h *ResourceHandler) RegisterResources(s *server.MCPServer) error {
	// Metadata resource
	metadataResource := mcpsdk.NewResource(
		"hasura://metadata",
		"Current Hasura metadata",
		mcpsdk.WithResourceDescription("Complete metadata export from the Hasura server"),
		mcpsdk.WithMIMEType("application/json"),
	)
	s.AddResource(metadataResource, h.HandleMetadataResource)

	// Version resource
	versionResource := mcpsdk.NewResource(
		"hasura://version",
		"Hasura server version",
		mcpsdk.WithResourceDescription("Version information for the connected Hasura server"),
		mcpsdk.WithMIMEType("application/json"),
	)
	s.AddResource(versionResource, h.HandleVersionResource)

	// Schema resource
	schemaResource := mcpsdk.NewResource(
		"hasura://schema",
		"GraphQL schema",
		mcpsdk.WithResourceDescription("Complete GraphQL schema introspection"),
		mcpsdk.WithMIMEType("application/json"),
	)
	s.AddResource(schemaResource, h.HandleSchemaResource)

	// Inconsistent metadata resource
	inconsistentResource := mcpsdk.NewResource(
		"hasura://inconsistent-metadata",
		"Inconsistent metadata objects",
		mcpsdk.WithResourceDescription("List of metadata objects that are inconsistent"),
		mcpsdk.WithMIMEType("application/json"),
	)
	s.AddResource(inconsistentResource, h.HandleInconsistentMetadataResource)

	return nil
}

// HandleMetadataResource handles requests for the metadata resource
func (h *ResourceHandler) HandleMetadataResource(ctx context.Context, request mcpsdk.ReadResourceRequest) ([]mcpsdk.ResourceContents, error) {
	var op errors.Op = "handlers.ResourceHandler.HandleMetadataResource"

	metadataReader, err := h.client.V1Metadata.ExportMetadata()
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to export metadata: %w", err))
	}

	metadataBytes, err := io.ReadAll(metadataReader)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to read metadata: %w", err))
	}

	// Pretty print JSON
	var jsonData interface{}
	if err := json.Unmarshal(metadataBytes, &jsonData); err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to parse metadata: %w", err))
	}
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format metadata: %w", err))
	}

	return []mcpsdk.ResourceContents{
		mcpsdk.TextResourceContents{
			URI:      request.Params.URI,
			Text:     string(prettyJSON),
			MIMEType: "application/json",
		},
	}, nil
}

// HandleVersionResource handles requests for the version resource
func (h *ResourceHandler) HandleVersionResource(ctx context.Context, request mcpsdk.ReadResourceRequest) ([]mcpsdk.ResourceContents, error) {
	var op errors.Op = "handlers.ResourceHandler.HandleVersionResource"

	version, err := h.client.V1Version.GetVersion()
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to get version: %w", err))
	}

	versionJSON, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format version: %w", err))
	}

	return []mcpsdk.ResourceContents{
		mcpsdk.TextResourceContents{
			URI:      request.Params.URI,
			Text:     string(versionJSON),
			MIMEType: "application/json",
		},
	}, nil
}

// HandleSchemaResource handles requests for the schema resource
func (h *ResourceHandler) HandleSchemaResource(ctx context.Context, request mcpsdk.ReadResourceRequest) ([]mcpsdk.ResourceContents, error) {
	var op errors.Op = "handlers.ResourceHandler.HandleSchemaResource"

	schema, err := h.client.V1Graphql.GetIntrospectionSchema()
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to get schema: %w", err))
	}

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format schema: %w", err))
	}

	return []mcpsdk.ResourceContents{
		mcpsdk.TextResourceContents{
			URI:      request.Params.URI,
			Text:     string(schemaJSON),
			MIMEType: "application/json",
		},
	}, nil
}

// HandleInconsistentMetadataResource handles requests for inconsistent metadata
func (h *ResourceHandler) HandleInconsistentMetadataResource(ctx context.Context, request mcpsdk.ReadResourceRequest) ([]mcpsdk.ResourceContents, error) {
	var op errors.Op = "handlers.ResourceHandler.HandleInconsistentMetadataResource"

	response, err := h.client.V1Metadata.GetInconsistentMetadata()
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to get inconsistent metadata: %w", err))
	}

	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("failed to format response: %w", err))
	}

	return []mcpsdk.ResourceContents{
		mcpsdk.TextResourceContents{
			URI:      request.Params.URI,
			Text:     string(resultJSON),
			MIMEType: "application/json",
		},
	}, nil
}
