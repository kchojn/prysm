package sync

import (
	"fmt"
	"sync"
	"time"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/sirupsen/logrus"
)

const (
	percentFactor = 100
)

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
// It keeps track of successful and failed attestations, along with failure reasons,
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

	attestationVerificationSuccess.Inc()
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

	attestationVerificationFailure.Inc()
	attestationVerificationFailureReasons.WithLabelValues(reason).Inc()
}

// recordLatency records the duration of attestation verification.
// Parameters:
//   - duration: The time taken to verify the attestation
func (as *attestationStats) recordLatency(duration time.Duration) {
	attestationVerificationLatency.Observe(duration.Seconds())
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

// outputEpochSummary outputs the statistics for the current epoch and resets counters if needed.
// This method logs the success rate and failure breakdown for the current epoch.
// If the current epoch is greater than the last recorded epoch, all statistics are reset.
// Parameters:
//   - currentEpoch: The current epoch number for which to output statistics
func (as *attestationStats) outputEpochSummary(currentEpoch primitives.Epoch) {
	as.mu.Lock()
	defer as.mu.Unlock()

	successRate := float64(0)
	total := as.successCount + as.failureCount

	if total > 0 {
		successRate = float64(as.successCount) / float64(total) * percentFactor
	}

	log.WithFields(logrus.Fields{
		"epoch":           currentEpoch,
		"context":         "attestation_verification",
		"successes":       as.successCount,
		"failures":        as.failureCount,
		"success_rate":    fmt.Sprintf("%.2f%%", successRate),
		"total_processed": total,
	}).Info("Attestation verification epoch summary")

	if as.failureCount > 0 {
		log.WithField("failure_reasons", as.failureReasons).Info("Attestation verification failure breakdown")
	}

	if currentEpoch > as.lastEpoch {
		as.successCount = 0
		as.failureCount = 0
		as.failureReasons = make(map[string]uint64)
		as.lastEpoch = currentEpoch
	}
}
