package sequence_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	sdkstore "github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/internal/sequence"
)

type MockContext struct {
	db    *dbm.MemDB
	store storetypes.CommitMultiStore
}

func (m MockContext) KVStore(key storetypes.StoreKey) storetypes.KVStore {
	if s := m.store.GetCommitKVStore(key); s != nil {
		return s
	}
	m.store.MountStoreWithDB(key, storetypes.StoreTypeIAVL, m.db)
	if err := m.store.LoadLatestVersion(); err != nil {
		panic(err)
	}
	return m.store.GetCommitKVStore(key)
}

func NewMockContext() *MockContext {
	db := dbm.NewMemDB()
	return &MockContext{
		db:    dbm.NewMemDB(),
		store: sdkstore.NewCommitMultiStore(db),
	}
}

func TestSequenceUniqueConstraint(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	seq := sequence.NewSequence[uint32]([]byte{0x1})
	err := seq.InitVal(store, 0)
	require.NoError(t, err)
	err = seq.InitVal(store, 1)
	require.True(t, sequence.ErrSequenceUniqueConstraint.Is(err))
}

func TestSequenceIncrementsUint256(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))
	seq := sequence.NewSequence[math.Uint]([]byte{0x1})
	max := math.NewUint(10)
	i := math.ZeroUint()
	for i.LT(max) {
		id := seq.NextVal(store)
		curId := seq.CurVal(store)
		i = i.Incr()
		assert.True(t, i.Equal(id))
		assert.True(t, i.Equal(curId))
		fmt.Print("i= ", i.Uint64(), "id=", id.Uint64(), "curID", curId.Uint64())
	}
}

func TestSequenceIncrementsU32(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))
	seq := sequence.NewSequence[uint32]([]byte{0x1})
	max := uint32(10)
	i := uint32(0)
	for i < max {
		id := seq.NextVal(store)
		curId := seq.CurVal(store)
		i++
		assert.Equal(t, i, id)
		assert.Equal(t, i, curId)
		fmt.Print("i= ", i, "id=", id, "curID", curId)
	}
}

func TestSequenceU32(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	seq := sequence.NewSequence[uint32]([]byte{0x1})
	err := seq.InitVal(store, 0)
	require.NoError(t, err)
	n := seq.NextVal(store)
	require.Equal(t, n, uint32(1))
}

func TestSequenceU256(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	seq := sequence.NewSequence[math.Uint]([]byte{0x1})
	err := seq.InitVal(store, math.ZeroUint())
	require.NoError(t, err)
	n := seq.NextVal(store)
	require.Equal(t, n, math.OneUint())
}
