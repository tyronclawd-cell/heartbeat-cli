package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendHeartbeat(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json")
		}
		if r.Header.Get("Origin") != "http://localhost:3000" {
			t.Errorf("Expected Origin header")
		}

		// Check path
		expectedPath := "/api/agents/test-agent/heartbeat"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse body
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode body: %v", err)
		}
		if payload["status"] != "online" {
			t.Errorf("Expected status 'online', got %s", payload["status"])
		}

		// Send response
		resp := HeartbeatResponse{OK: true, Agent: "test-agent", Status: "online"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test successful heartbeat
	err := sendHeartbeat("test-agent", server.URL, "http://localhost:3000")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendHeartbeatFailure(t *testing.T) {
	// Create a mock server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Test failed heartbeat
	err := sendHeartbeat("test-agent", server.URL, "http://localhost:3000")
	if err == nil {
		t.Error("Expected error for 500 response")
	}
}

func TestHeartbeatResponseStructure(t *testing.T) {
	resp := HeartbeatResponse{
		OK:     true,
		Agent:  "mac mini gateway",
		Status: "online",
	}

	if !resp.OK {
		t.Error("Expected OK to be true")
	}
	if resp.Agent != "mac mini gateway" {
		t.Errorf("Expected agent 'mac mini gateway', got %s", resp.Agent)
	}
	if resp.Status != "online" {
		t.Errorf("Expected status 'online', got %s", resp.Status)
	}
}

func TestSendHeartbeatTimeout(t *testing.T) {
	// Create a mock server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second)
	}))
	defer server.Close()

	// Test timeout - this should fail due to client timeout
	err := sendHeartbeat("test-agent", server.URL, "http://localhost:3000")
	if err == nil {
		t.Error("Expected timeout error")
	}
}
