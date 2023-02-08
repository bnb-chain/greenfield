package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

type MockContext struct {
	db    *dbm.MemDB
	store storetypes.CommitMultiStore
}

func (m MockContext) KVStore(key storetypes.StoreKey) sdk.KVStore {
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
		store: store.NewCommitMultiStore(db),
	}
}

func TestSequenceUniqueConstraint(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	seq := NewSequence([]byte{0x1})
	err := seq.InitVal(store, math.NewUint(0))
	require.NoError(t, err)
	err = seq.InitVal(store, math.NewUint(1))
	require.True(t, types.ErrSequenceUniqueConstraint.Is(err))
}

func TestSequenceIncrements(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))
	seq := NewSequence([]byte{0x1})
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
