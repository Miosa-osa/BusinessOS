package chaos

import (
	"testing"
	"time"
)

// TestPostgresOutage tests PostgreSQL container outage scenario
// This is a simpler test that uses Postgres which we know runs reliably
func TestPostgresOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Postgres_Container_Outage",
		breakdown: func() error {
			t.Log("Stopping PostgreSQL container...")
			return StopService("postgres")
		},
		recover: func() error {
			t.Log("Starting PostgreSQL container...")
			return StartService("postgres")
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  120 * time.Second, // Postgres takes ~92s to recover (observed increasing pattern)
	}

	chaos.Run(t)
}

// TestRedisOutage tests Redis container outage scenario
func TestRedisOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Redis_Container_Outage",
		breakdown: func() error {
			t.Log("Stopping Redis container...")
			return StopService("redis")
		},
		recover: func() error {
			t.Log("Starting Redis container...")
			return StartService("redis")
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  15 * time.Second,
	}

	chaos.Run(t)
}

// TestPm4pyRustOutage tests pm4py-rust container outage scenario
func TestPm4pyRustOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Pm4pyRust_Container_Outage",
		breakdown: func() error {
			t.Log("Stopping pm4py-rust container...")
			return StopService("pm4py-rust")
		},
		recover: func() error {
			t.Log("Starting pm4py-rust container...")
			return StartService("pm4py-rust")
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  20 * time.Second,
	}

	chaos.Run(t)
}

// TestMultipleServiceOutage tests cascading failures across multiple services
func TestMultipleServiceOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Multiple_Service_Outage_Postgres_Redis",
		breakdown: func() error {
			t.Log("Stopping PostgreSQL and Redis containers...")
			if err := StopService("postgres"); err != nil {
				return err
			}
			if err := StopService("redis"); err != nil {
				return err
			}
			return nil
		},
		recover: func() error {
			t.Log("Starting PostgreSQL and Redis containers...")
			if err := StartService("postgres"); err != nil {
				return err
			}
			if err := StartService("redis"); err != nil {
				return err
			}
			return nil
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  45 * time.Second,
	}

	chaos.Run(t)
}

// TestBackendOutage tests BusinessOS backend container outage scenario
func TestBackendOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	chaos := &ChaosTest{
		name: "Backend_Container_Outage",
		breakdown: func() error {
			t.Log("Stopping BusinessOS backend container...")
			return StopService("backend")
		},
		recover: func() error {
			t.Log("Starting BusinessOS backend container...")
			return StartService("backend")
		},
		maxDetection: 5 * time.Second,
		maxRecovery:  30 * time.Second,
	}

	chaos.Run(t)
}
