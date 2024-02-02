package app

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bridgemoduletypes "github.com/bnb-chain/greenfield/x/bridge/types"
	paymentmodule "github.com/bnb-chain/greenfield/x/payment"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagemoduletypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmodule "github.com/bnb-chain/greenfield/x/virtualgroup"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (app *App) RegisterUpgradeHandlers(chainID string, serverCfg *serverconfig.Config) error {
	// Register the plans from server config
	err := app.UpgradeKeeper.RegisterUpgradePlan(chainID, serverCfg.Upgrade)
	if err != nil {
		return err
	}

	// Register the upgrade handlers here
	app.registerNagquUpgradeHandler()
	app.registerPampasUpgradeHandler()
	app.registerManchurianUpgradeHandler()
	app.registerHulunbeierUpgradeHandler()
	app.registerHulunbeierPatchUpgradeHandler()
	app.registerUralUpgradeHandler()
	app.registerPawneeUpgradeHandler()
	// app.register...()
	// ...
	return nil
}

// registerPublicDelegationUpgradeHandler registers the upgrade handlers for the public delegation upgrade.
// it will be enabled at the future version.
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

func (app *App) registerNagquUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Nagqu,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Nagqu,
		func() error {
			app.Logger().Info("Init Nagqu upgrade")
			mm, ok := app.mm.Modules[paymenttypes.ModuleName].(*paymentmodule.AppModule)
			if !ok {
				panic("*paymentmodule.AppModule not found")
			}
			mm.SetConsensusVersion(2)
			return nil
		})
}

func (app *App) registerPampasUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Pampas,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)

			// open resource channels for opbnb
			app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestOpChainId), bridgemoduletypes.SyncParamsChannelID, sdk.ChannelAllow)
			app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestOpChainId), storagemoduletypes.BucketChannelId, sdk.ChannelAllow)
			app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestOpChainId), storagemoduletypes.ObjectChannelId, sdk.ChannelAllow)
			app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestOpChainId), storagemoduletypes.GroupChannelId, sdk.ChannelAllow)

			// disable sp exit
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.virtualgroup.MsgSwapOut")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.virtualgroup.MsgCompleteSwapOut")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.virtualgroup.MsgCancelSwapOut")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.virtualgroup.MsgStorageProviderExit")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.virtualgroup.MsgCompleteStorageProviderExit")

			// disable bucket migration.
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.storage.MsgMigrateBucket")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.storage.MsgCancelMigrateBucket")
			app.GashubKeeper.DeleteMsgGasParams(ctx, "/greenfield.storage.MsgCompleteMigrateBucket")

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Pampas,
		func() error {
			app.Logger().Info("Init Pampas upgrade")

			// enable chain id for opbnb
			app.CrossChainKeeper.SetDestOpChainID(sdk.ChainID(app.appConfig.CrossChain.DestOpChainId))
			return nil
		})
}

func (app *App) registerManchurianUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Manchurian,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)

			typeUrl := sdk.MsgTypeURL(&storagemoduletypes.MsgSetTag{})
			msgSetTagGasParams := gashubtypes.NewMsgGasParamsWithFixedGas(typeUrl, 1.2e3)
			app.GashubKeeper.SetMsgGasParams(ctx, *msgSetTagGasParams)

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Manchurian,
		func() error {
			app.Logger().Info("Init Manchurian upgrade")
			return nil
		})
}

func (app *App) registerHulunbeierUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Hulunbeier,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)

			// enable SP exit
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgReserveSwapIn{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgCancelSwapIn{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgCompleteSwapIn{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgStorageProviderForcedExit{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgStorageProviderExit{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&virtualgrouptypes.MsgCompleteStorageProviderExit{}), 1.2e3))

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Hulunbeier,
		func() error {
			app.Logger().Info("Init Hulunbeier upgrade")
			mm, ok := app.mm.Modules[virtualgrouptypes.ModuleName].(*virtualgroupmodule.AppModule)
			if !ok {
				panic("*virtualgroupmodule.AppModule not found")
			}
			mm.SetConsensusVersion(2)
			return nil
		})
}

func (app *App) registerHulunbeierPatchUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.HulunbeierPatch,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)

			app.PermissionmoduleKeeper.MigrateAccountPolicyForResources(ctx)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.HulunbeierPatch,
		func() error {
			app.Logger().Info("Init Hulunbeier patch upgrade")
			return nil
		})
}

func (app *App) registerUralUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Ural,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Ural,
		func() error {
			app.Logger().Info("Init Ural upgrade")
			return nil
		})
}

func (app *App) registerPawneeUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Pawnee,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Pawnee,
		func() error {
			app.Logger().Info("Init Pawnee upgrade")
			return nil
		})
}
