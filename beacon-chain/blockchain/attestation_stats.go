package blockchain

import (
	"sync"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
)

const (
	percentFactor = 100
)

// EpochSummary represents a summary of attestation stats for a single epoch.
type EpochSummary struct {
	Epoch          primitives.Epoch
	Successes      uint64
	Failures       uint64
	SuccessRate    float64
	TotalProcessed uint64
	FailureReasons map[string]uint64
	ResetOccurred  bool
}

// StatsCollector defines the interface for collecting attestation statistics.
type StatsCollector interface {
	// recordSuccess records a successful attestation verification.
	recordSuccess()
	// recordFailure records a failed attestation verification with a reason.
	recordFailure(reason string)
	// getStats returns current statistics about attestation verifications.
	getStats() (uint64, uint64, map[string]uint64)
}

// attestationStats handles attestation verification statistics collection.
// It keeps track of successful and failed attestations, along with failure reasonsddoc
// and provides thread-safe access to these statistics.
type attestationStats struct {
	mu             sync.RWMutex
	successCount   uint64
	failureCount   uint64
	failureReasons map[string]uint64
	lastEpoch      primitives.Epoch
}

// New creates a new attestationStats instance.
func newAttestationStats() *attestationStats {
	return &attestationStats{
		failureReasons: make(map[string]uint64),
	}
}

// recordSuccess increments the successful attestation counter and updates related metrics.
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (as *attestationStats) recordSuccess() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.successCount++
}

// recordFailure increments the failed attestation counter and records the failure reason.
// This method is thread-safe and can be called concurrently from multiple goroutines.
// If an empty reason is provided, it will be recorded as "unknown".
// Parameters:
//   - reason: The reason for the attestation verification failure
func (as *attestationStats) recordFailure(reason string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if reason == "" {
		reason = "unknown"
	}

	as.failureCount++
	as.failureReasons[reason]++
}

// getStats returns the current attestation verification statistics.
// The returned map is a copy of the internal failure reasons map, making it safe for concurrent access.
// Returns:
//   - successCount: Number of successful attestation verifications
//   - failureCount: Number of failed attestation verifications
//   - failureReasons: Map of failure reasons and their counts
func (as *attestationStats) getStats() (uint64, uint64, map[string]uint64) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	failureCopy := make(map[string]uint64, len(as.failureReasons))

	for k, v := range as.failureReasons {
		failureCopy[k] = v
	}

	return as.successCount, as.failureCount, failureCopy
}

// outputEpochSummary computes and returns the statistics for the given epoch.
// If currentEpoch > as.lastEpoch, the internal counters are reset.
func (as *attestationStats) outputEpochSummary(currentEpoch primitives.Epoch) EpochSummary {
	as.mu.Lock()
	defer as.mu.Unlock()

	total := as.successCount + as.failureCount
	successRate := float64(0)
	if total > 0 {
		successRate = float64(as.successCount) / float64(total) * percentFactor
	}

	reasonsCopy := make(map[string]uint64, len(as.failureReasons))
	for k, v := range as.failureReasons {
		reasonsCopy[k] = v
	}

	summary := EpochSummary{
		Epoch:          currentEpoch,
		Successes:      as.successCount,
		Failures:       as.failureCount,
		SuccessRate:    successRate,
		TotalProcessed: total,
		FailureReasons: reasonsCopy,
		ResetOccurred:  false,
	}

	if currentEpoch > as.lastEpoch {
		as.reset()
		as.lastEpoch = currentEpoch
		summary.ResetOccurred = true
	}

	return summary
}

// reset resets the internal counters and failure reasons map.
// Caller must hold the write lock.
func (as *attestationStats) reset() {
	as.successCount = 0
	as.failureCount = 0
	as.failureReasons = make(map[string]uint64)
}
