package container

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// ContainerMonitor monitors container health, metrics, and performs cleanup
type ContainerMonitor struct {
	manager        *ContainerManager
	config         *MonitorConfig
	metrics        *ContainerMetrics
	stopChan       chan struct{}
	wg             sync.WaitGroup
	containerStats map[string]*ContainerStats
	statsMutex     sync.RWMutex
}

// MonitorConfig defines monitoring behavior
type MonitorConfig struct {
	IdleTimeout         time.Duration // Default: 30m
	CleanupInterval     time.Duration // Default: 5m
	HealthCheckInterval time.Duration // Default: 30s
	MaxMemoryBytes      int64         // Default: 512MB
	MaxCPUPercent       float64       // Default: 50%
}

// ContainerStats tracks per-container statistics
type ContainerStats struct {
	ContainerID     string
	UserID          string
	State           string
	LastActivity    time.Time
	LastHealthCheck time.Time
	MemoryUsage     uint64
	CPUPercent      float64
	IsHealthy       bool
	IsZombie        bool
	HealthErrors    int
}

// ContainerMetrics tracks global metrics with atomic counters
type ContainerMetrics struct {
	mu               sync.RWMutex
	ActiveContainers int64
	TotalStarted     int64
	TotalStopped     int64
	TotalErrors      int64
	OrphansRemoved   int64
	IdleRemoved      int64
	LastCleanup      time.Time
	MonitorStartTime time.Time
}

// DefaultMonitorConfig returns default monitoring configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		IdleTimeout:         30 * time.Minute,
		CleanupInterval:     5 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		MaxMemoryBytes:      512 * 1024 * 1024, // 512MB
		MaxCPUPercent:       50.0,
	}
}

// NewContainerMonitor creates a new container monitor
func NewContainerMonitor(manager *ContainerManager, config *MonitorConfig) *ContainerMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	return &ContainerMonitor{
		manager:        manager,
		config:         config,
		metrics:        NewContainerMetrics(),
		stopChan:       make(chan struct{}),
		containerStats: make(map[string]*ContainerStats),
	}
}

// NewContainerMetrics creates a new metrics instance
func NewContainerMetrics() *ContainerMetrics {
	return &ContainerMetrics{
		MonitorStartTime: time.Now(),
	}
}

// StartMonitoring starts all monitoring goroutines
func (cm *ContainerMonitor) StartMonitoring(ctx context.Context) error {
	log.Printf("[Monitor] Starting container monitoring (health_check=%v, cleanup=%v, idle_timeout=%v)",
		cm.config.HealthCheckInterval, cm.config.CleanupInterval, cm.config.IdleTimeout)

	// Start health check loop
	cm.wg.Add(1)
	go cm.healthCheckLoop(ctx)

	// Start cleanup loop
	cm.wg.Add(1)
	go cm.cleanupLoop(ctx)

	// Perform initial orphan cleanup
	go func() {
		if err := cm.CleanupOrphans(ctx); err != nil {
			log.Printf("[Monitor] Error: Initial orphan cleanup failed: %v", err)
		}
	}()

	return nil
}

// StopMonitoring stops all monitoring goroutines
func (cm *ContainerMonitor) StopMonitoring() error {
	log.Printf("[Monitor] Stopping container monitoring")
	close(cm.stopChan)
	cm.wg.Wait()
	log.Printf("[Monitor] Container monitoring stopped")
	return nil
}

// GetMetrics returns current metrics snapshot
func (cm *ContainerMonitor) GetMetrics() *ContainerMetrics {
	cm.metrics.mu.RLock()
	defer cm.metrics.mu.RUnlock()

	// Update active container count
	cm.manager.mu.RLock()
	activeCount := int64(len(cm.manager.containers))
	cm.manager.mu.RUnlock()

	atomic.StoreInt64(&cm.metrics.ActiveContainers, activeCount)

	return cm.metrics
}

