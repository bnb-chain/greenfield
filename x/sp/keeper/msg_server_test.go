package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestMsgCreateStorageProvider() {
	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	// 1. create new newStorageProvider and grant

	operatorAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")
	fundingAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")
	sealAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")
	approvalAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")
	gcAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "error should be nil")

	blsPubKeyHex := sample.RandBlsPubKeyHex()

	s.accountKeeper.EXPECT().GetAccount(gomock.Any(), fundingAddr).Return(authtypes.NewBaseAccountWithAddress(fundingAddr)).AnyTimes()
	s.accountKeeper.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCreateStorageProvider
	}{
		{
			Name:      "invalid funding address",
			ExceptErr: true,
			req: types.MsgCreateStorageProvider{
				Creator: govAddr.String(),
				Description: sptypes.Description{
					Moniker:  "sp_test",
					Identity: "",
				},
				SpAddress:       operatorAddr.String(),
				FundingAddress:  sample.AccAddress(),
				SealAddress:     sealAddr.String(),
				ApprovalAddress: approvalAddr.String(),
				GcAddress:       gcAddr.String(),
				SealBlsKey:      blsPubKeyHex,
				Deposit: sdk.Coin{
					Denom:  types.Denom,
					Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
				},
			},
		},
		{
			Name:      "invalid endpoint",
			ExceptErr: true,
			req: types.MsgCreateStorageProvider{
				Creator: govAddr.String(),
				Description: sptypes.Description{
					Moniker:  "sp_test",
					Identity: "",
				},
				SpAddress:       operatorAddr.String(),
				FundingAddress:  fundingAddr.String(),
				SealAddress:     sealAddr.String(),
				ApprovalAddress: approvalAddr.String(),
				GcAddress:       gcAddr.String(),
				SealBlsKey:      blsPubKeyHex,
				Endpoint:        "sp.io",
				Deposit: sdk.Coin{
					Denom:  types.Denom,
					Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
				},
			},
		},
		{
			Name:      "invalid bls pub key",
			ExceptErr: true,
			req: types.MsgCreateStorageProvider{
				Creator: govAddr.String(),
				Description: sptypes.Description{
					Moniker:  "sp_test",
					Identity: "",
				},
				SpAddress:       operatorAddr.String(),
				FundingAddress:  fundingAddr.String(),
				SealAddress:     sealAddr.String(),
				ApprovalAddress: approvalAddr.String(),
				GcAddress:       gcAddr.String(),
				SealBlsKey:      "InValidBlsPubkey",
				Endpoint:        "sp.io",
				Deposit: sdk.Coin{
					Denom:  types.Denom,
					Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
				},
			},
		},
		{
			Name:      "success",
			ExceptErr: true,
			req: types.MsgCreateStorageProvider{
				Creator: govAddr.String(),
				Description: sptypes.Description{
					Moniker:  "MsgServer_sp_test",
					Identity: "",
				},
				SpAddress:       operatorAddr.String(),
				FundingAddress:  fundingAddr.String(),
				SealAddress:     sealAddr.String(),
				ApprovalAddress: approvalAddr.String(),
				GcAddress:       gcAddr.String(),
				SealBlsKey:      blsPubKeyHex,
				Deposit: sdk.Coin{
					Denom:  types.Denom,
					Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
				},
			},
		},
	}
	for _, testCase := range testCases {
		s.Suite.T().Run(testCase.Name, func(t *testing.T) {
			_, err := s.msgServer.CreateStorageProvider(s.ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})

	}

}
