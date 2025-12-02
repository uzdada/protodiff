// Package scanner provides the core orchestration logic for schema drift detection.
//
// The scanner continuously discovers gRPC-enabled pods in Kubernetes, fetches their
// schemas via gRPC reflection, retrieves the canonical schemas from BSR, and compares
// them to detect drift. Results are stored in-memory and surfaced through the web UI.
//
// The main workflow is:
//  1. Load service-to-BSR mappings from ConfigMap
//  2. Discover pods for services specified in ConfigMap (or fallback to label-based discovery)
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

	// Get service names from ConfigMap for targeted discovery
	serviceNames := mappings.GetServiceNames()

	var pods []k8s.PodInfo
	if len(serviceNames) > 0 {
		// Use ConfigMap-based discovery for better efficiency
		log.Printf("Using ConfigMap-based discovery for %d services", len(serviceNames))
		pods, err = s.k8sClient.DiscoverPodsForServices(ctx, serviceNames)
		if err != nil {
			return fmt.Errorf("failed to discover pods for services: %w", err)
		}
	} else {
		// Fallback to label-based discovery if ConfigMap is empty
		log.Println("ConfigMap is empty, falling back to label-based discovery")
		pods, err = s.k8sClient.DiscoverGRPCPods(ctx)
		if err != nil {
			return fmt.Errorf("failed to discover pods: %w", err)
		}
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

// validatePod validates a single pod's schema against BSR.
// It orchestrates the validation workflow: creating result, resolving BSR module,
// fetching schemas, comparing them, and storing the result.
func (s *Scanner) validatePod(ctx context.Context, pod k8s.PodInfo, mappings domain.ServiceMappings) {
	result := s.createScanResult(pod)

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

	// Fetch and compare schemas
	s.fetchAndCompareSchemas(ctx, pod, bsrModule, result)

	s.store.Set(result)
	log.Printf("Validated %s/%s: %s", pod.Namespace, pod.Name, result.Status)
}

// createScanResult initializes a new ScanResult from pod information.
// All results start with StatusUnknown until validation completes.
func (s *Scanner) createScanResult(pod k8s.PodInfo) *domain.ScanResult {
	return &domain.ScanResult{
		PodName:      pod.Name,
		PodNamespace: pod.Namespace,
		ServiceName:  pod.ServiceName,
		PodIP:        pod.IP,
		GRPCPort:     pod.GRPCPort,
		LastChecked:  time.Now(),
		Status:       domain.StatusUnknown,
	}
}

// fetchAndCompareSchemas retrieves schemas from both the live pod and BSR,
// then compares them to detect drift. Updates the result with comparison outcome.
func (s *Scanner) fetchAndCompareSchemas(ctx context.Context, pod k8s.PodInfo, bsrModule string, result *domain.ScanResult) {
	// Fetch live schema via gRPC reflection
	address := fmt.Sprintf("%s:%d", pod.IP, pod.GRPCPort)
	log.Printf("Connecting to %s/%s at %s (port %d)", pod.Namespace, pod.Name, address, pod.GRPCPort)
	liveSchema, err := s.grpcClient.FetchSchema(ctx, address)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to fetch live schema: %v", err)
		result.Status = domain.StatusUnknown
		return
	}

	// Fetch truth schema from BSR
	log.Printf("Fetching BSR schema for module: %s", bsrModule)
	truthSchema, err := s.bsrClient.FetchSchema(ctx, bsrModule)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to fetch BSR schema: %v", err)
		result.Status = domain.StatusUnknown
		log.Printf("BSR fetch error for %s: %v", bsrModule, err)
		return
	}
	log.Printf("BSR schema fetched: %d services, %d messages", len(truthSchema.Services), len(truthSchema.Messages))

	// Compare schemas and get detailed diff
	match, diff := s.compareSchemas(liveSchema, truthSchema)
	result.SchemaDiff = diff

	if match {
		result.Status = domain.StatusSync
		result.Message = "Schemas are in sync"
		log.Printf("✓ Schemas match for %s/%s", pod.Namespace, pod.Name)
	} else {
		result.Status = domain.StatusMismatch
		result.Message = s.buildDiffMessage(diff)
		log.Printf("✗ Schema mismatch for %s/%s: %s", pod.Namespace, pod.Name, result.Message)
	}
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

