package blockchain

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
)

const (
	percentFactor = 100
)

var (
	ErrNewEpochLessThanCurrent = errors.New("new epoch must be greater than current epoch")
)

// AttestationRecorder handles recording attestation verification outcomes.
type AttestationRecorder interface {
	// RecordSuccess records a successful attestation verification.
	RecordSuccess()
	// RecordFailure records a failed attestation verification with a reason.
	RecordFailure(reason string)
}

// AttestationStatsReader provides read access to attestation metrics.
type AttestationStatsReader interface {
	// GetCurrentMetrics returns the current verification statistics for the current epoch, without resetting.
	GetCurrentMetrics() AttestationMetrics
}

// AttestationMetricsCollector collects and provides attestation verification metrics.
type AttestationMetricsCollector interface {
	AttestationRecorder
	AttestationStatsReader

	// AdvanceEpoch finalizes metrics for the current epoch, resets counters, and sets the new current epoch.
	// It returns the metrics of the epoch that just ended.
	AdvanceEpoch(newEpoch primitives.Epoch) (AttestationMetrics, error)
}

// AttestationMetrics represents verification metrics for an epoch.
type AttestationMetrics struct {
	Epoch          primitives.Epoch
	Successes      uint64
	Failures       uint64
	SuccessRate    float64
	TotalProcessed uint64
	FailureReasons map[string]uint64
}

// metricsCollector is a concurrency-safe implementation of AttestationMetricsCollector.
type metricsCollector struct {
	mu             sync.Mutex
	successCount   uint64
	failureCount   uint64
	failureReasons map[string]uint64
	currentEpoch   primitives.Epoch
}

// NewMetricsCollector creates a new metrics collector instance.
func NewMetricsCollector() AttestationMetricsCollector {
	return &metricsCollector{
		failureReasons: make(map[string]uint64),
	}
}

// RecordSuccess increments the successful attestation counter thread-safely.
func (mc *metricsCollector) RecordSuccess() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.successCount++
}

// RecordFailure increments failure counter and records failure reason thread-safely.
func (mc *metricsCollector) RecordFailure(reason string) {
	if reason == "" {
		reason = "unknown"
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.failureCount++
	mc.failureReasons[reason]++
}

// GetCurrentMetrics returns the current statistics without resetting.
func (mc *metricsCollector) GetCurrentMetrics() AttestationMetrics {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.getCurrentMetricsLocked()
}

// AdvanceEpoch finalizes the current epoch's metrics, resets counters, and updates the current epoch.
func (mc *metricsCollector) AdvanceEpoch(newEpoch primitives.Epoch) (AttestationMetrics, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if newEpoch <= mc.currentEpoch {
		return AttestationMetrics{}, ErrNewEpochLessThanCurrent
	}

	currentMetrics := mc.getCurrentMetricsLocked()

	mc.reset()
	mc.currentEpoch = newEpoch

	return currentMetrics, nil
}

// getCurrentMetricsLocked is a helper method that assumes the lock is already held.
func (mc *metricsCollector) getCurrentMetricsLocked() AttestationMetrics {
	total := mc.successCount + mc.failureCount

	successRate := 0.0
	if total > 0 {
		successRate = float64(mc.successCount) / float64(total) * percentFactor
	}

	reasonsCopy := make(map[string]uint64, len(mc.failureReasons))
	for k, v := range mc.failureReasons {
		reasonsCopy[k] = v
	}

	return AttestationMetrics{
		Epoch:          mc.currentEpoch,
		Successes:      mc.successCount,
		Failures:       mc.failureCount,
		SuccessRate:    successRate,
		TotalProcessed: total,
		FailureReasons: reasonsCopy,
	}
}

// reset resets all counters and maps. Caller must hold write lock.
func (mc *metricsCollector) reset() {
	mc.successCount = 0
	mc.failureCount = 0
	mc.failureReasons = make(map[string]uint64)
}
