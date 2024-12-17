package blockchain

import (
	"sync"
	"testing"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
)

func TestMetricsCollector_RecordSuccess(t *testing.T) {
	t.Parallel()

	collector := NewMetricsCollector()

	collector.RecordSuccess()
	metrics := collector.GetCurrentMetrics()

	require.Equal(t, uint64(1), metrics.Successes)

	for i := 0; i < 5; i++ {
		collector.RecordSuccess()
	}

	metrics = collector.GetCurrentMetrics()

	require.Equal(t, uint64(6), metrics.Successes)
}

func TestMetricsCollector_RecordFailure(t *testing.T) {
	t.Parallel()

	collector := NewMetricsCollector()

	collector.RecordFailure("decode_error")
	collector.RecordFailure("invalid_signature")
	collector.RecordFailure("invalid_signature")
	collector.RecordFailure("invalid_committee_index")

	metrics := collector.GetCurrentMetrics()

	require.Equal(t, uint64(4), metrics.Failures)
	require.Equal(t, uint64(1), metrics.FailureReasons["decode_error"])
	require.Equal(t, uint64(2), metrics.FailureReasons["invalid_signature"])
	require.Equal(t, uint64(1), metrics.FailureReasons["invalid_committee_index"])
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	collector := NewMetricsCollector()

	const (
		workers    = 100
		iterations = 100
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	failureReasons := []string{
		"decode_error",
		"invalid_signature",
		"invalid_committee_index",
		"nil_attestation",
		"wrong_message_type",
	}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if j%2 == 0 {
					collector.RecordSuccess()
				} else {
					reason := failureReasons[j%len(failureReasons)]
					collector.RecordFailure(reason)
				}
			}
		}()
	}
	wg.Wait()

	metrics := collector.GetCurrentMetrics()
	expectedTotal := uint64(workers * iterations)
	actualTotal := metrics.Successes + metrics.Failures

	require.Equal(t, expectedTotal, actualTotal)

	for _, reason := range failureReasons {
		count := metrics.FailureReasons[reason]
		require.NotEqual(t, uint64(0), count, "Failure reason "+reason+" did not occur")
	}

	var totalReasonCounts uint64

	for _, count := range metrics.FailureReasons {
		totalReasonCounts += count
	}

	require.Equal(t, metrics.Failures, totalReasonCounts)
	require.NotEqual(t, uint64(0), metrics.Successes)
	require.NotEqual(t, uint64(0), metrics.Failures)
	require.NotEqual(t, expectedTotal, metrics.Failures)
	require.NotEqual(t, expectedTotal, metrics.Successes)
}

func TestMetricsCollector_AdvanceEpoch(t *testing.T) {
	t.Parallel()

	collector := NewMetricsCollector()

	collector.RecordSuccess()
	collector.RecordSuccess()
	collector.RecordFailure("decode_error")
	collector.RecordFailure("invalid_signature")

	oldMetrics := collector.AdvanceEpoch(primitives.Epoch(1))

	require.Equal(t, primitives.Epoch(0), oldMetrics.Epoch)
	require.Equal(t, uint64(2), oldMetrics.Successes)
	require.Equal(t, uint64(2), oldMetrics.Failures)
	require.Equal(t, (2.0/4.0)*100, oldMetrics.SuccessRate)
	require.Equal(t, uint64(4), oldMetrics.TotalProcessed)
	require.NotNil(t, oldMetrics.FailureReasons["decode_error"])
	require.NotNil(t, oldMetrics.FailureReasons["invalid_signature"])

	metrics := collector.GetCurrentMetrics()
	require.Equal(t, uint64(0), metrics.Successes)
	require.Equal(t, uint64(0), metrics.Failures)

	collector.RecordSuccess()
	collector.RecordFailure("decode_error")

	oldMetrics = collector.AdvanceEpoch(primitives.Epoch(2))

	require.Equal(t, primitives.Epoch(1), oldMetrics.Epoch)
	require.Equal(t, uint64(1), oldMetrics.Successes)
	require.Equal(t, uint64(1), oldMetrics.Failures)
	require.Equal(t, uint64(2), oldMetrics.TotalProcessed)
	require.Equal(t, uint64(1), oldMetrics.FailureReasons["decode_error"])

	metrics = collector.GetCurrentMetrics()

	require.Equal(t, uint64(0), metrics.Successes)
	require.Equal(t, uint64(0), metrics.Failures)
}
