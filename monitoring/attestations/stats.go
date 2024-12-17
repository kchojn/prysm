package attestations

import (
	"sync"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"

	"github.com/sirupsen/logrus"
)

// StatsCollector handles attestation verification statistics collection.
type StatsCollector struct {
	mu              sync.RWMutex
	successfulCount uint64
	failedCount     uint64
	failureReasons  map[string]uint64
	currentEpoch    primitives.Epoch
}

// New creates a new StatsCollector instance.
func New() *StatsCollector {
	return &StatsCollector{
		failureReasons: make(map[string]uint64),
	}
}

// RecordSuccess increments the successful attestation counter.
func (s *StatsCollector) RecordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.successfulCount++
}

// RecordFailure increments the failed attestation counter and records the failure reason.
func (s *StatsCollector) RecordFailure(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failedCount++
	s.failureReasons[reason]++
}

// GetStats returns the current attestation verification statistics.
// The returned map is a copy of the internal failure reasons map, so it is safe to read concurrently.
func (s *StatsCollector) GetStats() (uint64, uint64, map[string]uint64, primitives.Epoch) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	failureCopy := make(map[string]uint64, len(s.failureReasons))
	for k, v := range s.failureReasons {
		failureCopy[k] = v
	}
	return s.successfulCount, s.failedCount, failureCopy, s.currentEpoch
}

// OutputEpochSummary outputs the statistics for the current epoch.
func (s *StatsCollector) OutputEpochSummary(epoch primitives.Epoch) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"epoch":            epoch,
		"successful_count": s.successfulCount,
		"failed_count":     s.failedCount,
		"failure_reasons":  s.failureReasons,
	}).Info("Attestation verification summary for epoch")

	s.successfulCount = 0
	s.failedCount = 0
	s.failureReasons = make(map[string]uint64)
	s.currentEpoch = epoch
}