// GetContainerStats returns stats for a specific container
func (cm *ContainerMonitor) GetContainerStats(containerID string) (*ContainerStats, error) {
	cm.statsMutex.RLock()
	defer cm.statsMutex.RUnlock()

	stats, exists := cm.containerStats[containerID]
	if !exists {
		return nil, fmt.Errorf("stats not found for container %s", containerID)
	}

	return stats, nil
}

// healthCheckLoop performs periodic health checks on all containers
func (cm *ContainerMonitor) healthCheckLoop(ctx context.Context) {
	defer cm.wg.Done()

	ticker := time.NewTicker(cm.config.HealthCheckInterval)
	defer ticker.Stop()

	log.Printf("[Monitor] Health check loop started (interval=%v)", cm.config.HealthCheckInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Monitor] Health check loop stopped (context cancelled)")
			return
		case <-cm.stopChan:
			log.Printf("[Monitor] Health check loop stopped")
			return
		case <-ticker.C:
			cm.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks checks health of all managed containers
func (cm *ContainerMonitor) performHealthChecks(ctx context.Context) {
	cm.manager.mu.RLock()
	containerIDs := make([]string, 0, len(cm.manager.containers))
	for id := range cm.manager.containers {
		containerIDs = append(containerIDs, id)
	}
	cm.manager.mu.RUnlock()

	for _, containerID := range containerIDs {
		cm.performHealthCheck(ctx, containerID)
	}
}

// performHealthCheck checks health of a single container
func (cm *ContainerMonitor) performHealthCheck(ctx context.Context, containerID string) {
	// Get container info from manager
	cm.manager.mu.RLock()
	containerInfo, exists := cm.manager.containers[containerID]
	cm.manager.mu.RUnlock()

	if !exists {
		return
	}

	// Inspect container
	inspect, err := cm.manager.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Printf("[Monitor] Warning: Health check failed to inspect container %s: %v", containerID[:12], err)
		cm.updateContainerStats(containerID, containerInfo.UserID, func(stats *ContainerStats) {
			stats.IsHealthy = false
			stats.HealthErrors++
			stats.State = "error"
		})
		cm.metrics.IncrementErrors()
		return
	}

	// Get container stats
	statsResponse, err := cm.manager.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		log.Printf("[Monitor] Warning: Health check failed to get stats for container %s: %v", containerID[:12], err)
		cm.updateContainerStats(containerID, containerInfo.UserID, func(stats *ContainerStats) {
			stats.IsHealthy = inspect.State.Running
			stats.State = inspect.State.Status
			stats.LastHealthCheck = time.Now()
		})
		return
	}
	defer statsResponse.Body.Close()

	// Parse stats
	var dockerStats container.StatsResponse
	if err := json.NewDecoder(statsResponse.Body).Decode(&dockerStats); err != nil {
		log.Printf("[Monitor] Warning: Failed to decode container stats for %s: %v", containerID[:12], err)
		return
	}

	// Calculate CPU percentage
	cpuPercent := calculateCPUPercent(&dockerStats)
	memoryUsage := dockerStats.MemoryStats.Usage

	// Update stats
	cm.updateContainerStats(containerID, containerInfo.UserID, func(stats *ContainerStats) {
		stats.State = inspect.State.Status
		stats.IsHealthy = inspect.State.Running
		stats.MemoryUsage = memoryUsage
		stats.CPUPercent = cpuPercent
		stats.LastHealthCheck = time.Now()
		stats.HealthErrors = 0

		// Check for zombie state (not running but exists)
		if !inspect.State.Running && inspect.State.Status != "created" {
			stats.IsZombie = true
		}

		// Check resource limits
		if memoryUsage > uint64(cm.config.MaxMemoryBytes) {
			log.Printf("[Monitor] Warning: Container %s exceeding memory limit (usage=%d, limit=%d)",
				containerID[:12], memoryUsage, cm.config.MaxMemoryBytes)
		}

		if cpuPercent > cm.config.MaxCPUPercent {
			log.Printf("[Monitor] Warning: Container %s exceeding CPU limit (usage=%.2f%%, limit=%.2f%%)",
				containerID[:12], cpuPercent, cm.config.MaxCPUPercent)
		}
	})
}

