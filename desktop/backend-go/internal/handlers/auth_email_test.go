package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// normaliseEmail tests
// ---------------------------------------------------------------------------

func TestNormaliseEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "user@example.com", "user@example.com"},
		{"uppercase", "USER@Example.COM", "user@example.com"},
		{"mixed case", "UsEr@ExAmPlE.CoM", "user@example.com"},
		{"leading space", " user@example.com", "user@example.com"},
		{"trailing space", "user@example.com ", "user@example.com"},
		{"both spaces", " USER@EXAMPLE.COM ", "user@example.com"},
		{"tabs", "\tuser@example.com\t", "user@example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normaliseEmail(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// validatePassword tests
// ---------------------------------------------------------------------------

func TestValidatePassword_Valid(t *testing.T) {
	// Meets all requirements
	msg := validatePassword("SecureP@ss1")
	assert.Empty(t, msg)
}

func TestValidatePassword_TooShort(t *testing.T) {
	msg := validatePassword("Ab1!")
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "at least 8 characters")
}

func TestValidatePassword_TooLong(t *testing.T) {
	long := make([]byte, 129)
	for i := range long {
		long[i] = 'A'
	}
	msg := validatePassword(string(long))
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "not exceed 128 characters")
}

func TestValidatePassword_MissingUppercase(t *testing.T) {
	msg := validatePassword("lowercase1!")
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "uppercase")
}

func TestValidatePassword_MissingLowercase(t *testing.T) {
	msg := validatePassword("UPPERCASE1!")
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "lowercase")
}

func TestValidatePassword_MissingDigit(t *testing.T) {
	msg := validatePassword("NoDigits!")
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "digit")
}

func TestValidatePassword_MissingSpecial(t *testing.T) {
	msg := validatePassword("NoSpecial1")
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "special character")
}

func TestValidatePassword_Empty(t *testing.T) {
	msg := validatePassword("")
	assert.NotEmpty(t, msg)
}

func TestValidatePassword_CommonPatterns(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"strong password", "MyStr0ng!Pass", true},
		{"with spaces", "My Str0ng! Pass", true},
		{"unicode special", "P@ssw0rd\u00A9", true},
		{"only letters", "abcdefgh", false},
		{"only numbers", "12345678", false},
		{"only special", "!@#$%^&*", false},
		{"missing digit", "Password!", false},
		{"missing uppercase", "password1!", false},
		{"missing lowercase", "PASSWORD1!", false},
		{"missing special", "Password1", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := validatePassword(tc.password)
			if tc.valid {
				assert.Empty(t, msg, "password %q should be valid", tc.password)
			} else {
				assert.NotEmpty(t, msg, "password %q should be invalid", tc.password)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Login lockout logic tests
// ---------------------------------------------------------------------------

func TestIsLockedOut_Initial(t *testing.T) {
	// A fresh email should not be locked out
	assert.False(t, isLockedOut("fresh@example.com"))
}

func TestRecordFailedAttempt_BelowThreshold(t *testing.T) {
	// Reset any previous state
	resetLoginAttempts("below@example.com")

	for i := 0; i < loginMaxAttempts-1; i++ {
		locked := recordFailedAttempt("below@example.com")
		assert.False(t, locked, "should not lock on attempt %d", i+1)
	}

	// Should not be locked yet
	assert.False(t, isLockedOut("below@example.com"))

	// Clean up
	resetLoginAttempts("below@example.com")
}

func TestRecordFailedAttempt_AtThreshold(t *testing.T) {
	resetLoginAttempts("at@example.com")

	// Record loginMaxAttempts failures
	var locked bool
	for i := 0; i < loginMaxAttempts; i++ {
		locked = recordFailedAttempt("at@example.com")
	}

	assert.True(t, locked, "should be locked after %d failures", loginMaxAttempts)
	assert.True(t, isLockedOut("at@example.com"))

	// Clean up
	resetLoginAttempts("at@example.com")
}

func TestResetLoginAttempts(t *testing.T) {
	resetLoginAttempts("reset@example.com")

	// Record some failures
	for i := 0; i < 3; i++ {
		recordFailedAttempt("reset@example.com")
	}

	// Reset
	resetLoginAttempts("reset@example.com")

	// Should not be locked
	assert.False(t, isLockedOut("reset@example.com"))
}

func TestLoginMaxAttempts_Constant(t *testing.T) {
	assert.Equal(t, 5, loginMaxAttempts)
}

func TestLoginLockDuration_Constant(t *testing.T) {
	// 15 minutes
	assert.Equal(t, 15*60*1000*1000*1000, loginLockDuration.Nanoseconds())
}

// ---------------------------------------------------------------------------
// Request struct validation tests
// ---------------------------------------------------------------------------

func TestSignUpRequest_Fields(t *testing.T) {
	req := SignUpRequest{
		Email:    "test@example.com",
		Password: "SecureP@ss1",
		Name:     "Test User",
	}
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "SecureP@ss1", req.Password)
	assert.Equal(t, "Test User", req.Name)
}

func TestSignInRequest_Fields(t *testing.T) {
	req := SignInRequest{
		Email:    "test@example.com",
		Password: "SecureP@ss1",
	}
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "SecureP@ss1", req.Password)
}
