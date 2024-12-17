package attestations

import (
	"sync"
	"testing"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/stretchr/testify/require"
)

func TestStatsCollector_RecordSuccess(t *testing.T) {
	t.Parallel()

	collector := New()

	collector.RecordSuccess()

	require.Equal(t, uint64(1), collector.successfulCount)

	for i := 0; i < 5; i++ {
		collector.RecordSuccess()
	}

	require.Equal(t, uint64(6), collector.successfulCount)
}

func TestStatsCollector_RecordFailure(t *testing.T) {
	t.Parallel()

	collector := New()

	reason := "invalid_signature"

	collector.RecordFailure(reason)

	require.Equal(t, uint64(1), collector.failedCount)
	require.Equal(t, uint64(1), collector.failureReasons[reason])

	for i := 0; i < 3; i++ {
		collector.RecordFailure(reason)
	}

	require.Equal(t, uint64(4), collector.failedCount)
	require.Equal(t, uint64(4), collector.failureReasons[reason])
}

func TestStatsCollector_Concurrent(t *testing.T) {
	t.Parallel()

	collector := New()

	const (
		workers    = 100
		iterations = 100
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if j%2 == 0 {
					collector.RecordSuccess()
				} else {
					collector.RecordFailure("test_reason")
				}
			}
		}()
	}
	wg.Wait()

	expectedTotal := uint64(workers * iterations)
	actualTotal := collector.successfulCount + collector.failedCount

	require.Equal(t, expectedTotal, actualTotal)
}

func TestStatsCollector_OutputEpochSummary(t *testing.T) {
	t.Parallel()

	collector := New()

	collector.RecordSuccess()
	collector.RecordFailure("reason1")
	collector.RecordFailure("reason2")
	collector.OutputEpochSummary(primitives.Epoch(1))

	require.Equal(t, uint64(0), collector.successfulCount)
	require.Equal(t, uint64(0), collector.failedCount)
	require.Equal(t, 0, len(collector.failureReasons))
}
