// Package grpc provides a client for fetching schemas from live gRPC services
// using server reflection.
//
// Server reflection (https://github.com/grpc/grpc/blob/master/doc/server-reflection.md)
// allows gRPC clients to discover service definitions at runtime without needing
// proto files. This package leverages reflection to fetch schemas from running pods.
//
// The ReflectionClient connects to a gRPC server, queries available services,
// and converts the discovered schema into domain.SchemaDescriptor format.
//
// Example usage:
//
//	client := grpc.NewReflectionClient()
//	schema, err := client.FetchSchema(ctx, "10.0.1.5:9090")
package grpc

import (
	"context"
	"fmt"

	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/uzdada/protodiff/internal/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const (
	// reflectionServiceName is the gRPC reflection service to be skipped
	// when listing services, as it's a meta-service not part of the schema
	reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"
)

// ReflectionClient provides gRPC server reflection capabilities
type ReflectionClient struct{}

// NewReflectionClient creates a new gRPC reflection client
func NewReflectionClient() *ReflectionClient {
	return &ReflectionClient{}
}

// FetchSchema retrieves the schema from a gRPC server using reflection
func (r *ReflectionClient) FetchSchema(ctx context.Context, address string) (*domain.SchemaDescriptor, error) {
	// Connect to the gRPC server
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	// Create reflection client using the connection
	refClient := grpcreflect.NewClientV1Alpha(ctx, reflectpb.NewServerReflectionClient(conn))
	defer refClient.Reset()

	// List all services
	services, err := refClient.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	schema := &domain.SchemaDescriptor{
		Services: make([]domain.ServiceDescriptor, 0),
		Messages: make([]string, 0),
	}

	// Extract service and method information
	for _, serviceName := range services {
		// Skip the reflection service itself
		if serviceName == reflectionServiceName {
			continue
		}

		serviceDesc, err := refClient.ResolveService(serviceName)
		if err != nil {
			continue // Skip services we can't resolve
		}

		methods := make([]string, 0)
		for _, method := range serviceDesc.GetMethods() {
			methods = append(methods, method.GetName())
		}

		schema.Services = append(schema.Services, domain.ServiceDescriptor{
			Name:    serviceName,
			Methods: methods,
		})
	}

	return schema, nil
}
