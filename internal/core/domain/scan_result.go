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
	// LastChecked is the timestamp of the last validation
	LastChecked time.Time `json:"last_checked"`
	// PodIP is the IP address used for gRPC reflection
	PodIP string `json:"pod_ip"`
	// GRPCPort is the port used for gRPC reflection
	GRPCPort int32 `json:"grpc_port"`
}
