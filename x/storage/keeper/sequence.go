package keeper

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

// sequenceKey a fix key to read/ write data on the storage layer
var sequenceKey = []byte{0x1}

// sequence is a persistent unique key generator based on a counter.
type Sequence struct {
	prefix []byte
}

func NewSequence(prefix []byte) Sequence {
	return Sequence{
		prefix: prefix,
	}
}

// NextVal increments and persists the counter by one and returns the value.
func (s Sequence) NextVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	seq := types.MustUnmarshalUint(v)
	seq = seq.Incr()

	pStore.Set(sequenceKey, types.MustMarshalUint(seq))
	return seq
}

// CurVal returns the last value used. 0 if none.
func (s Sequence) CurVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	return types.MustUnmarshalUint(v)
}

// PeekNextVal returns the CurVal + increment step. Not persistent.
func (s Sequence) PeekNextVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	seq := types.MustUnmarshalUint(v)
	seq = seq.Incr()
	return seq
}

// InitVal sets the start value for the sequence. It must be called only once on an empty DB.
// Otherwise an error is returned when the key exists. The given start value is stored as current
// value.
//
// It is recommended to call this method only for a sequence start value other than `1` as the
// method consumes unnecessary gas otherwise. A scenario would be an import from genesis.
func (s Sequence) InitVal(store sdk.KVStore, seq math.Uint) error {
	pStore := prefix.NewStore(store, s.prefix)
	if pStore.Has(sequenceKey) {
		return types.ErrSequenceUniqueConstraint
	}

	pStore.Set(sequenceKey, types.MustMarshalUint(seq))
	return nil
}
