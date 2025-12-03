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

// Package domain contains the core domain models for ProtoDiff.
//
// This package defines the business entities and value objects used throughout
// the application, following Domain-Driven Design principles:
//   - ScanResult: Schema validation results for a pod
//   - SchemaDescriptor: Protobuf schema definitions
//   - ServiceMappings: Service-to-BSR module mappings
//   - DiffStatus: Schema comparison status enumeration
//
// All types in this package are framework-agnostic and represent pure business
// logic without dependencies on infrastructure concerns (HTTP, Kubernetes, etc).
package domain

// ServiceMapping represents the configuration mapping between a service and its BSR module
type ServiceMapping struct {
	// ServiceName is the logical service identifier
	ServiceName string
	// BSRModule is the fully qualified BSR module path (e.g., buf.build/acme/user)
	BSRModule string
}

// ServiceMappings is a collection of service-to-BSR module mappings
// It provides a type-safe wrapper around the underlying map structure
type ServiceMappings struct {
	mappings map[string]string
}

// NewServiceMappings creates a new ServiceMappings from a raw map
func NewServiceMappings(data map[string]string) ServiceMappings {
	if data == nil {
		data = make(map[string]string)
	}
	return ServiceMappings{mappings: data}
}

// Get retrieves the BSR module for a given service name
func (sm ServiceMappings) Get(serviceName string) (string, bool) {
	module, exists := sm.mappings[serviceName]
	return module, exists
}

// GetAll returns all mappings as a slice of ServiceMapping structs
func (sm ServiceMappings) GetAll() []ServiceMapping {
	result := make([]ServiceMapping, 0, len(sm.mappings))
	for serviceName, bsrModule := range sm.mappings {
		result = append(result, ServiceMapping{
			ServiceName: serviceName,
			BSRModule:   bsrModule,
		})
	}
	return result
}

// Count returns the number of mappings
func (sm ServiceMappings) Count() int {
	return len(sm.mappings)
}

// Has checks if a mapping exists for the given service name
func (sm ServiceMappings) Has(serviceName string) bool {
	_, exists := sm.mappings[serviceName]
	return exists
}

// GetServiceNames returns all service names from the mappings
func (sm ServiceMappings) GetServiceNames() []string {
	names := make([]string, 0, len(sm.mappings))
	for serviceName := range sm.mappings {
		names = append(names, serviceName)
	}
	return names
}