// compareSchemas compares two schema descriptors and returns match status with detailed diff.
// Only compares services that exist in BOTH live and BSR (intersection).
// Services that exist only in live or only in BSR are tracked but don't affect sync status.
func (s *Scanner) compareSchemas(live, truth *domain.SchemaDescriptor) (bool, *domain.SchemaDiff) {
	diff := &domain.SchemaDiff{
		LiveServices: []string{},
		BSRServices:  []string{},
		MissingInLive: []string{},
		ExtraInLive: []string{},
		MethodMismatches: []domain.ServiceMethodMismatch{},
	}

	// Validate inputs are not nil
	if live == nil || truth == nil {
		return false, diff
	}

	// Create maps for comparison
	liveServicesMap := make(map[string][]string)
	truthServicesMap := make(map[string][]string)

	// Populate live services
	for _, svc := range live.Services {
		diff.LiveServices = append(diff.LiveServices, svc.Name)
		liveServicesMap[svc.Name] = svc.Methods
	}

	// Populate truth services
	for _, svc := range truth.Services {
		diff.BSRServices = append(diff.BSRServices, svc.Name)
		truthServicesMap[svc.Name] = svc.Methods
	}

	match := true

	// Track services in live but not in BSR (informational only, doesn't affect match)
	for liveSvcName := range liveServicesMap {
		if _, exists := truthServicesMap[liveSvcName]; !exists {
			diff.ExtraInLive = append(diff.ExtraInLive, liveSvcName)
			// NOTE: Extra services in live are OK - they're not compared
		}
	}

	// Track services in BSR but not in live (informational only, doesn't affect match)
	for truthSvcName := range truthServicesMap {
		if _, exists := liveServicesMap[truthSvcName]; !exists {
			diff.MissingInLive = append(diff.MissingInLive, truthSvcName)
			// NOTE: Missing services are OK - we only compare intersection
		}
	}

	// ONLY compare services that exist in BOTH live and BSR
	for liveSvcName, liveMethods := range liveServicesMap {
		if truthMethods, exists := truthServicesMap[liveSvcName]; exists {
			// This service exists in both - compare methods
			if !s.methodsMatch(liveMethods, truthMethods) {
				missing, extra := s.diffMethods(liveMethods, truthMethods)
				diff.MethodMismatches = append(diff.MethodMismatches, domain.ServiceMethodMismatch{
					ServiceName:    liveSvcName,
					LiveMethods:    len(liveMethods),
					BSRMethods:     len(truthMethods),
					MissingMethods: missing,
					ExtraMethods:   extra,
				})
				match = false // Only method mismatches cause MISMATCH status
			}
		}
	}

	return match, diff
}

// methodsMatch checks if two method slices are equal
func (s *Scanner) methodsMatch(live, truth []string) bool {
	if len(live) != len(truth) {
		return false
	}

	// Create a map for quick lookup
	truthMap := make(map[string]bool)
	for _, method := range truth {
		truthMap[method] = true
	}

	// Check all live methods exist in truth
	for _, method := range live {
		if !truthMap[method] {
			return false
		}
	}

	return true
}

// diffMethods returns missing and extra methods
func (s *Scanner) diffMethods(live, truth []string) (missing []string, extra []string) {
	liveMap := make(map[string]bool)
	truthMap := make(map[string]bool)

	for _, m := range live {
		liveMap[m] = true
	}
	for _, m := range truth {
		truthMap[m] = true
	}

	// Find missing methods (in truth but not in live)
	for _, m := range truth {
		if !liveMap[m] {
			missing = append(missing, m)
		}
	}

	// Find extra methods (in live but not in truth)
	for _, m := range live {
		if !truthMap[m] {
			extra = append(extra, m)
		}
	}

	return missing, extra
}

// buildDiffMessage creates a human-readable diff message
func (s *Scanner) buildDiffMessage(diff *domain.SchemaDiff) string {
	if diff == nil {
		return "Schema drift detected"
	}

	var msg strings.Builder

	// Count services in both (intersection)
	commonServices := 0
	for _, liveSvc := range diff.LiveServices {
		for _, bsrSvc := range diff.BSRServices {
			if liveSvc == bsrSvc {
				commonServices++
				break
			}
		}
	}

	if len(diff.MethodMismatches) > 0 {
		msg.WriteString("Method mismatches: ")
		for i, mismatch := range diff.MethodMismatches {
			if i > 0 {
				msg.WriteString("; ")
			}
			msg.WriteString(fmt.Sprintf("%s (live:%d, BSR:%d)",
				mismatch.ServiceName, mismatch.LiveMethods, mismatch.BSRMethods))

			if len(mismatch.MissingMethods) > 0 {
				msg.WriteString(fmt.Sprintf(" missing:%s", strings.Join(mismatch.MissingMethods, ",")))
			}
			if len(mismatch.ExtraMethods) > 0 {
				msg.WriteString(fmt.Sprintf(" extra:%s", strings.Join(mismatch.ExtraMethods, ",")))
			}
		}
		return msg.String()
	}

	// No method mismatches - schemas are in sync
	if commonServices > 0 {
		return fmt.Sprintf("All %d common service(s) in sync", commonServices)
	}

	// No common services to compare
	return "No common services to compare"
}
