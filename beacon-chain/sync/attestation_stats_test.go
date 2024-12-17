package sync

import (
	"sync"
	"testing"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
)

func TestAttestationStats_RecordSuccess(t *testing.T) {
	t.Parallel()

	stats := newAttestationStats()

	stats.recordSuccess()
	successes, _, _ := stats.getStats()
	require.Equal(t, uint64(1), successes)

	for i := 0; i < 5; i++ {
		stats.recordSuccess()
	}

	successes, _, _ = stats.getStats()
	require.Equal(t, uint64(6), successes)
}

func TestAttestationStats_RecordFailure(t *testing.T) {
	t.Parallel()

	stats := newAttestationStats()

	stats.recordFailure("decode_error")
	stats.recordFailure("invalid_signature")
	stats.recordFailure("invalid_signature")
	stats.recordFailure("invalid_committee_index")

	_, failures, reasons := stats.getStats()
	require.Equal(t, uint64(4), failures)
	require.Equal(t, uint64(1), reasons["decode_error"])
	require.Equal(t, uint64(2), reasons["invalid_signature"])
	require.Equal(t, uint64(1), reasons["invalid_committee_index"])
}

func TestAttestationStats_Concurrent(t *testing.T) {
	t.Parallel()

	stats := newAttestationStats()

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
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if j%2 == 0 {
					stats.recordSuccess()
				} else {
					reason := failureReasons[j%len(failureReasons)]
					stats.recordFailure(reason)
				}
			}
		}(i)
	}
	wg.Wait()

	successes, failures, reasons := stats.getStats()
	expectedTotal := uint64(workers * iterations)
	actualTotal := successes + failures

	require.Equal(t, expectedTotal, actualTotal, "Total operations count incorrect")

	for _, reason := range failureReasons {
		count := reasons[reason]
		require.NotEqual(t, uint64(0), count, "Failure reason "+reason+" did not occur")
	}

	var totalReasonCounts uint64
	for _, count := range reasons {
		totalReasonCounts += count
	}
	require.Equal(t, failures, totalReasonCounts, "Sum of failure reasons does not match total failures")

	require.NotEqual(t, uint64(0), successes, "No successes recorded")
	require.NotEqual(t, uint64(0), failures, "No failures recorded")
	require.NotEqual(t, expectedTotal, failures, "All operations were failures")
	require.NotEqual(t, expectedTotal, successes, "All operations were successes")
}

func TestAttestationStats_OutputEpochSummary(t *testing.T) {
	t.Parallel()

	stats := newAttestationStats()

	stats.recordSuccess()
	stats.recordSuccess()
	stats.recordFailure("decode_error")
	stats.recordFailure("invalid_signature")

	stats.outputEpochSummary(primitives.Epoch(1))

	successes, failures, reasons := stats.getStats()
	require.Equal(t, uint64(0), successes)
	require.Equal(t, uint64(0), failures)
	require.Equal(t, 0, len(reasons))
	require.Equal(t, primitives.Epoch(1), stats.lastEpoch)

	stats.recordSuccess()
	stats.recordFailure("decode_error")
	stats.outputEpochSummary(primitives.Epoch(1))

	successes, failures, reasons = stats.getStats()
	require.Equal(t, uint64(1), successes)
	require.Equal(t, uint64(1), failures)
	require.Equal(t, uint64(1), reasons["decode_error"])

	stats.outputEpochSummary(primitives.Epoch(2))
	successes, failures, reasons = stats.getStats()
	require.Equal(t, uint64(0), successes)
	require.Equal(t, uint64(0), failures)
	require.Equal(t, 0, len(reasons))
	require.Equal(t, primitives.Epoch(2), stats.lastEpoch)
}

func TestAttestationStats_GetStats(t *testing.T) {
	t.Parallel()

	stats := newAttestationStats()

	stats.recordSuccess()
	stats.recordFailure("decode_error")
	stats.recordFailure("decode_error")
	stats.recordFailure("invalid_signature")

	successes, failures, reasons := stats.getStats()
	require.Equal(t, uint64(1), successes)
	require.Equal(t, uint64(3), failures)
	require.Equal(t, uint64(2), reasons["decode_error"])
	require.Equal(t, uint64(1), reasons["invalid_signature"])

	reasons["decode_error"] = 999
	_, _, newReasons := stats.getStats()
	require.Equal(t, uint64(2), newReasons["decode_error"])
}
