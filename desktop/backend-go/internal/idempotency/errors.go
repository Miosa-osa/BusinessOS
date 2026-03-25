package idempotency

import "errors"

var (
	// ErrEmptyKey is returned when an empty idempotency key is provided.
	ErrEmptyKey = errors.New("idempotency key cannot be empty")
)
