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
