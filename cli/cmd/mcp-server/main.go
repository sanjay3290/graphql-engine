// Copyright 2025 Sanjay Ramadugu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of a derivative work based on Hasura GraphQL Engine.
// Original work Copyright (c) Hasura Inc.

package main

import (
	"fmt"
	"os"

	"github.com/hasura/graphql-engine/cli/v2/internal/mcp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	endpoint    string
	adminSecret string
	readOnly    bool
	version     bool
	serverName  string
	versionStr  string = "1.0.0"
)

func init() {
	pflag.StringVarP(&endpoint, "endpoint", "e", "", "Hasura GraphQL Engine endpoint (e.g., http://localhost:8080)")
	pflag.StringVarP(&adminSecret, "admin-secret", "s", "", "Hasura admin secret")
	pflag.BoolVarP(&readOnly, "read-only", "r", false, "Run server in read-only mode (prevents destructive operations)")
	pflag.BoolVarP(&version, "version", "v", false, "Print version information")
	pflag.StringVarP(&serverName, "name", "n", "hasura-mcp-server", "MCP server name")
}

func main() {
	pflag.Parse()

	// Print version and exit
	if version {
		fmt.Printf("Hasura MCP Server v%s\n", versionStr)
		os.Exit(0)
	}

	// Check required flags
	if endpoint == "" {
		endpoint = os.Getenv("HASURA_GRAPHQL_ENDPOINT")
		if endpoint == "" {
			logrus.Fatal("Error: --endpoint is required or set HASURA_GRAPHQL_ENDPOINT environment variable")
		}
	}

	if adminSecret == "" {
		adminSecret = os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")
		if adminSecret == "" {
			logrus.Warn("Warning: No admin secret provided. Set --admin-secret or HASURA_GRAPHQL_ADMIN_SECRET")
		}
	}

	// Create server configuration
	config := &mcp.ServerConfig{
		Endpoint:    endpoint,
		AdminSecret: adminSecret,
		ReadOnly:    readOnly,
		ServerName:  serverName,
		Version:     versionStr,
	}

	// Log startup information
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.WithFields(logrus.Fields{
		"endpoint":  endpoint,
		"read_only": readOnly,
		"name":      serverName,
		"version":   versionStr,
	}).Info("Starting Hasura MCP Server")

	// Create and start MCP server
	server, err := mcp.NewHasuraMCPServer(config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create MCP server")
	}

	// Log server capabilities
	logrus.Info("MCP server initialized successfully")
	logrus.Info("Registered capabilities:")
	logrus.Info("  Tools: export_metadata, apply_metadata, reload_metadata, clear_metadata, get_inconsistent_metadata, drop_inconsistent_metadata, run_sql, get_version, get_schema")
	logrus.Info("  Resources: hasura://metadata, hasura://version, hasura://schema, hasura://inconsistent-metadata")
	logrus.Info("  Prompts: setup_table, debug_metadata, create_migration, setup_permissions")

	if readOnly {
		logrus.Warn("Running in READ-ONLY mode - destructive operations are disabled")
	}

	logrus.Info("Server ready - listening on stdio")

	// Serve using stdio transport
	if err := server.Serve(); err != nil {
		logrus.WithError(err).Fatal("Server error")
	}
}
