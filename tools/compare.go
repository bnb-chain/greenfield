package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"

	db "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/state"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	cosmosIavl "github.com/cosmos/iavl"

	bridgemoduletypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengemoduletypes "github.com/bnb-chain/greenfield/x/challenge/types"
	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissionmoduletypes "github.com/bnb-chain/greenfield/x/permission/types"
	spmoduletypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagemoduletypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const reconStoreKey = "reconciliation"

func init() {
}

func openDB(root, dbName string) *db.GoLevelDB {
	db, err := db.NewGoLevelDB(dbName, path.Join(root, "data"))
	if err != nil {
		fmt.Printf("new levelDb err in path %s\n", path.Join(root, "data"))
		panic(err)
	}
	return db
}

func openAppDB(root string) *db.GoLevelDB {
	return openDB(root, "application")
}

func stateDiff(root1, root2 string) {
	s1 := getState(root1)
	s2 := getState(root2)
	fmt.Printf("State| Height:%d: AppHash:%X\n", s1.LastBlockHeight, s1.AppHash)
	fmt.Printf("State| Height:%d: AppHash:%X\n", s2.LastBlockHeight, s2.AppHash)
}

func getState(root string) state.State {
	stateDb := openDB(root, "state")
	defer stateDb.Close()

	stateStore := state.NewStore(stateDb, state.StoreOptions{})
	s, err := stateStore.Load()
	if err != nil {
		panic(err)
	}
	return s
}

func prepareCms(root string, appDB *db.GoLevelDB, storeKeys []storetypes.StoreKey, height int64) sdk.CommitMultiStore {
	cms := store.NewCommitMultiStore(appDB)
	for _, key := range storeKeys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	err := cms.LoadVersion(height)
	if err != nil {
		fmt.Printf("height %d does not exist in %s\n", height, root)
		panic(err)
	}
	return cms
}

func getNonTransientStoreKeys() (storeKeys []storetypes.StoreKey) {
	keys := []string{
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		feegrant.StoreKey, evidencetypes.StoreKey, consensusparamtypes.StoreKey,
		group.StoreKey, upgradetypes.StoreKey,
		crosschaintypes.StoreKey,
		oracletypes.StoreKey,
		bridgemoduletypes.StoreKey,
		gashubtypes.StoreKey,
		spmoduletypes.StoreKey,
		virtualgroupmoduletypes.StoreKey,
		paymentmoduletypes.StoreKey,
		permissionmoduletypes.StoreKey,
		storagemoduletypes.StoreKey,
		challengemoduletypes.StoreKey,
		reconStoreKey,
	}
	stokeKeys := make([]storetypes.StoreKey, 0)
	for _, k := range keys {
		stokeKeys = append(stokeKeys, storetypes.NewKVStoreKey(k))
	}

	return stokeKeys
}

func compare(height int64, root1, root2 string) {
	stateDiff(root1, root2)
	db1 := openAppDB(root1)
	defer db1.Close()
	db2 := openAppDB(root2)
	defer db2.Close()
	keys := getNonTransientStoreKeys()
	cms1 := prepareCms(root1, db1, keys, height)
	cms2 := prepareCms(root2, db2, keys, height)

	if bytes.Equal(cms1.LastCommitID().Hash, cms2.LastCommitID().Hash) {
		fmt.Printf("commitId is the same, %X", cms1.LastCommitID().Hash)
	}

	fmt.Printf("Commit Id| Height:%d, AppHash:%X\n", cms1.LastCommitID().Version, cms1.LastCommitID().Hash)
	fmt.Printf("Commit Id| Height:%d, AppHash:%X\n", cms2.LastCommitID().Version, cms2.LastCommitID().Hash)
	for _, key := range keys {
		tree1, err := cms1.GetCommitStore(key).(*iavl.Store).CloneMutableTree().GetImmutable(height)
		if err != nil {
			panic(err)
		}
		tree2, err := cms2.GetCommitStore(key).(*iavl.Store).CloneMutableTree().GetImmutable(height)
		if err != nil {
			panic(err)
		}
		hash1, err := tree1.Hash()
		if err != nil {
			panic(err)
		}
		hash2, err := tree2.Hash()
		if err != nil {
			panic(err)
		}

		if bytes.Equal(hash1, hash2) {
			fmt.Printf("identical, %-6s, %X\n", key.Name(), hash1)
			continue
		} else {
			fmt.Printf("diff found in: %s, %X %X \n", key.Name(), hash1, hash2)
		}

		diff(tree1, tree2)
	}
}

func diff(tree1 *cosmosIavl.ImmutableTree, tree2 *cosmosIavl.ImmutableTree) {
	fmt.Println("iterate tree1")
	iterate(tree1, tree2, false)
	fmt.Println("iterate tree2")
	iterate(tree2, tree1, false)
}

func iterate(source *cosmosIavl.ImmutableTree, target *cosmosIavl.ImmutableTree, printAll bool) {
	it1, err := source.Iterator(nil, nil, true)
	if err != nil {
		panic(err)
	}
	defer it1.Close()

	for ; it1.Valid(); it1.Next() {
		key := it1.Key()
		value1 := it1.Value()
		value2, err := target.Get(key)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(value1, value2) {
			fmt.Printf("diff key in two trees, %s, %X, %X, %X\n", key, key, value1, value2)
		}
		if printAll {
			fmt.Printf("key in two trees, %s, %X, %X, %X\n", key, key, value1, value2)
		}
	}
}

func main() {
	args := os.Args
	if len(args) != 4 {
		fmt.Printf("usage: ./compare height home_path1 home_path2")
		return
	}

	heightStr := os.Args[1]
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		fmt.Printf("parsing height[%s] error: %s", heightStr, err.Error())
		return
	}

	home1 := os.Args[2]
	home2 := os.Args[3]

	compare(height, home1, home2)
}
