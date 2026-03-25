package circuitbreaker

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for resilience
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	maxAttempts      int
	baseDelay        time.Duration
	maxDelay         time.Duration
	timeoutDuration  time.Duration
	cooldownPeriod   time.Duration
	halfOpenMaxCalls int

	// State
	state           State
	lastFailure      time.Time
	consecutiveFailures int
	successCount     int

	// Callbacks
	onStateChange    func(oldState, newState State)
	onFailure        func(error)
	onSuccess        func()
	onTimeout        func()

	// Monitoring
	totalCalls       int64
	successfulCalls  int64
	failedCalls      int64
	timeoutCalls     int64
}

// Config holds circuit breaker configuration
type Config struct {
	MaxAttempts      int           // Maximum consecutive failures before opening circuit
	BaseDelay        time.Duration // Base delay for exponential backoff
	MaxDelay         time.Duration // Maximum delay backoff
	TimeoutDuration  time.Duration // Request timeout duration
	CooldownPeriod   time.Duration // How long to stay in open state
	HalfOpenMaxCalls int           // Max calls in half-open state before closing
}

// NewCircuitBreaker creates a new circuit breaker with default configuration
func NewCircuitBreaker(config Config) *CircuitBreaker {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 5
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 10 * time.Second
	}
	if config.TimeoutDuration <= 0 {
		config.TimeoutDuration = 5 * time.Second
	}
	if config.CooldownPeriod <= 0 {
		config.CooldownPeriod = 30 * time.Second
	}
	if config.HalfOpenMaxCalls <= 0 {
		config.HalfOpenMaxCalls = 3
	}

	return &CircuitBreaker{
		maxAttempts:      config.MaxAttempts,
		baseDelay:       config.BaseDelay,
		maxDelay:        config.MaxDelay,
		timeoutDuration: config.TimeoutDuration,
		cooldownPeriod:  config.CooldownPeriod,
		halfOpenMaxCalls: config.HalfOpenMaxCalls,
		state:          StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.totalCalls++

	// Check if we should allow the call
	if !cb.allowCall() {
		cb.failedCalls++
		if cb.onStateChange != nil {
			cb.onStateChange(cb.state, cb.state)
		}
		return ErrCircuitOpen
	}

	// Execute the function with timeout
	err := cb.executeWithTimeout(ctx, fn)

	cb.recordResult(err)
	return err
}

// ExecuteWithFallback executes a function with fallback when circuit is open
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	err := cb.Execute(ctx, fn)
	if err == ErrCircuitOpen {
		return cb.executeWithTimeout(ctx, fallback)
	}
	return err
}

// allowCall determines if a call should be allowed based on current state
func (cb *CircuitBreaker) allowCall() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if cooldown period has passed
		return time.Since(cb.lastFailure) >= cb.cooldownPeriod
	case StateHalfOpen:
		// Limit calls in half-open state
		return cb.successCount < cb.halfOpenMaxCalls
	default:
		return false
	}
}

// executeWithTimeout executes a function with context timeout
func (cb *CircuitBreaker) executeWithTimeout(ctx context.Context, fn func() error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cb.timeoutDuration)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-timeoutCtx.Done():
		cb.timeoutCalls++
		return ErrTimeout
	}
}

// recordResult records the result of an execution and updates state
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.handleSuccess()
	} else {
		cb.handleFailure(err)
	}
}

// handleSuccess handles successful execution
func (cb *CircuitBreaker) handleSuccess() {
	cb.successfulCalls++
	cb.successCount++
	cb.consecutiveFailures = 0

	// Transition from half-open to closed
	if cb.state == StateHalfOpen {
		cb.transition(StateClosed, StateHalfOpen)
		cb.successCount = 0 // Reset for next half-open period
	}

	if cb.onSuccess != nil {
		cb.onSuccess()
	}
}

// handleFailure handles failed execution
func (cb *CircuitBreaker) handleFailure(err error) {
	cb.failedCalls++
	cb.consecutiveFailures++
	cb.lastFailure = time.Now()

	if cb.onFailure != nil {
		cb.onFailure(err)
	}

	// Check if we should open the circuit
	if cb.consecutiveFailures >= cb.maxAttempts && cb.state != StateOpen {
		cb.transition(StateOpen, cb.state)
		cb.consecutiveFailures = 0
	}

	// Transition from half-open back to open on failure
	if cb.state == StateHalfOpen {
		cb.transition(StateOpen, StateHalfOpen)
	}
}

