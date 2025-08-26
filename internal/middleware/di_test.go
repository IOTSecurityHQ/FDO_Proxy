package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fdo-server-wrapper/internal/ledger"
)

// MockLedgerClient implements proxy.LedgerClient for testing
type MockLedgerClient struct {
	passport *ledger.ProductItemPassport
	err      error
}

func (m *MockLedgerClient) GetProductItemPassport(ctx context.Context, uuid string) (*ledger.ProductItemPassport, error) {
	return m.passport, m.err
}

func (m *MockLedgerClient) CreateCommissioningPassport(ctx context.Context, req *ledger.CommissioningCreateRequest) error {
	return m.err
}

func TestNewDIMiddleware(t *testing.T) {
	mockClient := &MockLedgerClient{}
	middleware := NewDIMiddleware(mockClient, true)

	if middleware == nil {
		t.Fatal("expected middleware but got nil")
	}

	if middleware.ledgerClient != mockClient {
		t.Error("expected ledger client to be set")
	}

	if !middleware.enableProductPassport {
		t.Error("expected product passport to be enabled")
	}
}

func TestDIMiddleware_IsDIRequest(t *testing.T) {
	middleware := &DIMiddleware{}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "DI.AppStart request",
			path:     "/fdo/101/msg/10",
			expected: true,
		},
		{
			name:     "DI.SetCredentials request",
			path:     "/fdo/101/msg/12",
			expected: true,
		},
		{
			name:     "non-DI request",
			path:     "/fdo/101/msg/20",
			expected: false,
		},
		{
			name:     "different path",
			path:     "/api/health",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			result := middleware.isDIRequest(req)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestDIMiddleware_IsDIResponse(t *testing.T) {
	middleware := &DIMiddleware{}

	tests := []struct {
		name     string
		msgType  string
		expected bool
	}{
		{
			name:     "DI.SetCredentials response",
			msgType:  "11",
			expected: true,
		},
		{
			name:     "DI.Done response",
			msgType:  "13",
			expected: true,
		},
		{
			name:     "non-DI response",
			msgType:  "20",
			expected: false,
		},
		{
			name:     "empty message type",
			msgType:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: make(http.Header),
			}
			resp.Header.Set("Message-Type", tt.msgType)

			result := middleware.isDIResponse(resp)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for message type %s", tt.expected, result, tt.msgType)
			}
		})
	}
}

func TestDIMiddleware_ProcessRequest_DIAppStart(t *testing.T) {
	mockPassport := &ledger.ProductItemPassport{
		UUID: "test-uuid",
		Records: []ledger.ProductItemRecord{
			{
				UUID:       "record-uuid",
				Signature:  "test-signature",
				Descriptor: "PRODUCT PASSPORT",
			},
		},
	}

	mockClient := &MockLedgerClient{
		passport: mockPassport,
		err:      nil,
	}

	middleware := &DIMiddleware{
		ledgerClient:          mockClient,
		enableProductPassport: true,
	}

	// Create a request that looks like DI.AppStart
	body := strings.NewReader("some cbor data with productId")
	req := httptest.NewRequest("POST", "/fdo/101/msg/10", body)

	ctx := context.Background()
	err := middleware.ProcessRequest(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDIMiddleware_ProcessRequest_NonDI(t *testing.T) {
	middleware := &DIMiddleware{
		enableProductPassport: true,
	}

	// Create a non-DI request
	req := httptest.NewRequest("POST", "/api/health", nil)

	ctx := context.Background()
	err := middleware.ProcessRequest(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDIMiddleware_ProcessResponse_DISetCredentials(t *testing.T) {
	middleware := &DIMiddleware{}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "11") // DI.SetCredentials

	ctx := context.Background()
	err := middleware.ProcessResponse(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDIMiddleware_ProcessResponse_NonDI(t *testing.T) {
	middleware := &DIMiddleware{}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "20") // Non-DI message

	ctx := context.Background()
	err := middleware.ProcessResponse(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDIMiddleware_ExtractProductID(t *testing.T) {
	middleware := &DIMiddleware{}

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "contains productId",
			body:     "some data with productId field",
			expected: "example-product-id",
		},
		{
			name:     "no productId",
			body:     "some other data",
			expected: "",
		},
		{
			name:     "empty body",
			body:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.extractProductID([]byte(tt.body))

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s' for body '%s'", tt.expected, result, tt.body)
			}
		})
	}
}

func TestDIMiddleware_HandleDIAppStart_Disabled(t *testing.T) {
	middleware := &DIMiddleware{
		enableProductPassport: false,
	}

	req := httptest.NewRequest("POST", "/fdo/101/msg/10", strings.NewReader("test"))

	ctx := context.Background()
	err := middleware.handleDIAppStart(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDIMiddleware_HandleDIAppStart_NoClient(t *testing.T) {
	middleware := &DIMiddleware{
		enableProductPassport: true,
		ledgerClient:          nil,
	}

	req := httptest.NewRequest("POST", "/fdo/101/msg/10", strings.NewReader("test"))

	ctx := context.Background()
	err := middleware.handleDIAppStart(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
