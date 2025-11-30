// Package scanner provides the core orchestration logic for schema drift detection.
//
// The scanner continuously discovers gRPC-enabled pods in Kubernetes, fetches their
// schemas via gRPC reflection, retrieves the canonical schemas from BSR, and compares
// them to detect drift. Results are stored in-memory and surfaced through the web UI.
//
// The main workflow is:
//  1. Discover pods with grpc-service=true label
//  2. Load service-to-BSR mappings from ConfigMap
//  3. For each pod:
//     - Fetch live schema via gRPC reflection
//     - Fetch truth schema from BSR
//     - Compare schemas and detect drift
//  4. Store results for dashboard display
//
// The scanner runs continuously on a configurable interval (default: 30 minutes).
package scanner

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/uzdada/protodiff/internal/adapters/bsr"
	"github.com/uzdada/protodiff/internal/adapters/grpc"
	"github.com/uzdada/protodiff/internal/adapters/k8s"
	"github.com/uzdada/protodiff/internal/config"
	"github.com/uzdada/protodiff/internal/core/domain"
	"github.com/uzdada/protodiff/internal/core/store"
)

// Scanner orchestrates the schema validation workflow
type Scanner struct {
	k8sClient     *k8s.Client
	grpcClient    *grpc.ReflectionClient
	bsrClient     bsr.Client
	store         *store.Store
	configMapNS   string
	configMapName string
	bsrTemplate   string
	scanInterval  time.Duration
}

// NewScanner creates a new scanner instance
func NewScanner(
	k8sClient *k8s.Client,
	grpcClient *grpc.ReflectionClient,
	bsrClient bsr.Client,
	store *store.Store,
	cfg config.Config,
) *Scanner {
	return &Scanner{
		k8sClient:     k8sClient,
		grpcClient:    grpcClient,
		bsrClient:     bsrClient,
		store:         store,
		configMapNS:   cfg.ConfigMapNamespace,
		configMapName: cfg.ConfigMapName,
		bsrTemplate:   cfg.BSRTemplate,
		scanInterval:  cfg.ScanInterval,
	}
}

// Start begins the continuous scanning loop
func (s *Scanner) Start(ctx context.Context) error {
	log.Printf("Starting scanner with interval: %s", s.scanInterval)

	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	// Run initial scan immediately
	if err := s.runScan(ctx); err != nil {
		log.Printf("Initial scan failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Scanner stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.runScan(ctx); err != nil {
				log.Printf("Scan failed: %v", err)
			}
		}
	}
}

// runScan performs a single scan cycle
func (s *Scanner) runScan(ctx context.Context) error {
	log.Println("Starting scan cycle...")

	// Load service mappings from ConfigMap
	mappings, err := s.loadServiceMappings(ctx)
	if err != nil {
		log.Printf("Warning: Failed to load ConfigMap: %v", err)
		mappings = domain.NewServiceMappings(nil) // Empty mappings
	}

	// Discover gRPC pods
	pods, err := s.k8sClient.DiscoverGRPCPods(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover pods: %w", err)
	}

	log.Printf("Discovered %d gRPC pods", len(pods))

	// Validate each pod
	for _, pod := range pods {
		s.validatePod(ctx, pod, mappings)
	}

	log.Printf("Scan cycle completed. Results stored: %d", s.store.Count())
	return nil
}

// loadServiceMappings loads the ConfigMap or returns empty mappings on error
func (s *Scanner) loadServiceMappings(ctx context.Context) (domain.ServiceMappings, error) {
	return s.k8sClient.LoadServiceMappings(ctx, s.configMapNS, s.configMapName)
}

// validatePod validates a single pod's schema against BSR
func (s *Scanner) validatePod(ctx context.Context, pod k8s.PodInfo, mappings domain.ServiceMappings) {
	result := &domain.ScanResult{
		PodName:      pod.Name,
		PodNamespace: pod.Namespace,
		ServiceName:  pod.ServiceName,
		PodIP:        pod.IP,
		GRPCPort:     pod.GRPCPort,
		LastChecked:  time.Now(),
		Status:       domain.StatusUnknown,
	}

	// Resolve BSR module
	bsrModule := s.resolveBSRModule(pod.ServiceName, mappings)
	result.BSRModule = bsrModule

	if bsrModule == "" {
		result.Message = "No BSR module mapping found"
		s.store.Set(result)
		return
	}

	// Validate pod IP is not empty
	if pod.IP == "" {
		result.Message = "Pod IP is empty, cannot connect to gRPC service"
		result.Status = domain.StatusUnknown
		s.store.Set(result)
		return
	}

	// Fetch live schema via gRPC reflection
	address := fmt.Sprintf("%s:%d", pod.IP, pod.GRPCPort)
	liveSchema, err := s.grpcClient.FetchSchema(ctx, address)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to fetch live schema: %v", err)
		result.Status = domain.StatusUnknown
		s.store.Set(result)
		return
	}

	// Fetch truth schema from BSR
	truthSchema, err := s.bsrClient.FetchSchema(ctx, bsrModule)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to fetch BSR schema: %v", err)
		result.Status = domain.StatusUnknown
		s.store.Set(result)
		return
	}

	// Compare schemas
	if s.schemasMatch(liveSchema, truthSchema) {
		result.Status = domain.StatusSync
		result.Message = "Schemas are in sync"
	} else {
		result.Status = domain.StatusMismatch
		result.Message = "Schema drift detected"
	}

	s.store.Set(result)
	log.Printf("Validated %s/%s: %s", pod.Namespace, pod.Name, result.Status)
}

// resolveBSRModule determines the BSR module for a service
func (s *Scanner) resolveBSRModule(serviceName string, mappings domain.ServiceMappings) string {
	// Check ConfigMap first
	if module, exists := mappings.Get(serviceName); exists {
		return module
	}

	// Fallback to template
	if s.bsrTemplate != "" {
		return strings.ReplaceAll(s.bsrTemplate, "{service}", serviceName)
	}

	return ""
}

// schemasMatch compares two schema descriptors
func (s *Scanner) schemasMatch(live, truth *domain.SchemaDescriptor) bool {
	// Validate inputs are not nil
	if live == nil || truth == nil {
		return false
	}

	// Simple comparison: check if service names and methods match
	if len(live.Services) != len(truth.Services) {
		return false
	}

	// Create a map of truth services for quick lookup
	truthServices := make(map[string][]string)
	for _, svc := range truth.Services {
		truthServices[svc.Name] = svc.Methods
	}

	// Check if all live services exist in truth and have matching methods
	for _, liveSvc := range live.Services {
		truthMethods, exists := truthServices[liveSvc.Name]
		if !exists {
			return false
		}

		// Compare methods (simple length check for MVP)
		if len(liveSvc.Methods) != len(truthMethods) {
			return false
		}
	}

	return true
}
