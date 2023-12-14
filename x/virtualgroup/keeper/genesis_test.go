package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/x/virtualgroup"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *TestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	s.accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), gomock.Any()).Return(types2.NewEmptyModuleAccount(types.ModuleName))
	s.accountKeeper.EXPECT().SetModuleAccount(gomock.Any(), gomock.Any()).Return()
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), gomock.Any()).Return(sdk.NewCoins(sdk.NewCoin(genesisState.Params.DepositDenom, sdk.ZeroInt())))
	virtualgroup.InitGenesis(s.ctx, *s.virtualgroupKeeper, genesisState)

	got := virtualgroup.ExportGenesis(s.ctx, *s.virtualgroupKeeper)
	s.Require().NotNil(got)
	s.Require().Equal(genesisState.Params, got.Params)
}
