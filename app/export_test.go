package app_test

import (
	"testing"

	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/testutil"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestExportAppStateAndValidators(t *testing.T) {
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app, _, _ := testutil.NewTestApp(logger, db, nil, true, test.TEST_CHAIN_ID)

	_, err := app.ExportAppStateAndValidators(false, nil, []string{banktypes.ModuleName})
	if err != nil {
		t.Fatalf("error exporting state: %s", err)
	}
}
