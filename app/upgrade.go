package app

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func UpgradeInitializerAndHandler(
	accountKeeper authkeeper.AccountKeeper,
) (map[string]upgradetypes.UpgradeInitializer, map[string]upgradetypes.UpgradeHandler) {
	upgradeInitializer := map[string]upgradetypes.UpgradeInitializer{
		// Add specific actions when restarting and upgrade happen
		// ex.
		// uprgadeTypes.BEP111: BEP111Initializer(accountKeeper),
	}
	upgradeHandler := map[string]upgradetypes.UpgradeHandler{
		// Add specific actions when the upgrade happen
		// ex.
		// uprgadeTypes.BEP111: BEP111Handler(accountKeeper),
	}

	return upgradeInitializer, upgradeHandler
}

// func BEP111Initializer(
// 	accountKeeper authkeeper.AccountKeeper,
// ) upgradetypes.UpgradeInitializer {
// 	return func() error {
// 		return nil
// 	}
// }

// func BEP111Handler(
// 	accountKeeper authkeeper.AccountKeeper,
// ) upgradetypes.UpgradeHandler {
// 	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
// 		return fromVM, nil
// 	}
// }
