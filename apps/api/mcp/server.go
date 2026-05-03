package mcp

import (
	"context"

	"4ks/apps/api/app"
	"4ks/apps/api/utils"
)

// Server is a placeholder for the future MCP transport and tool handlers.
type Server struct{}

// New wires the MCP server stub.
func New(_ *utils.RuntimeConfig, _ app.Services) *Server {
	return &Server{}
}

// Start blocks until shutdown until MCP tool handlers are implemented.
func (s *Server) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
