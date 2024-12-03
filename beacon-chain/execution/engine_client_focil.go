package execution

import (
	"context"
	"time"

	"github.com/prysmaticlabs/prysm/v5/config/params"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/monitoring/tracing/trace"
)

const (
	NewPayloadMethodV5               = "engine_newPayloadV5" // Do we really need this?
	GetInclusionListV1               = "engine_getInclusionListV1"
	UpdatePayloadWithInclusionListV1 = "engine_updatePayloadWithInclusionListV1"
)

type InclusionListV1 struct {
	Transactions [][]byte `json:"transactions"  gencodec:"required"`
}

func (s *Service) GetInclusionList(ctx context.Context, parentHash [32]byte) ([][]byte, error) {
	ctx, span := trace.StartSpan(ctx, "execution.GetInclusionList")
	defer span.End()
	start := time.Now()
	defer func() {
		getInclusionListLatency.Observe(float64(time.Since(start).Milliseconds()))
	}()

	d := time.Now().Add(time.Duration(params.BeaconConfig().ExecutionEngineTimeoutValue) * time.Second)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	result := &InclusionListV1{}
	err := s.rpcClient.CallContext(ctx, result, GetInclusionListV1, parentHash)
	if err != nil {
		return nil, handleRPCError(err)
	}
	return result.Transactions, nil
}

func (s *Service) UpdatePayloadWithInclusionList(ctx context.Context, payloadID primitives.PayloadID, txs [][]byte) (*primitives.PayloadID, error) {
	ctx, span := trace.StartSpan(ctx, "execution.UpdatePayloadWithInclusionList")
	defer span.End()
	start := time.Now()
	defer func() {
		updatePayloadWithInclusionListLatency.Observe(float64(time.Since(start).Milliseconds()))
	}()

	d := time.Now().Add(time.Duration(params.BeaconConfig().ExecutionEngineTimeoutValue) * time.Second)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	result := &primitives.PayloadID{}
	err := s.rpcClient.CallContext(ctx, result, UpdatePayloadWithInclusionListV1, &InclusionListV1{
		txs,
	})
	if err != nil {
		return nil, handleRPCError(err)
	}
	return result, nil
}
