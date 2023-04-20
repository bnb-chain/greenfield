package sequence

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
)

var (
	ErrSequenceUniqueConstraint = errors.Register("sequence_u256", 1, "sequence already initialized")
)

// U256 is a persistent unique key generator based on a counter.
type U256 struct {
	storeKey []byte
}

func NewSequence256(prefix []byte) U256 {
	return U256{
		storeKey: prefix,
	}
}

// NextVal increments and persists the counter by one and returns the value.
func (s U256) NextVal(store storetypes.KVStore) math.Uint {
	v := store.Get(s.storeKey)
	seq := DecodeSequence(v)
	seq = seq.Incr()
	store.Set(s.storeKey, EncodeSequence(seq))
	return seq
}

// CurVal returns the last value used. 0 if none.
func (s U256) CurVal(store storetypes.KVStore) math.Uint {
	v := store.Get(s.storeKey)
	return DecodeSequence(v)
}

// PeekNextVal returns the CurVal + increment step. Not persistent.
func (s U256) PeekNextVal(store storetypes.KVStore) math.Uint {
	v := store.Get(s.storeKey)
	seq := DecodeSequence(v)
	seq = seq.Incr()
	return seq
}

// InitVal this function sets the starting value for a sequence and can only be called once
// on an empty database. If the key already exists, an error will be returned. The provided
// start value will be stored as the current value. It is advised to use this function only
// when the sequence start value is not '1', as calling it unnecessarily will consume
// unnecessary gas. An example scenario would be importing from genesis.
func (s U256) InitVal(store storetypes.KVStore, seq math.Uint) error {
	if store.Has(s.storeKey) {
		return ErrSequenceUniqueConstraint
	}

	store.Set(s.storeKey, EncodeSequence(seq))
	return nil
}

func EncodeSequence(u math.Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) math.Uint {
	u := math.NewUint(0)
	return u.SetBytes(bz)
}
