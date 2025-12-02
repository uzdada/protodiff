package domain

import "time"

// ScanResult represents the validation result for a single pod
type ScanResult struct {
	// PodName is the Kubernetes pod name
	PodName string `json:"pod_name"`
	// PodNamespace is the Kubernetes namespace
	PodNamespace string `json:"pod_namespace"`
	// ServiceName is the logical service name (from labels)
	ServiceName string `json:"service_name"`
	// BSRModule is the Buf Schema Registry module reference
	BSRModule string `json:"bsr_module"`
	// Status indicates the drift detection status
	Status DiffStatus `json:"status"`
	// Message provides additional context (error message, etc.)
	Message string `json:"message,omitempty"`
	// SchemaDiff contains detailed diff information
	SchemaDiff *SchemaDiff `json:"schema_diff,omitempty"`
	// LastChecked is the timestamp of the last validation
	LastChecked time.Time `json:"last_checked"`
	// PodIP is the IP address used for gRPC reflection
	PodIP string `json:"pod_ip"`
	// GRPCPort is the port used for gRPC reflection
	GRPCPort int32 `json:"grpc_port"`
}

// SchemaDiff contains detailed diff information between live and BSR schemas
type SchemaDiff struct {
	// LiveServices are the services found in the live pod
	LiveServices []string `json:"live_services,omitempty"`
	// BSRServices are the services defined in BSR
	BSRServices []string `json:"bsr_services,omitempty"`
	// MissingInLive are services in BSR but not in live pod
	MissingInLive []string `json:"missing_in_live,omitempty"`
	// ExtraInLive are services in live pod but not in BSR
	ExtraInLive []string `json:"extra_in_live,omitempty"`
	// MethodMismatches are services with different methods
	MethodMismatches []ServiceMethodMismatch `json:"method_mismatches,omitempty"`
	// MatchedServices are services with matching methods
	MatchedServices []ServiceMethodMatch `json:"matched_services,omitempty"`
	// MessageFieldMismatches are messages with different fields
	MessageFieldMismatches []MessageFieldMismatch `json:"message_field_mismatches,omitempty"`
	// MatchedMessages are messages with matching fields
	MatchedMessages []MessageFieldMatch `json:"matched_messages,omitempty"`
}

// ServiceMethodMismatch represents a method count mismatch for a service
type ServiceMethodMismatch struct {
	ServiceName    string   `json:"service_name"`
	LiveMethods    int      `json:"live_methods"`
	BSRMethods     int      `json:"bsr_methods"`
	MissingMethods []string `json:"missing_methods,omitempty"`
	ExtraMethods   []string `json:"extra_methods,omitempty"`
}

// ServiceMethodMatch represents a service with matching methods
type ServiceMethodMatch struct {
	ServiceName string   `json:"service_name"`
	Methods     []string `json:"methods"`
}

// MessageFieldMismatch represents a message with field differences
type MessageFieldMismatch struct {
	MessageName    string              `json:"message_name"`
	LiveFields     int                 `json:"live_fields"`
	BSRFields      int                 `json:"bsr_fields"`
	MissingFields  []FieldDescriptor   `json:"missing_fields,omitempty"`
	ExtraFields    []FieldDescriptor   `json:"extra_fields,omitempty"`
	ModifiedFields []FieldModification `json:"modified_fields,omitempty"`
}

// FieldModification represents a field that exists in both but has different properties
type FieldModification struct {
	FieldName  string `json:"field_name"`
	LiveType   string `json:"live_type"`
	BSRType    string `json:"bsr_type"`
	LiveNumber int32  `json:"live_number"`
	BSRNumber  int32  `json:"bsr_number"`
	ChangeType string `json:"change_type"` // "type_changed", "number_changed", "repeated_changed"
}

// MessageFieldMatch represents a message with matching fields
type MessageFieldMatch struct {
	MessageName string            `json:"message_name"`
	Fields      []FieldDescriptor `json:"fields"`
}
