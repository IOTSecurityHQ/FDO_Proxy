package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTO2Middleware(t *testing.T) {
	mockClient := &MockLedgerClient{}
	ownerID := "test-owner"

	middleware := NewTO2Middleware(mockClient, ownerID)

	if middleware == nil {
		t.Fatal("expected middleware but got nil")
	}

	if middleware.ledgerClient != mockClient {
		t.Error("expected ledger client to be set")
	}

	if middleware.ownerID != ownerID {
		t.Errorf("expected owner ID '%s', got '%s'", ownerID, middleware.ownerID)
	}
}

func TestTO2Middleware_IsTO2Request(t *testing.T) {
	middleware := &TO2Middleware{}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "TO2.HelloDevice request",
			path:     "/fdo/101/msg/60",
			expected: true,
		},
		{
			name:     "TO2.ProveDevice request",
			path:     "/fdo/101/msg/70",
			expected: true,
		},
		{
			name:     "non-TO2 request",
			path:     "/fdo/101/msg/80",
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
			result := middleware.isTO2Request(req)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestTO2Middleware_IsTO2Response(t *testing.T) {
	middleware := &TO2Middleware{}

	tests := []struct {
		name     string
		msgType  string
		expected bool
	}{
		{
			name:     "TO2.Done2 response",
			msgType:  "71",
			expected: true,
		},
		{
			name:     "non-TO2 response",
			msgType:  "80",
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

			result := middleware.isTO2Response(resp)

			if result != tt.expected {
				t.Errorf("expected %v, got %v for message type %s", tt.expected, result, tt.msgType)
			}
		})
	}
}

func TestTO2Middleware_ProcessRequest_TO2HelloDevice(t *testing.T) {
	middleware := &TO2Middleware{}

	// Create a request that looks like TO2.HelloDevice
	req := httptest.NewRequest("POST", "/fdo/101/msg/60", nil)

	ctx := context.Background()
	err := middleware.ProcessRequest(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_ProcessRequest_NonTO2(t *testing.T) {
	middleware := &TO2Middleware{}

	// Create a non-TO2 request
	req := httptest.NewRequest("POST", "/api/health", nil)

	ctx := context.Background()
	err := middleware.ProcessRequest(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_ProcessResponse_TO2Done2(t *testing.T) {
	mockClient := &MockLedgerClient{
		err: nil, // No error from commissioning passport creation
	}

	middleware := &TO2Middleware{
		ledgerClient: mockClient,
		ownerID:      "test-owner",
	}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "71") // TO2.Done2

	ctx := context.Background()
	err := middleware.ProcessResponse(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_ProcessResponse_NonTO2(t *testing.T) {
	middleware := &TO2Middleware{}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "80") // Non-TO2 message

	ctx := context.Background()
	err := middleware.ProcessResponse(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_HandleTO2Done2_NoClient(t *testing.T) {
	middleware := &TO2Middleware{
		ledgerClient: nil,
		ownerID:      "test-owner",
	}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "71")

	ctx := context.Background()
	err := middleware.handleTO2Done2(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_HandleTO2Done2_WithClient(t *testing.T) {
	mockClient := &MockLedgerClient{
		err: nil,
	}

	middleware := &TO2Middleware{
		ledgerClient: mockClient,
		ownerID:      "test-owner",
	}

	resp := &http.Response{
		Header: make(http.Header),
	}
	resp.Header.Set("Message-Type", "71")

	ctx := context.Background()
	err := middleware.handleTO2Done2(ctx, resp)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTO2Middleware_ExtractDeviceGUID(t *testing.T) {
	middleware := &TO2Middleware{}

	resp := &http.Response{
		Header: make(http.Header),
	}

	result := middleware.extractDeviceGUID(resp)

	// Currently returns placeholder value
	expected := "example-device-guid"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestTO2Middleware_HandleTO2HelloDevice(t *testing.T) {
	middleware := &TO2Middleware{}

	req := httptest.NewRequest("POST", "/fdo/101/msg/60", nil)

	ctx := context.Background()
	err := middleware.handleTO2HelloDevice(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
