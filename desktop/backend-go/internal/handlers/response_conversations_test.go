package handlers

import (
	"testing"
	"time"
)

func TestConversationTimestampUsesDatabaseZone(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	decoded := time.Date(2026, 9, 30, 12, 11, 0, 0, time.UTC)
	got := formatConversationTime(decoded, zone)
	if got != "2026-09-30T12:11:00-04:00" {
		t.Fatalf("incorrect wall clock conversion: %s", got)
	}
}
