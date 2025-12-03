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

// Package store provides thread-safe in-memory storage for scan results.
//
// The store maintains a map of scan results keyed by pod identifier
// (namespace/name). It uses a read-write mutex to ensure safe concurrent
// access from multiple goroutines.
//
// The store supports:
//   - Setting results (upsert operation)
//   - Getting a specific result by pod identifier
//   - Retrieving all results
//   - Deleting results
//   - Counting total stored results
//
// Example usage:
//
//	store := store.New()
//	store.Set(scanResult)
//	result, exists := store.Get("default", "my-pod")
package store

import (
	"sync"

	"github.com/uzdada/protodiff/internal/core/domain"
)

// Store provides thread-safe in-memory storage for scan results
type Store struct {
	mu      sync.RWMutex
	results map[string]*domain.ScanResult // key: podNamespace/podName
}

// New creates a new Store instance
func New() *Store {
	return &Store{
		results: make(map[string]*domain.ScanResult),
	}
}

// Set stores or updates a scan result for a pod
func (s *Store) Set(result *domain.ScanResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(result.PodNamespace, result.PodName)
	s.results[key] = result
}

// Get retrieves a scan result for a specific pod
func (s *Store) Get(namespace, podName string) (*domain.ScanResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(namespace, podName)
	result, exists := s.results[key]
	return result, exists
}

// GetAll retrieves all scan results
func (s *Store) GetAll() []*domain.ScanResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*domain.ScanResult, 0, len(s.results))
	for _, result := range s.results {
		results = append(results, result)
	}
	return results
}

// Delete removes a scan result for a specific pod
func (s *Store) Delete(namespace, podName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(namespace, podName)
	delete(s.results, key)
}

// Count returns the total number of stored results
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.results)
}

// makeKey creates a composite key from namespace and pod name
func (s *Store) makeKey(namespace, podName string) string {
	return namespace + "/" + podName
}
