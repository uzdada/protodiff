package bsr

// BufClient is the production-ready BSR client that uses buf CLI for schema fetching.
// It exports proto files from BSR using `buf export` command and parses them with protoparse.
//
// This approach is more reliable than the HTTP API because:
// - Gets complete proto file definitions (not just FileDescriptorSet)
// - Includes all message types and field definitions
// - Properly resolves type references
//
// Requirements:
// - buf CLI must be installed in the container
// - Writable /tmp directory for exports and cache
// - HOME environment variable set to writable directory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/uzdada/protodiff/internal/core/domain"
)

// BufClient uses buf CLI to fetch schemas from BSR
type BufClient struct {
	token string
}

// NewBufClient creates a new buf CLI based client
func NewBufClient() *BufClient {
	token := os.Getenv(envBSRToken)
	return &BufClient{
		token: token,
	}
}

// FetchSchema fetches schema from BSR using buf export
func (c *BufClient) FetchSchema(ctx context.Context, module string) (*domain.SchemaDescriptor, error) {
	// Create temp directory for export
	tmpDir, err := os.MkdirTemp("", "bsr-export-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Export proto files from BSR
	cmd := exec.CommandContext(ctx, "buf", "export", module, "-o", tmpDir)
	if c.token != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", envBSRToken, c.token))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("buf export failed: %w (output: %s)", err, string(output))
	}

	// Parse exported proto files
	protoFiles, err := filepath.Glob(filepath.Join(tmpDir, "**/*.proto"))
	if err != nil {
		return nil, fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		// Try without ** glob
		protoFiles, err = filepath.Glob(filepath.Join(tmpDir, "*.proto"))
		if err != nil {
			return nil, fmt.Errorf("failed to find proto files: %w", err)
		}
	}

	// Find all proto files recursively
	protoFiles = []string{}
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found in exported directory")
	}

	// Parse proto files
	parser := protoparse.Parser{
		ImportPaths: []string{tmpDir},
	}

	// Get relative paths
	var relPaths []string
	for _, file := range protoFiles {
		relPath, err := filepath.Rel(tmpDir, file)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}
		relPaths = append(relPaths, relPath)
	}

	fileDescs, err := parser.ParseFiles(relPaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto files: %w", err)
	}

	// Convert to SchemaDescriptor
	return fileDescriptorsToSchema(fileDescs), nil
}

// fileDescriptorsToSchema converts file descriptors to domain SchemaDescriptor
func fileDescriptorsToSchema(fileDescs []*desc.FileDescriptor) *domain.SchemaDescriptor {
	var services []domain.ServiceDescriptor
	var messages []string

	for _, fd := range fileDescs {
		// Extract services
		for _, svc := range fd.GetServices() {
			var methods []string
			for _, method := range svc.GetMethods() {
				methods = append(methods, method.GetName())
			}

			services = append(services, domain.ServiceDescriptor{
				Name:    svc.GetFullyQualifiedName(),
				Methods: methods,
			})
		}

		// Extract messages
		for _, msg := range fd.GetMessageTypes() {
			messages = append(messages, msg.GetFullyQualifiedName())
		}
	}

	return &domain.SchemaDescriptor{
		Services: services,
		Messages: messages,
	}
}

// Ensure BufClient implements Client interface
var _ Client = (*BufClient)(nil)