// cleanupLoop performs periodic cleanup of idle and orphaned containers
func (cm *ContainerMonitor) cleanupLoop(ctx context.Context) {
	defer cm.wg.Done()

	ticker := time.NewTicker(cm.config.CleanupInterval)
	defer ticker.Stop()

	log.Printf("[Monitor] Cleanup loop started (interval=%v)", cm.config.CleanupInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Monitor] Cleanup loop stopped (context cancelled)")
			return
		case <-cm.stopChan:
			log.Printf("[Monitor] Cleanup loop stopped")
			return
		case <-ticker.C:
			cm.performCleanup(ctx)
		}
	}
}

// performCleanup removes idle containers and orphans
func (cm *ContainerMonitor) performCleanup(ctx context.Context) {
	log.Printf("[Monitor] Starting cleanup cycle")

	cleanupStart := time.Now()
	idleRemoved := 0
	orphansRemoved := 0

	// Cleanup idle containers
	cm.statsMutex.RLock()
	idleContainers := make([]string, 0)
	for containerID, stats := range cm.containerStats {
		if time.Since(stats.LastActivity) > cm.config.IdleTimeout {
			idleContainers = append(idleContainers, containerID)
		}
	}
	cm.statsMutex.RUnlock()

	for _, containerID := range idleContainers {
		idleTime := time.Since(cm.containerStats[containerID].LastActivity)
		log.Printf("[Monitor] Removing idle container %s (idle for %v)", containerID[:12], idleTime)

		if err := cm.manager.RemoveContainer(containerID, true); err != nil {
			log.Printf("[Monitor] Error: Failed to remove idle container %s: %v", containerID[:12], err)
			cm.metrics.IncrementErrors()
		} else {
			idleRemoved++
			cm.metrics.IncrementStopped()
			atomic.AddInt64(&cm.metrics.IdleRemoved, 1)

			// Remove from stats
			cm.statsMutex.Lock()
			delete(cm.containerStats, containerID)
			cm.statsMutex.Unlock()
		}
	}

	// Cleanup orphaned containers
	if err := cm.CleanupOrphans(ctx); err != nil {
		log.Printf("[Monitor] Error: Orphan cleanup failed: %v", err)
	} else {
		// Count orphans removed this cycle
		currentOrphans := atomic.LoadInt64(&cm.metrics.OrphansRemoved)
		orphansRemoved = int(currentOrphans)
	}

	// Update metrics
	cm.metrics.mu.Lock()
	cm.metrics.LastCleanup = time.Now()
	cm.metrics.mu.Unlock()

	log.Printf("[Monitor] Cleanup cycle completed (idle_removed=%d, orphans_removed=%d, duration=%v)",
		idleRemoved, orphansRemoved, time.Since(cleanupStart))
}

// CleanupOrphans removes containers with BusinessOS label not in manager's map
func (cm *ContainerMonitor) CleanupOrphans(ctx context.Context) error {
	// List all containers with businessos label
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "app=businessos")

	containers, err := cm.manager.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	orphanCount := 0
	cm.manager.mu.RLock()
	for _, container := range containers {
		// Check if container is not in manager's map
		if _, exists := cm.manager.containers[container.ID]; !exists {
			cm.manager.mu.RUnlock()

			log.Printf("[Monitor] Found orphaned container %s (names=%v)", container.ID[:12], container.Names)

			// Remove orphaned container
			if err := cm.manager.RemoveContainer(container.ID, true); err != nil {
				log.Printf("[Monitor] Error: Failed to remove orphaned container %s: %v", container.ID[:12], err)
				cm.metrics.IncrementErrors()
			} else {
				orphanCount++
				atomic.AddInt64(&cm.metrics.OrphansRemoved, 1)
			}

			cm.manager.mu.RLock()
		}
	}
	cm.manager.mu.RUnlock()

	if orphanCount > 0 {
		log.Printf("[Monitor] Orphan cleanup completed (orphans_removed=%d)", orphanCount)
	}

	return nil
}

