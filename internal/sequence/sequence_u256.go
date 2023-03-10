package sequence

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sequenceKey a fix key to read/ write data on the storage layer
var (
	sequenceKey                 = []byte{0x1}
	ErrSequenceUniqueConstraint = errors.Register("sequence_u256", 1, "sequence already initialized")
)

// SequenceU256 is a persistent unique key generator based on a counter.
type U256 struct {
	prefix []byte
}

func NewSequence256(prefix []byte) U256 {
	return U256{
		prefix: prefix,
	}
}

// NextVal increments and persists the counter by one and returns the value.
func (s U256) NextVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	seq := DecodeSequence(v)
	seq = seq.Incr()
	pStore.Set(sequenceKey, EncodeSequence(seq))
	return seq
}

// CurVal returns the last value used. 0 if none.
func (s U256) CurVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	return DecodeSequence(v)
}

// PeekNextVal returns the CurVal + increment step. Not persistent.
func (s U256) PeekNextVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceKey)
	seq := DecodeSequence(v)
	seq = seq.Incr()
	return seq
}

// InitVal sets the start value for the sequence. It must be called only once on an empty DB.
// Otherwise an error is returned when the key exists. The given start value is stored as current
// value.
//
// It is recommended to call this method only for a sequence start value other than `1` as the
// method consumes unnecessary gas otherwise. A scenario would be an import from genesis.
func (s U256) InitVal(store sdk.KVStore, seq math.Uint) error {
	pStore := prefix.NewStore(store, s.prefix)
	if pStore.Has(sequenceKey) {
		return ErrSequenceUniqueConstraint
	}

	pStore.Set(sequenceKey, EncodeSequence(seq))
	return nil
}

func EncodeSequence(u math.Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) math.Uint {
	u := math.NewUint(0)
	return u.SetBytes(bz)
}
