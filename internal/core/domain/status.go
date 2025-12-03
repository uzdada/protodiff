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
