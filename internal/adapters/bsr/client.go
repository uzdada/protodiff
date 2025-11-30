// Package bsr provides clients for interacting with the Buf Schema Registry (BSR).
//
// BSR is the source of truth for protobuf schemas. This package provides both
// a production HTTP client and a mock client for testing.
//
// The Client interface allows for easy swapping between implementations:
//   - HTTPClient: Production client that makes real HTTP requests to BSR API
//   - MockClient: Test client with pre-configured schemas
//
// Example usage:
//
//	client := bsr.NewHTTPClient()
//	schema, err := client.FetchSchema(ctx, "buf.build/acme/user")
package bsr

import (
	"context"

	"github.com/uzdada/protodiff/internal/core/domain"
)

// Client defines the interface for interacting with Buf Schema Registry
type Client interface {
	// FetchSchema retrieves the schema definition from BSR for a given module
	FetchSchema(ctx context.Context, module string) (*domain.SchemaDescriptor, error)
}
