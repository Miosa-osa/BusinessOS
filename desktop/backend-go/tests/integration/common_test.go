package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	businessOSURL = getEnvOrDefault("BUSINESSOS_URL", "http://localhost:8001")
	oxigraphURL   = getEnvOrDefault("OXIGRAPH_URL", "http://localhost:6379")
	testTimeout   = 30 * time.Second
)

// TestMain skips the entire integration package when BusinessOS is not reachable.
func TestMain(m *testing.M) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(businessOSURL + "/health")
	if err != nil {
		fmt.Printf("SKIP integration tests: BusinessOS not reachable at %s (%v)\n", businessOSURL, err)
		os.Exit(0)
	}
	resp.Body.Close()
	os.Exit(m.Run())
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RDFTriple represents an RDF triple
type RDFTriple struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// requireRoute skips the test if the given endpoint returns 404 — route not yet deployed.
func requireRoute(t *testing.T, method, url string) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Skipf("could not build request for %s %s: %v", method, url, err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("route %s %s unreachable: %v", method, url, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		t.Skipf("route %s %s not deployed yet (404) — rebuild server to activate", method, url)
	}
}

// makeRequest makes an HTTP request and returns the response body, status code, and error.
func makeRequest(method, url string, body interface{}) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: testTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}
