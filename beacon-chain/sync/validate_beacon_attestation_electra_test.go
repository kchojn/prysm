package sync

import (
	"context"
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/testing/assert"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
)

func Test_validateCommitteeIndexElectra(t *testing.T) {
	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		ci, res, err := validateCommitteeIndexElectra(ctx, &ethpb.SingleAttestation{Data: &ethpb.AttestationData{}, CommitteeId: 1})
		require.NoError(t, err)
		assert.Equal(t, pubsub.ValidationAccept, res)
		assert.Equal(t, primitives.CommitteeIndex(1), ci)
	})
	t.Run("non-zero committee index in att data", func(t *testing.T) {
		_, res, err := validateCommitteeIndexElectra(ctx, &ethpb.SingleAttestation{Data: &ethpb.AttestationData{CommitteeIndex: 1}})
		assert.NotNil(t, err)
		assert.Equal(t, pubsub.ValidationReject, res)
	})
}
