package helpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v5/config/params"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/prysmaticlabs/prysm/v5/math"
	eth "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/runtime/version"
	"github.com/prysmaticlabs/prysm/v5/time/slots"
)

var (
	errNilIl             = errors.New("nil inclusion list")
	errNilCommitteeRoot  = errors.New("nil inclusion list committee root")
	errNilSignature      = errors.New("nil signature")
	errIncorrectState    = errors.New("incorrect state version")
	errCommitteeOverflow = errors.New("committee overflow")
)

func ValidateNilSignedInclusionList(il *eth.SignedInclusionList) error {
	if il == nil {
		return errNilIl
	}
	if il.Signature == nil {
		return errNilSignature
	}
	return ValidateNilInclusionList(il.Message)
}

func ValidateNilInclusionList(il *eth.InclusionList) error {
	if il == nil {
		return errNilIl
	}
	if il.InclusionListCommitteeRoot == nil {
		return errNilCommitteeRoot
	}
	return nil
}

func GetInclusionListCommittee(ctx context.Context, state state.ReadOnlyBeaconState, slot primitives.Slot) (indices []primitives.ValidatorIndex, err error) {
	if state.Version() < version.Electra {
		return nil, errIncorrectState
	}
	if slots.ToEpoch(state.Slot()) < params.BeaconConfig().FocilForkEpoch {
		return nil, errIncorrectState
	}

	committees, err := BeaconCommittees(ctx, state, slot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get beacon committees")
	}
	committeesPerSlot, membersPerCommittee := PtcAllocation(len(committees))
	for i, committee := range committees {
		if uint64(i) >= committeesPerSlot {
			return
		}
		if uint64(len(committee)) < membersPerCommittee {
			return nil, errCommitteeOverflow
		}
		indices = append(indices, committee[:membersPerCommittee]...)
	}
	return
}

func InInclusionListCommittee(ctx context.Context, state state.ReadOnlyBeaconState, slot primitives.Slot, idx primitives.ValidatorIndex) (bool, error) {
	ptc, err := GetInclusionListCommittee(ctx, state, slot)
	if err != nil {
		return false, err
	}
	for _, i := range ptc {
		if i == idx {
			return true, nil
		}
	}
	return false, nil
}

func PtcAllocation(slotCommittees int) (committeesPerSlot, membersPerCommittee uint64) {
	committeesPerSlot = largestPowerOf2(math.Min(uint64(slotCommittees), params.BeaconConfig().InclusionListCommitteeSize))
	membersPerCommittee = params.BeaconConfig().InclusionListCommitteeSize / committeesPerSlot
	return
}

func ValidatePayloadAttestationMessageSignature(ctx context.Context, st state.ReadOnlyBeaconState, il *eth.SignedInclusionList) error {
	if err := ValidateNilSignedInclusionList(il); err != nil {
		return err
	}
	val, err := st.ValidatorAtIndex(il.Message.ValidatorIndex)
	if err != nil {
		return err
	}
	pub, err := bls.PublicKeyFromBytes(val.PublicKey)
	if err != nil {
		return err
	}
	sig, err := bls.SignatureFromBytes(il.Signature)
	if err != nil {
		return err
	}
	currentEpoch := slots.ToEpoch(st.Slot())
	domain, err := signing.Domain(st.Fork(), currentEpoch, params.BeaconConfig().DomainIlcommittee, st.GenesisValidatorsRoot())
	if err != nil {
		return err
	}
	root, err := signing.ComputeSigningRoot(il.Message, domain)
	if err != nil {
		return err
	}
	if !sig.Verify(pub, root[:]) {
		return signing.ErrSigFailedToVerify
	}
	return nil
}

func largestPowerOf2(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n &^ (n >> 1)
}
