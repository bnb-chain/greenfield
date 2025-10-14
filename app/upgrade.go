package app

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bridgemoduletypes "github.com/bnb-chain/greenfield/x/bridge/types"
	paymentmodule "github.com/bnb-chain/greenfield/x/payment"
	paymentmodulekeeper "github.com/bnb-chain/greenfield/x/payment/keeper"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagemodulekeeper "github.com/bnb-chain/greenfield/x/storage/keeper"
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
	app.registerSerengetiUpgradeHandler()
	app.registerErdosUpgradeHandler()
	app.registerVeldUpgradeHandler()
	app.registerMongolianUpgradeHandler()
	app.registerAltaiUpgradeHandler()
	app.registerSavannaUpgradeHandler()
	app.registerTundraUpgradeHandler()
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
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgUpdateObjectContent{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgCancelUpdateObjectContent{}), 1.2e3))
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Pawnee,
		func() error {
			app.Logger().Info("Init Pawnee upgrade")
			return nil
		})
}

func (app *App) registerSerengetiUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Serengeti,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			app.VirtualgroupKeeper.MigrateGlobalVirtualGroupFamiliesForSP(ctx)
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgToggleSPAsDelegatedAgent{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgDelegateCreateObject{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgDelegateUpdateObjectContent{}), 1.2e3))
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgSealObjectV2{}), 1.2e2))
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Serengeti,
		func() error {
			app.Logger().Info("Init Serengeti upgrade")
			return nil
		})
}

func (app *App) registerErdosUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Erdos,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			app.GashubKeeper.SetMsgGasParams(ctx, *gashubtypes.NewMsgGasParamsWithFixedGas(sdk.MsgTypeURL(&storagemoduletypes.MsgSetBucketFlowRateLimit{}), 1.2e3))
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Erdos,
		func() error {
			app.Logger().Info("Init Erdos upgrade")
			executorApp := storagemodulekeeper.NewExecutorApp(app.StorageKeeper, storagemodulekeeper.NewMsgServerImpl(app.StorageKeeper), paymentmodulekeeper.NewMsgServerImpl(app.PaymentKeeper))
			err := app.CrossChainKeeper.RegisterChannel(storagemoduletypes.ExecutorChannel, storagemoduletypes.ExecutorChannelId, executorApp)
			if err != nil {
				app.Logger().Error("register channel error ", err.Error())
			}
			return nil
		})
}

func (app *App) registerVeldUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Veld,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Veld,
		func() error {
			app.Logger().Info("Init Veld upgrade")
			return nil
		})
}

func (app *App) registerMongolianUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Mongolian,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Mongolian,
		func() error {
			app.Logger().Info("Init Mongolian upgrade")
			return nil
		})
}

func (app *App) registerAltaiUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Altai,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Altai,
		func() error {
			app.Logger().Info("Init Altai upgrade")
			return nil
		})
}

func (app *App) registerSavannaUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Savanna,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Savanna,
		func() error {
			app.Logger().Info("Init Savanna upgrade")
			return nil
		})
}

func (app *App) registerTundraUpgradeHandler() {
	// Register the upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(upgradetypes.Tundra,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("upgrade to ", plan.Name)
			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	// Register the upgrade initializer
	app.UpgradeKeeper.SetUpgradeInitializer(upgradetypes.Tundra,
		func() error {
			app.Logger().Info("Init Tundra upgrade")
			return nil
		})
}
