package bsr

import (
	"context"
	"fmt"

	"github.com/uzdada/protodiff/internal/core/domain"
)

// MockClient is a mock implementation of the BSR Client for testing
type MockClient struct {
	// schemas stores predefined schemas for testing
	schemas map[string]*domain.SchemaDescriptor
}

// NewMockClient creates a new mock BSR client with sample data
func NewMockClient() *MockClient {
	return &MockClient{
		schemas: map[string]*domain.SchemaDescriptor{
			"buf.build/example/greeter": {
				Services: []domain.ServiceDescriptor{
					{
						Name: "greeter.Greeter",
						Methods: []string{
							"SayHello",
							"SayHelloAgain",
						},
					},
				},
				Messages: []string{
					"greeter.HelloRequest",
					"greeter.HelloReply",
				},
			},
			"buf.build/example/user": {
				Services: []domain.ServiceDescriptor{
					{
						Name: "user.UserService",
						Methods: []string{
							"GetUser",
							"CreateUser",
							"ListUsers",
						},
					},
				},
				Messages: []string{
					"user.GetUserRequest",
					"user.UserResponse",
					"user.CreateUserRequest",
					"user.ListUsersRequest",
					"user.ListUsersResponse",
				},
			},
			"buf.build/acme/user": {
				Services: []domain.ServiceDescriptor{
					{
						Name: "user.v1.UserService",
						Methods: []string{
							"GetUser",
							"CreateUser",
							"UpdateUser",
							"DeleteUser",
							"ListUsers",
						},
					},
				},
				Messages: []string{
					"user.v1.User",
					"user.v1.GetUserRequest",
					"user.v1.GetUserResponse",
					"user.v1.CreateUserRequest",
					"user.v1.CreateUserResponse",
				},
			},
			"buf.build/acme/order": {
				Services: []domain.ServiceDescriptor{
					{
						Name: "order.v1.OrderService",
						Methods: []string{
							"CreateOrder",
							"GetOrder",
							"ListOrders",
							"CancelOrder",
						},
					},
				},
				Messages: []string{
					"order.v1.Order",
					"order.v1.CreateOrderRequest",
					"order.v1.GetOrderRequest",
					"order.v1.ListOrdersRequest",
				},
			},
			"buf.build/acme/payment": {
				Services: []domain.ServiceDescriptor{
					{
						Name: "payment.v1.PaymentService",
						Methods: []string{
							"ProcessPayment",
							"RefundPayment",
							"GetPaymentStatus",
						},
					},
				},
				Messages: []string{
					"payment.v1.Payment",
					"payment.v1.ProcessPaymentRequest",
					"payment.v1.RefundPaymentRequest",
				},
			},
		},
	}
}

// FetchSchema retrieves a mock schema from the predefined set
func (m *MockClient) FetchSchema(ctx context.Context, module string) (*domain.SchemaDescriptor, error) {
	schema, exists := m.schemas[module]
	if !exists {
		return nil, fmt.Errorf("schema not found for module: %s", module)
	}

	// Return a copy to avoid mutations
	schemaCopy := &domain.SchemaDescriptor{
		Services: make([]domain.ServiceDescriptor, len(schema.Services)),
		Messages: make([]string, len(schema.Messages)),
	}
	copy(schemaCopy.Services, schema.Services)
	copy(schemaCopy.Messages, schema.Messages)

	return schemaCopy, nil
}

// AddMockSchema adds a custom schema to the mock client for testing
func (m *MockClient) AddMockSchema(module string, schema *domain.SchemaDescriptor) {
	m.schemas[module] = schema
}