// transition changes the circuit breaker state
func (cb *CircuitBreaker) transition(newState State, oldState State) {
	if newState != oldState {
		cb.state = newState
		if cb.onStateChange != nil {
			cb.onStateChange(oldState, newState)
		}
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var successRate float64
	if cb.totalCalls > 0 {
		successRate = float64(cb.successfulCalls) / float64(cb.totalCalls) * 100
	}

	return Stats{
		TotalCalls:       cb.totalCalls,
		SuccessfulCalls: cb.successfulCalls,
		FailedCalls:      cb.failedCalls,
		TimeoutCalls:     cb.timeoutCalls,
		SuccessRate:      successRate,
		State:           cb.state,
		LastFailure:      cb.lastFailure,
		ConsecutiveFailures: cb.consecutiveFailures,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transition(StateClosed, cb.state)
	cb.consecutiveFailures = 0
	cb.successCount = 0
	cb.lastFailure = time.Time{}
}

// OnStateChange sets a callback for state changes
func (cb *CircuitBreaker) OnStateChange(callback func(oldState, newState State)) {
	cb.onStateChange = callback
}

// OnFailure sets a callback for failures
func (cb *CircuitBreaker) OnFailure(callback func(error)) {
	cb.onFailure = callback
}

// OnSuccess sets a callback for successes
func (cb *CircuitBreaker) OnSuccess(callback func()) {
	cb.onSuccess = callback
}

// OnTimeout sets a callback for timeouts
func (cb *CircuitBreaker) OnTimeout(callback func()) {
	cb.onTimeout = callback
}

// GetNextRetryDelay returns the delay before next retry based on current state
func (cb *CircuitBreaker) GetNextRetryDelay() time.Duration {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return 0
	case StateOpen:
		// Exponential backoff with jitter
		attempts := float64(cb.consecutiveFailures)
		delay := float64(cb.baseDelay) * math.Pow(2.0, attempts)
		delay = math.Min(delay, float64(cb.maxDelay))

		// Add jitter to avoid thundering herd
		jitter := delay * 0.2 * (rand.Float64() - 0.5)
		return time.Duration(delay + jitter)
	case StateHalfOpen:
		return cb.baseDelay / 2 // Faster retry in half-open state
	default:
		return cb.baseDelay
	}
}

// Stats contains circuit breaker statistics
type Stats struct {
	TotalCalls         int64
	SuccessfulCalls    int64
	FailedCalls        int64
	TimeoutCalls       int64
	SuccessRate        float64
	State             State
	LastFailure        time.Time
	ConsecutiveFailures int
}

// Error types
var (
	ErrCircuitOpen = NewCircuitBreakerError("circuit is open")
	ErrTimeout     = NewCircuitBreakerError("request timeout")
)

// CircuitBreakerError represents circuit breaker errors
type CircuitBreakerError struct {
	message string
}

func NewCircuitBreakerError(message string) *CircuitBreakerError {
	return &CircuitBreakerError{message: message}
}

func (e *CircuitBreakerError) Error() string {
	return e.message
}

// IsCircuitOpenError checks if error is circuit open error
func IsCircuitOpenError(err error) bool {
	return err == ErrCircuitOpen
}

// IsTimeoutError checks if error is timeout error
func IsTimeoutError(err error) bool {
	return err == ErrTimeout
}

// Builder for fluent configuration
type CircuitBreakerBuilder struct {
	cb *CircuitBreaker
}

// NewBuilder creates a new circuit breaker builder
func NewBuilder() *CircuitBreakerBuilder {
	return &CircuitBreakerBuilder{
		cb: &CircuitBreaker{
			state: StateClosed,
		},
	}
}

// WithConfig sets the circuit breaker configuration
func (b *CircuitBreakerBuilder) WithConfig(config Config) *CircuitBreakerBuilder {
	b.cb.maxAttempts = config.MaxAttempts
	b.cb.baseDelay = config.BaseDelay
	b.cb.maxDelay = config.MaxDelay
	b.cb.timeoutDuration = config.TimeoutDuration
	b.cb.cooldownPeriod = config.CooldownPeriod
	b.cb.halfOpenMaxCalls = config.HalfOpenMaxCalls
	return b
}

// WithMaxAttempts sets the maximum number of attempts
func (b *CircuitBreakerBuilder) WithMaxAttempts(maxAttempts int) *CircuitBreakerBuilder {
	b.cb.maxAttempts = maxAttempts
	return b
}

// WithBackoff sets the backoff configuration
func (b *CircuitBreakerBuilder) WithBackoff(baseDelay, maxDelay time.Duration) *CircuitBreakerBuilder {
	b.cb.baseDelay = baseDelay
	b.cb.maxDelay = maxDelay
	return b
}

// WithTimeout sets the request timeout
func (b *CircuitBreakerBuilder) WithTimeout(timeout time.Duration) *CircuitBreakerBuilder {
	b.cb.timeoutDuration = timeout
	return b
}

// WithCooldown sets the cooldown period
func (b *CircuitBreakerBuilder) WithCooldown(cooldown time.Duration) *CircuitBreakerBuilder {
	b.cb.cooldownPeriod = cooldown
	return b
}

// Build creates the circuit breaker
func (b *CircuitBreakerBuilder) Build() *CircuitBreaker {
	return b.cb
}

// Pre-defined circuit breaker configurations for different use cases

// ComplianceServiceConfig returns circuit breaker config for compliance service
func ComplianceServiceConfig() Config {
	return Config{
		MaxAttempts:      3,  // Fewer attempts for compliance (critical service)
		BaseDelay:       1 * time.Second,
		MaxDelay:        30 * time.Second,
		TimeoutDuration: 5 * time.Second,  // OSA call timeout
		CooldownPeriod:  60 * time.Second,
		HalfOpenMaxCalls: 1,  // Only try once in half-open state
	}
}

// DatabaseConfig returns circuit breaker config for database operations
func DatabaseConfig() Config {
	return Config{
		MaxAttempts:      5,  // Standard database retry
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		TimeoutDuration: 2 * time.Second,
		CooldownPeriod:  10 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// ExternalAPIConfig returns circuit breaker config for external API calls
func ExternalAPIConfig() Config {
	return Config{
		MaxAttempts:      5,
		BaseDelay:       200 * time.Millisecond,
		MaxDelay:        10 * time.Second,
		TimeoutDuration: 10 * time.Second,
		CooldownPeriod:  30 * time.Second,
		HalfOpenMaxCalls: 2,
	}
}