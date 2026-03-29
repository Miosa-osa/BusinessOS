package chaos

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMain skips all chaos tests gracefully when the full Docker stack is not
// available. Chaos tests require the businessos-backend container running inside
// Docker so network-isolation faults actually affect it.
// This matches the skip pattern used by tests/integration/common_test.go.
func TestMain(m *testing.M) {
	// Require businessos-backend container running (not just a local binary).
	// Network partition only affects containers on businessos_businessos-network.
	cmd := exec.Command("docker", "compose", "ps", "--filter", "status=running", "-q", "backend")
	cmd.Dir = "/Users/sac/chatmangpt/BusinessOS"
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		fmt.Println("SKIP chaos tests: businessos-backend container not running — run `make dev` first")
		os.Exit(0)
	}

	// Sanity-check OSA health too (TestNetworkPartition targets it).
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:8089/health")
	if err != nil || resp.StatusCode >= 500 {
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Println("SKIP chaos tests: osa health endpoint not responding — run `make dev` first")
		os.Exit(0)
	}
	resp.Body.Close()
	os.Exit(m.Run())
}

// ChaosTest defines a single chaos engineering test scenario
type ChaosTest struct {
	name         string
	breakdown    func() error
	recover      func() error
	maxDetection time.Duration
	maxRecovery  time.Duration
}

// Run executes a chaos test with timing and verification
func (c *ChaosTest) Run(t *testing.T) {
	t.Run(c.name, func(t *testing.T) {
		t.Logf("Starting chaos test: %s", c.name)

		// Phase 1: Breakdown
		t.Logf("Phase 1: Injecting failure...")
		breakdownStart := time.Now()
		if err := c.breakdown(); err != nil {
			t.Fatalf("Failed to inject breakdown: %v", err)
		}
		breakdownDuration := time.Since(breakdownStart)
		t.Logf("Breakdown completed in %v", breakdownDuration)

		// Phase 2: Detection
		t.Logf("Phase 2: Waiting for detection (max: %v)...", c.maxDetection)
		detectionStart := time.Now()
		detected := c.waitForDetection(t, c.maxDetection)
		detectionDuration := time.Since(detectionStart)

		if !detected {
			t.Errorf("FAILURE: System did not detect failure within %v", c.maxDetection)
		} else {
			t.Logf("SUCCESS: Failure detected in %v (threshold: %v)", detectionDuration, c.maxDetection)
			if detectionDuration > c.maxDetection {
				t.Errorf("FAILURE: Detection time %v exceeds threshold %v", detectionDuration, c.maxDetection)
			}
		}

		// Phase 3: Recovery
		t.Logf("Phase 3: Initiating recovery...")
		recoveryStart := time.Now()
		if err := c.recover(); err != nil {
			t.Fatalf("Failed to recover: %v", err)
		}

		// Phase 4: Verification
		t.Logf("Phase 4: Verifying recovery (max: %v)...", c.maxRecovery)
		recovered := c.waitForRecovery(t, c.maxRecovery)
		recoveryDuration := time.Since(recoveryStart)

		if !recovered {
			t.Errorf("FAILURE: System did not recover within %v", c.maxRecovery)
		} else {
			t.Logf("SUCCESS: System recovered in %v (threshold: %v)", recoveryDuration, c.maxRecovery)
			if recoveryDuration > c.maxRecovery {
				t.Errorf("FAILURE: Recovery time %v exceeds threshold %v", recoveryDuration, c.maxRecovery)
			}
		}

		// Report metrics
		t.Logf("=== Chaos Test Metrics ===")
		t.Logf("Detection Time: %v / %v", detectionDuration, c.maxDetection)
		t.Logf("Recovery Time: %v / %v", recoveryDuration, c.maxRecovery)
		t.Logf("Total Incident Duration: %v", detectionDuration+recoveryDuration)
	})
}

// waitForDetection polls until system detects the failure
func (c *ChaosTest) waitForDetection(t *testing.T, timeout time.Duration) bool {
	ctx := make(chan interface{})
	timeoutChan := time.After(timeout)

	go func() {
		// Default detection logic: check if system reports degraded state
		for {
			select {
			case <-ctx:
				return
			default:
				if c.isDetected(t) {
					close(ctx)
					return
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	select {
	case <-ctx:
		return true
	case <-timeoutChan:
		close(ctx)
		return false
	}
}

// waitForRecovery polls until system is healthy again
func (c *ChaosTest) waitForRecovery(t *testing.T, timeout time.Duration) bool {
	ctx := make(chan interface{})
	timeoutChan := time.After(timeout)

	go func() {
		for {
			select {
			case <-ctx:
				return
			default:
				if c.isRecovered(t) {
					close(ctx)
					return
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	select {
	case <-ctx:
		return true
	case <-timeoutChan:
		close(ctx)
		return false
	}
}

// isDetected checks if system has detected the failure (override per test)
func (c *ChaosTest) isDetected(t *testing.T) bool {
	// Default: check if health endpoint fails
	return !checkHealth(t, "http://localhost:8089/health")
}

// isRecovered checks if system has recovered (override per test)
func (c *ChaosTest) isRecovered(t *testing.T) bool {
	// Default: check if health endpoint passes
	return checkHealth(t, "http://localhost:8089/health")
}

// checkHealth performs a simple health check
func checkHealth(t *testing.T, url string) bool {
	client := NewHTTPClient(2 * time.Second)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// TestOSAOutage tests complete OSA container outage scenario
func TestOSAOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "OSA_Container_Outage",
		breakdown: func() error {
			t.Log("Stopping OSA container...")
			return StopService("osa")
		},
		recover: func() error {
			t.Log("Starting OSA container...")
			return StartService("osa")
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  180 * time.Second, // OSA takes ~120s to fully boot (observed)
	}

	chaos.Run(t)
}

// TestNetworkPartition tests network partition between services
func TestNetworkPartition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Network_Partition_OSA_BusinessOS",
		breakdown: func() error {
			t.Log("Injecting network partition...")
			return IsolateService("osa")
		},
		recover: func() error {
			t.Log("Restoring network connectivity...")
			return RestoreService("osa")
		},
		maxDetection: 10 * time.Second,
		maxRecovery:  30 * time.Second,
	}

	chaos.Run(t)
}