// updateContainerStats updates or creates stats for a container
func (cm *ContainerMonitor) updateContainerStats(containerID, userID string, updateFn func(*ContainerStats)) {
	cm.statsMutex.Lock()
	defer cm.statsMutex.Unlock()

	stats, exists := cm.containerStats[containerID]
	if !exists {
		stats = &ContainerStats{
			ContainerID:  containerID,
			UserID:       userID,
			LastActivity: time.Now(),
		}
		cm.containerStats[containerID] = stats
	}

	updateFn(stats)
}

// UpdateActivity updates the last activity time for a container
func (cm *ContainerMonitor) UpdateActivity(containerID string) {
	cm.statsMutex.Lock()
	defer cm.statsMutex.Unlock()

	if stats, exists := cm.containerStats[containerID]; exists {
		stats.LastActivity = time.Now()
	}
}

// calculateCPUPercent calculates CPU usage percentage
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		return (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return 0.0
}

// Metrics atomic operations

// IncrementStarted increments the total started counter
func (m *ContainerMetrics) IncrementStarted() {
	atomic.AddInt64(&m.TotalStarted, 1)
}

// IncrementStopped increments the total stopped counter
func (m *ContainerMetrics) IncrementStopped() {
	atomic.AddInt64(&m.TotalStopped, 1)
}

// IncrementErrors increments the total errors counter
func (m *ContainerMetrics) IncrementErrors() {
	atomic.AddInt64(&m.TotalErrors, 1)
}

// ToJSON converts metrics to JSON-serializable map
func (m *ContainerMetrics) ToJSON() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_containers": atomic.LoadInt64(&m.ActiveContainers),
		"total_started":     atomic.LoadInt64(&m.TotalStarted),
		"total_stopped":     atomic.LoadInt64(&m.TotalStopped),
		"total_errors":      atomic.LoadInt64(&m.TotalErrors),
		"orphans_removed":   atomic.LoadInt64(&m.OrphansRemoved),
		"idle_removed":      atomic.LoadInt64(&m.IdleRemoved),
		"last_cleanup":      m.LastCleanup.Format(time.RFC3339),
		"uptime_seconds":    time.Since(m.MonitorStartTime).Seconds(),
	}
}

// GetAllContainerStats returns stats for all containers
func (cm *ContainerMonitor) GetAllContainerStats() map[string]*ContainerStats {
	cm.statsMutex.RLock()
	defer cm.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := make(map[string]*ContainerStats, len(cm.containerStats))
	for id, stats := range cm.containerStats {
		statsCopy[id] = &ContainerStats{
			ContainerID:     stats.ContainerID,
			UserID:          stats.UserID,
			State:           stats.State,
			LastActivity:    stats.LastActivity,
			LastHealthCheck: stats.LastHealthCheck,
			MemoryUsage:     stats.MemoryUsage,
			CPUPercent:      stats.CPUPercent,
			IsHealthy:       stats.IsHealthy,
			IsZombie:        stats.IsZombie,
			HealthErrors:    stats.HealthErrors,
		}
	}

	return statsCopy
}

// RegisterContainer registers a new container with the monitor
func (cm *ContainerMonitor) RegisterContainer(containerID, userID string) {
	cm.statsMutex.Lock()
	defer cm.statsMutex.Unlock()

	cm.containerStats[containerID] = &ContainerStats{
		ContainerID:  containerID,
		UserID:       userID,
		State:        "created",
		LastActivity: time.Now(),
		IsHealthy:    true,
	}

	cm.metrics.IncrementStarted()
}

// UnregisterContainer removes a container from monitoring
func (cm *ContainerMonitor) UnregisterContainer(containerID string) {
	cm.statsMutex.Lock()
	defer cm.statsMutex.Unlock()

	delete(cm.containerStats, containerID)
	cm.metrics.IncrementStopped()
}
