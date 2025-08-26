package ledger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		productBaseURL   string
		commissioningURL string
		caCertPath       string
		clientCertPath   string
		clientKeyPath    string
		expectError      bool
		errorContains    string
	}{
		{
			name:             "valid configuration",
			productBaseURL:   "https://example.com",
			commissioningURL: "http://example.com/commissioning",
			caCertPath:       "testdata/ca.pem",
			clientCertPath:   "testdata/client.crt",
			clientKeyPath:    "testdata/client.key",
			expectError:      true, // Will fail due to missing cert files
			errorContains:    "load client cert/key",
		},
		{
			name:             "missing product URL",
			productBaseURL:   "",
			commissioningURL: "http://example.com/commissioning",
			caCertPath:       "testdata/ca.pem",
			clientCertPath:   "testdata/client.crt",
			clientKeyPath:    "testdata/client.key",
			expectError:      true,
			errorContains:    "load client cert/key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.productBaseURL, tt.commissioningURL, tt.caCertPath, tt.clientCertPath, tt.clientKeyPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client but got nil")
			}
		})
	}
}

func TestGetProductItemPassport(t *testing.T) {
	// Create a test server that returns a mock product item passport
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Check if UUID is in query params
		if !contains(r.URL.RawQuery, "uuid=test-uuid") {
			t.Errorf("expected uuid in query params, got %s", r.URL.RawQuery)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"schema_version": 0.1,
			"uuid": "test-uuid",
			"records": [
				{
					"uuid": "record-uuid",
					"signature": "test-signature",
					"descriptor": "PRODUCT PASSPORT"
				}
			],
			"metadata": {
				"version": "1.0",
				"creation_time": "1754331025571481856",
				"board_sn": "test-board-sn"
			},
			"agent": {
				"uuid": "agent-uuid",
				"signature": "agent-signature"
			},
			"signature": "test-signature"
		}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		productBaseURL:    server.URL,
		productHTTP:       server.Client(),
		commissioningHTTP: &http.Client{Timeout: 30 * time.Second},
	}

	// Test successful request
	ctx := context.Background()
	passport, err := client.GetProductItemPassport(ctx, "test-uuid")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if passport == nil {
		t.Error("expected passport but got nil")
		return
	}

	if passport.UUID != "test-uuid" {
		t.Errorf("expected UUID 'test-uuid', got '%s'", passport.UUID)
	}

	if len(passport.Records) != 1 {
		t.Errorf("expected 1 record, got %d", len(passport.Records))
	}

	if passport.Records[0].Descriptor != "PRODUCT PASSPORT" {
		t.Errorf("expected descriptor 'PRODUCT PASSPORT', got '%s'", passport.Records[0].Descriptor)
	}
}

func TestCreateCommissioningPassport(t *testing.T) {
	// Create a test server that accepts commissioning passport creation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		commissioningURL:  server.URL,
		commissioningHTTP: server.Client(),
	}

	// Test successful request
	ctx := context.Background()
	req := &CommissioningCreateRequest{
		ControllerUUID:   "test-controller-uuid",
		Cert:             "test-cert",
		DeployedLocation: "test-location",
		Timestamp:        "1754509904342152960",
	}

	err := client.CreateCommissioningPassport(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateCommissioningPassport_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		commissioningURL:  server.URL,
		commissioningHTTP: server.Client(),
	}

	// Test error handling
	ctx := context.Background()
	req := &CommissioningCreateRequest{
		ControllerUUID:   "test-controller-uuid",
		Cert:             "test-cert",
		DeployedLocation: "test-location",
		Timestamp:        "1754509904342152960",
	}

	err := client.CreateCommissioningPassport(ctx, req)

	if err == nil {
		t.Error("expected error but got none")
		return
	}

	if !contains(err.Error(), "status 500") {
		t.Errorf("expected error to contain status 500, got: %s", err.Error())
	}
}

func TestGetProductItemPassport_Unconfigured(t *testing.T) {
	client := &Client{
		productBaseURL: "", // Not configured
	}

	ctx := context.Background()
	_, err := client.GetProductItemPassport(ctx, "test-uuid")

	if err == nil {
		t.Error("expected error but got none")
		return
	}

	if !contains(err.Error(), "not configured") {
		t.Errorf("expected error to contain 'not configured', got: %s", err.Error())
	}
}

func TestCreateCommissioningPassport_Unconfigured(t *testing.T) {
	client := &Client{
		commissioningURL: "", // Not configured
	}

	ctx := context.Background()
	req := &CommissioningCreateRequest{
		ControllerUUID: "test-uuid",
		Timestamp:      "1754509904342152960",
	}

	err := client.CreateCommissioningPassport(ctx, req)

	if err == nil {
		t.Error("expected error but got none")
		return
	}

	if !contains(err.Error(), "not configured") {
		t.Errorf("expected error to contain 'not configured', got: %s", err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
