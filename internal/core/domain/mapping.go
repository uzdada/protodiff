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
