package keeper

import (
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sequenceIDPrefix a fix key to read/ write data on the storage layer
var sequenceIDPrefix = []byte{0x1}

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
	v := pStore.Get(sequenceIDPrefix)
	seq := math.ZeroUint()
	var err error
	err = seq.Unmarshal(v)
	if err != nil {
		panic(err)
	}
	seq = seq.Incr()

	var bz []byte
	if bz, err = seq.Marshal(); err != nil {
		panic(err)
	}
	pStore.Set(sequenceIDPrefix, bz)
	return seq
}

// CurVal returns the last value used. 0 if none.
func (s Sequence) CurVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceIDPrefix)
	var seq math.Uint
	_ = seq.Unmarshal(v)
	return seq
}

// PeekNextVal returns the CurVal + increment step. Not persistent.
func (s Sequence) PeekNextVal(store sdk.KVStore) math.Uint {
	pStore := prefix.NewStore(store, s.prefix)
	v := pStore.Get(sequenceIDPrefix)
	var seq math.Uint
	err := seq.Unmarshal(v)
	if err != nil {
		panic(err)
	}
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
	if pStore.Has(sequenceIDPrefix) {
		return types.ErrSequenceUniqueConstraint
	}
	var bz []byte
	var err error
	if bz, err = seq.Marshal(); err != nil {
		return err
	}
	pStore.Set(sequenceIDPrefix, bz)
	return nil
}
