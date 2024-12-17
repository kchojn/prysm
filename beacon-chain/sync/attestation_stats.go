package sync

import (
	"fmt"
	"sync"
	"time"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/sirupsen/logrus"
)

// attestationStats handles attestation verification statistics collection.
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

// recordSuccess increments the successful attestation counter.
func (as *attestationStats) recordSuccess() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.successCount++

	attestationVerificationSuccess.Inc()
}

// recordFailure increments the failed attestation counter and records the failure reason.
func (as *attestationStats) recordFailure(reason string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.failureCount++
	as.failureReasons[reason]++

	attestationVerificationFailure.Inc()
	attestationVerificationFailureReasons.WithLabelValues(reason).Inc()
}

func (as *attestationStats) recordLatency(duration time.Duration) {
	attestationVerificationLatency.Observe(duration.Seconds())
}

// getStats returns the current attestation verification statisticas.
// The returned map is a copy of the internal failure reasons map, so it is safe to read concurrently.
func (as *attestationStats) getStats() (uint64, uint64, map[string]uint64) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	failureCopy := make(map[string]uint64, len(as.failureReasons))

	for k, v := range as.failureReasons {
		failureCopy[k] = v
	}

	return as.successCount, as.failureCount, failureCopy
}

// outputEpochSummary outputs the statistics for the current epoch.
func (as *attestationStats) outputEpochSummary(currentEpoch primitives.Epoch) {
	as.mu.Lock()
	defer as.mu.Unlock()

	successRate := float64(0)
	total := as.successCount + as.failureCount

	if total > 0 {
		successRate = float64(as.successCount) / float64(total) * 100 //nolint:mnd // 100 is the percentage factor.
	}

	log.WithFields(logrus.Fields{
		"epoch":           currentEpoch,
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
