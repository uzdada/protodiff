// Copyright 2025 ProtoDiff Authors
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

// Package bsr provides clients for interacting with the Buf Schema Registry (BSR).
//
// BSR is the source of truth for protobuf schemas. This package provides both
// a production HTTP client and a mock client for testing.
//
// The Client interface allows for easy swapping between implementations:
//   - HTTPClient: Production client that makes real HTTP requests to BSR API
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
