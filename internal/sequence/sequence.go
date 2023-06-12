package sequence

import (
	"encoding/binary"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

var (
	ErrSequenceUniqueConstraint = errors.Register("sequence_u256", 1, "sequence already initialized")
)

type Number interface {
	uint32 | math.Uint
}

type Sequence[T Number] struct {
	storeKey []byte
}

func NewSequence[T Number](prefix []byte) Sequence[T] {
	return Sequence[T]{
		storeKey: prefix,
	}
}

func (s Sequence[T]) NextVal(store storetypes.KVStore) T {
	v := store.Get(s.storeKey)
	seq := s.DecodeSequence(v)
	seq = s.IncreaseSequence(seq)
	store.Set(s.storeKey, s.EncodeSequence(seq))
	return any(seq).(T)
}

// CurVal returns the last value used. 0 if none.
func (s Sequence[T]) CurVal(store storetypes.KVStore) T {
	v := store.Get(s.storeKey)
	ret := s.DecodeSequence(v)
	return any(ret).(T)
}

// PeekNextVal returns the CurVal + increment step. Not persistent.
func (s Sequence[T]) PeekNextVal(store storetypes.KVStore) T {
	v := store.Get(s.storeKey)
	seq := s.DecodeSequence(v)
	seq = s.IncreaseSequence(seq)
	return any(seq).(T)
}

// InitVal this function sets the starting value for a sequence and can only be called once
// on an empty database. If the key already exists, an error will be returned. The provided
// start value will be stored as the current value. It is advised to use this function only
// when the sequence start value is not '1', as calling it unnecessarily will consume
// unnecessary gas. An example scenario would be importing from genesis.
func (s Sequence[T]) InitVal(store storetypes.KVStore, seq T) error {
	if store.Has(s.storeKey) {
		return ErrSequenceUniqueConstraint
	}
	store.Set(s.storeKey, s.EncodeSequence(seq))
	return nil
}

func (s Sequence[T]) ToUint32(seq T) uint32 {
	var t T
	switch ret := any(t).(type) {
	case uint32:
		return ret
	default:
		return 0
	}
}

func (s Sequence[T]) ToUint256(seq T) math.Uint {
	var t T
	switch ret := any(t).(type) {
	case math.Uint:
		return ret
	default:
		return math.ZeroUint()
	}
}

const EncodedSeqLength = 4

func (s Sequence[T]) EncodeSequence(t T) []byte {
	switch ret := any(t).(type) {
	case math.Uint:
		return ret.Bytes()
	case uint32:
		bz := make([]byte, EncodedSeqLength)
		binary.BigEndian.PutUint32(bz, ret)
		return bz
	default:
		return nil
	}
}

func (s Sequence[T]) DecodeSequence(bz []byte) T {
	var t T
	switch any(t).(type) {
	case math.Uint:
		u := math.ZeroUint()
		if bz != nil {
			u = u.SetBytes(bz)
		}
		return any(u).(T)
	case uint32:
		u := uint32(0)
		if bz != nil {
			u = binary.BigEndian.Uint32(bz)
		}
		return any(u).(T)
	default:
		return t
	}
}

func (s Sequence[T]) IncreaseSequence(t T) T {
	switch ret := any(t).(type) {
	case math.Uint:
		ret = ret.Incr()
		return any(ret).(T)
	case uint32:
		ret++
		return any(ret).(T)
	default:
		return t
	}
}
