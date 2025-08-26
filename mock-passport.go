package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ProductItemPassport struct {
	SchemaVersion string `json:"schema_version"`
	UUID          string `json:"uuid"`
	Records       []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"records"`
	Metadata struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	} `json:"metadata"`
	Agent struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"agent"`
	Signature string `json:"signature"`
}

func main() {
	// Product Item Passport endpoint (mTLS simulation)
	http.HandleFunc("/product_item/", func(w http.ResponseWriter, r *http.Request) {
		uuid := r.URL.Query().Get("uuid")
		if uuid == "" {
			http.Error(w, "Missing uuid parameter", http.StatusBadRequest)
			return
		}

		passport := ProductItemPassport{
			SchemaVersion: "1.0",
			UUID:          uuid,
			Records: []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			}{
				{ID: "record1", Type: "manufacturing"},
				{ID: "record2", Type: "certification"},
			},
			Metadata: struct {
				CreatedAt string `json:"created_at"`
				UpdatedAt string `json:"updated_at"`
			}{
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
			Agent: struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}{
				Name:    "Mock Passport Service",
				Version: "1.0.0",
			},
			Signature: "mock-signature-12345",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(passport)
	})

	// Commissioning Passport creation endpoint
	http.HandleFunc("/create-commissioning-passport", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ControllerUUID   string `json:"controller_uuid"`
			Cert             string `json:"cert"`
			DeployedLocation string `json:"deployed_location"`
			Timestamp        string `json:"timestamp"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.ControllerUUID == "" {
			http.Error(w, "Missing controller_uuid", http.StatusBadRequest)
			return
		}

		// Simulate successful creation
		response := map[string]string{
			"status":  "success",
			"message": "Commissioning passport created",
			"id":      fmt.Sprintf("commissioning-%s-%d", req.ControllerUUID, time.Now().Unix()),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Starting mock passport service on :8443 (product) and :8000 (commissioning)")

	// Start product passport service (mTLS simulation)
	go func() {
		log.Fatal(http.ListenAndServe(":8443", nil))
	}()

	// Start commissioning passport service
	log.Fatal(http.ListenAndServe(":8000", nil))
}
