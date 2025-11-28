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
