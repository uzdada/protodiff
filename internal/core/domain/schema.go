package domain

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
