package domain

// SchemaDescriptor represents a protobuf schema definition
type SchemaDescriptor struct {
	// Services is a list of gRPC service definitions
	Services []ServiceDescriptor `json:"services"`
	// Messages is a list of message type definitions with their fields
	Messages []MessageDescriptor `json:"messages"`
}

// ServiceDescriptor represents a single gRPC service definition
type ServiceDescriptor struct {
	// Name is the fully qualified service name
	Name string `json:"name"`
	// Methods is a list of RPC method names
	Methods []string `json:"methods"`
}

// MessageDescriptor represents a protobuf message type
type MessageDescriptor struct {
	// Name is the fully qualified message name
	Name string `json:"name"`
	// Fields is a list of message fields
	Fields []FieldDescriptor `json:"fields"`
}

// FieldDescriptor represents a single field in a message
type FieldDescriptor struct {
	// Name is the field name
	Name string `json:"name"`
	// Type is the field type (e.g., "string", "int32", "MyMessage")
	Type string `json:"type"`
	// Number is the field number in the proto definition
	Number int32 `json:"number"`
	// IsRepeated indicates if this is a repeated field
	IsRepeated bool `json:"is_repeated"`
}
