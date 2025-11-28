package domain

import "time"

// DiffStatus represents the schema comparison status between live pod and BSR
type DiffStatus string

const (
	// StatusSync indicates the schemas are in sync
	StatusSync DiffStatus = "SYNC"
	// StatusMismatch indicates schema drift has been detected
	StatusMismatch DiffStatus = "MISMATCH"
	// StatusUnknown indicates the status could not be determined
	StatusUnknown DiffStatus = "UNKNOWN"
)

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
	// LastChecked is the timestamp of the last validation
	LastChecked time.Time `json:"last_checked"`
	// PodIP is the IP address used for gRPC reflection
	PodIP string `json:"pod_ip"`
	// GRPCPort is the port used for gRPC reflection
	GRPCPort int32 `json:"grpc_port"`
}

// ServiceMapping represents the configuration mapping between a service and its BSR module
type ServiceMapping struct {
	// ServiceName is the logical service identifier
	ServiceName string
	// BSRModule is the fully qualified BSR module path (e.g., buf.build/acme/user)
	BSRModule string
}

// SchemaDescriptor represents a protobuf schema definition
type SchemaDescriptor struct {
	// Services is a list of gRPC service definitions
	Services []ServiceDescriptor `json:"services"`
	// Messages is a list of message type definitions
	Messages []string `json:"messages"`
}

// ServiceDescriptor represents a single gRPC service definition
type ServiceDescriptor struct {
	// Name is the fully qualified service name
	Name string `json:"name"`
	// Methods is a list of RPC method names
	Methods []string `json:"methods"`
}
