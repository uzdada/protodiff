package bsr

// DEPRECATED: HTTPClient is not production-ready due to incomplete data returned by BSR API.
// The FileDescriptorSet API returns schemas with missing message types and empty input/output references.
// Use BufClient instead, which uses `buf export` CLI for reliable schema fetching.
//
// This implementation is kept for educational purposes and potential future use if BSR API is fixed.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/uzdada/protodiff/internal/core/domain"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	defaultBSRURL     = "https://buf.build"
	fileDescriptorAPI = "/buf.reflect.v1beta1.FileDescriptorSetService/GetFileDescriptorSet"
	httpClientTimeout = 30 * time.Second
	envBSRToken       = "BSR_TOKEN"
	envBSRURL         = "BSR_URL"
)

// HTTPClient is a real implementation of BSR Client using HTTP
type HTTPClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewHTTPClient creates a new BSR HTTP client
func NewHTTPClient() *HTTPClient {
	token := os.Getenv(envBSRToken)
	baseURL := os.Getenv(envBSRURL)
	if baseURL == "" {
		baseURL = defaultBSRURL
	}

	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: httpClientTimeout,
		},
		baseURL: baseURL,
		token:   token,
	}
}

// NewHTTPClientWithToken creates a BSR client with explicit token
func NewHTTPClientWithToken(token string) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: httpClientTimeout,
		},
		baseURL: defaultBSRURL,
		token:   token,
	}
}

// FetchSchema retrieves the schema from BSR using the FileDescriptorSet API
func (c *HTTPClient) FetchSchema(ctx context.Context, module string) (*domain.SchemaDescriptor, error) {
	// Build request payload
	reqPayload := map[string]string{
		"module": module,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build HTTP request
	url := c.baseURL + fileDescriptorAPI
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("BSR API returned status %d (unable to read response body: %w)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("BSR API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var responseData struct {
		FileDescriptorSet *descriptorpb.FileDescriptorSet `json:"fileDescriptorSet"`
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(respBody, &responseData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if responseData.FileDescriptorSet == nil {
		return nil, fmt.Errorf("no FileDescriptorSet in response")
	}

	// Debug: log file descriptor set info
	fmt.Printf("DEBUG: FileDescriptorSet contains %d files\n", len(responseData.FileDescriptorSet.File))
	for i, file := range responseData.FileDescriptorSet.File {
		fmt.Printf("DEBUG: File %d: %s (deps: %d, messages: %d, services: %d)\n",
			i, file.GetName(), len(file.GetDependency()), len(file.GetMessageType()), len(file.GetService()))
		// Debug message types
		for _, msg := range file.GetMessageType() {
			fmt.Printf("  Message: %s\n", msg.GetName())
		}
		// Debug services
		for _, svc := range file.GetService() {
			fmt.Printf("  Service: %s\n", svc.GetName())
			for _, method := range svc.GetMethod() {
				fmt.Printf("    Method: %s (input: %s, output: %s)\n",
					method.GetName(), method.GetInputType(), method.GetOutputType())
			}
		}
	}

	// Convert FileDescriptorSet to SchemaDescriptor
	schema, err := c.fileDescriptorSetToSchema(responseData.FileDescriptorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to convert FileDescriptorSet: %w", err)
	}

	return schema, nil
}

// fileDescriptorSetToSchema converts a protobuf FileDescriptorSet to our domain SchemaDescriptor
func (c *HTTPClient) fileDescriptorSetToSchema(fds *descriptorpb.FileDescriptorSet) (*domain.SchemaDescriptor, error) {
	if fds == nil || len(fds.File) == 0 {
		return nil, fmt.Errorf("empty FileDescriptorSet")
	}

	// Parse file descriptors with dependencies
	// Use CreateFileDescriptors to handle dependencies properly
	fileDescs, err := desc.CreateFileDescriptors(fds.File)
	if err != nil {
		return nil, fmt.Errorf("failed to create file descriptors: %w", err)
	}

	// Extract services and messages
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
	}, nil
}

// Ensure HTTPClient implements Client interface
var _ Client = (*HTTPClient)(nil)
