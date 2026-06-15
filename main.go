package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

type HeartbeatResponse struct {
	OK     bool   `json:"ok"`
	Agent  string `json:"agent"`
	Status string `json:"status"`
}

func sendHeartbeat(agentID, baseURL, origin string) error {
	url := fmt.Sprintf("%s/api/agents/%s/heartbeat", baseURL, agentID)
	
	payload := map[string]string{"status": "online"}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", origin)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	
	var result HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	fmt.Printf("✓ Heartbeat sent: %s (status: %s)\n", result.Agent, result.Status)
	return nil
}

func runDaemon(agentID, baseURL, origin string, interval int) {
	fmt.Printf("Starting heartbeat daemon...\n")
	fmt.Printf("Agent: %s\n", agentID)
	fmt.Printf("URL: %s\n", baseURL)
	fmt.Printf("Interval: %ds\n\n", interval)
	
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	
	// Send first heartbeat immediately
	if err := sendHeartbeat(agentID, baseURL, origin); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Heartbeat failed: %v\n", err)
	}
	
	for range ticker.C {
		if err := sendHeartbeat(agentID, baseURL, origin); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Heartbeat failed: %v\n", err)
		}
	}
}

func main() {
	var (
		agentID  = flag.String("agent-id", "da8ad956-87ee-45e4-8f7a-e9e992c947d9", "Agent ID for heartbeat")
		baseURL  = flag.String("url", "http://localhost:8100", "Base URL for Mission Control")
		interval = flag.Int("interval", 60, "Heartbeat interval in seconds")
		origin   = flag.String("origin", "http://localhost:3000", "Origin header for auth")
		daemon   = flag.Bool("daemon", false, "Run in daemon mode")
	)
	flag.Parse()

	if *daemon {
		runDaemon(*agentID, *baseURL, *origin, *interval)
	} else {
		// Single heartbeat
		if err := sendHeartbeat(*agentID, *baseURL, *origin); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Heartbeat failed: %v\n", err)
			os.Exit(1)
		}
	}
}
