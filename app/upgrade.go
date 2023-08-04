package app

import (
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app *App) RegisterUpgradeHandlers(chainID string, serverCfg *serverconfig.Config) error {
	// Register the plans from server config
	err := app.UpgradeKeeper.RegisterUpgradePlan(chainID, serverCfg.Upgrade)
	if err != nil {
		return err
	}

	// Register the upgrade handlers here
	// app.registerPublicDelegationUpgradeHandler()
	app.registerBEP1001UpgradeHandler()

	return nil
}

// registerPublicDelegationUpgradeHandler registers the upgrade handlers for the public delegation upgrade.
// func (app *App) registerPublicDelegationUpgradeHandler() {
// 	// Register the upgrade handler
// 	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.EnablePublicDelegationUpgrade,
// 		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
// 			app.Logger().Info("upgrade to ", plan.Name)
// 			return fromVM, nil
// 		})

// 	// Register the upgrade initializer
// 	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.EnablePublicDelegationUpgrade,
// 		func() error {
// 			app.Logger().Info("Init enable public delegation upgrade")
// 			return nil
// 		},
// 	)
// }

// registerBEP1001UpgradeHandler registers the upgrade handlers for BEP1001.
func (app *App) registerBEP1001UpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Nagqu,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("processing upgrade handler", "name", plan.Name, "info", plan.Info)
			app.Logger().Info("register /greenfield.storage.MsgRenewGroupMember gas params", "name", plan.Name, "info", plan.Info)
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas("/greenfield.storage.MsgRenewGroupMember", 1.2e3))
			return fromVM, nil
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Nagqu,
		func() error {
			app.Logger().Info("processing upgrade initializer", "name", upgradetypes.Nagqu)
			// enable the expiration of the group member from cross-chain operation
			app.Logger().Info("register UpdateGroupMemberV2SynPackageType")
			storagetypes.RegisterUpdateGroupMemberV2SynPackageType()
			return nil
		},
	)
}
