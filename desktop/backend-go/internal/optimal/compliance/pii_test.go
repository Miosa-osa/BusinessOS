package compliance

import (
	"strings"
	"testing"
)

func TestScanEmail(t *testing.T) {
	got := Scan("hit me at roberto@cliniciq.com please")
	if len(got) != 1 || got[0].Kind != KindEmail || got[0].Value != "roberto@cliniciq.com" {
		t.Fatalf("unexpected matches: %+v", got)
	}
}

func TestScanSSNFilter(t *testing.T) {
	// 000-12-3456 is invalid per SSA rules and must be rejected.
	if Any("ssn: 000-12-3456") {
		t.Fatal("expected invalid SSN to be rejected")
	}
	// 123-45-6789 is well-formed.
	if !Any("ssn: 123-45-6789") {
		t.Fatal("expected valid SSN to match")
	}
}

func TestScanCreditCardLuhn(t *testing.T) {
	// Known test VISA that satisfies Luhn: 4111 1111 1111 1111.
	text := "card 4111 1111 1111 1111 end"
	matches := Scan(text)
	found := false
	for _, m := range matches {
		if m.Kind == KindCreditCard {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a credit_card match, got %+v", matches)
	}
	// Random 16-digit number that fails Luhn.
	if Any("card 1234 5678 9012 3456 end-fail") {
		t.Fatal("expected Luhn-invalid number to be rejected")
	}
}

func TestRedactPlaceholder(t *testing.T) {
	text := "email roberto@cliniciq.com"
	out := RedactString(text, RedactOpts{})
	if !strings.Contains(out, "<REDACTED:email>") {
		t.Fatalf("expected placeholder substitution, got %q", out)
	}
}

func TestRedactMaskPreservesLength(t *testing.T) {
	text := "email roberto@cliniciq.com done"
	out := RedactString(text, RedactOpts{Strategy: StrategyMask})
	if len(out) != len(text) {
		t.Fatalf("mask should preserve length: got %d want %d", len(out), len(text))
	}
}

func TestRedactHashIsDeterministic(t *testing.T) {
	a := RedactString("ping roberto@cliniciq.com", RedactOpts{Strategy: StrategyHash})
	b := RedactString("ping roberto@cliniciq.com", RedactOpts{Strategy: StrategyHash})
	if a != b {
		t.Fatalf("hash strategy must be deterministic: %q vs %q", a, b)
	}
}

func TestRedactRemoveDropsMatch(t *testing.T) {
	out := RedactString("x 123-45-6789 y", RedactOpts{Strategy: StrategyRemove})
	if strings.Contains(out, "123-45-6789") {
		t.Fatalf("remove strategy left PII behind: %q", out)
	}
}

func TestRedactExceptKeepsEmails(t *testing.T) {
	text := "roberto@cliniciq.com and 4111 1111 1111 1111"
	out := RedactString(text, RedactOpts{Except: []Kind{KindEmail}})
	if !strings.Contains(out, "roberto@cliniciq.com") {
		t.Fatalf("email should have been preserved via Except: %q", out)
	}
}

func TestKindsPresent(t *testing.T) {
	text := "roberto@cliniciq.com ip 192.168.1.1"
	kinds := KindsPresent(text)
	want := map[Kind]bool{KindEmail: true, KindIPv4: true}
	for _, k := range kinds {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Fatalf("missing kinds: %+v (got %+v)", want, kinds)
	}
}
