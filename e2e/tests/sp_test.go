package tests

import (
	"testing"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type StorageProviderTestSuite struct {
	core.BaseSuite
}

func (s *StorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageProviderTestSuite) SetupTest() {
}

func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
	deposit := sdk.Coin{
		Denom:  "bnb",
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp0",
		Identity: "",
	}
	// CreateStorageProvider
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	msgCreateStorageProvider, err := sptypes.NewMsgCreateStorageProvider(sdk.AccAddress("0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2"), s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.FundingKey.GetAddr(),
		s.StorageProvider.SealKey.GetAddr(),
		s.StorageProvider.ApprovalKey.GetAddr(), description,
		"sp0.greenfield.io", deposit)
	//msgCreateSP

	s.Require().NoError(err)
	s.SendTxBlock(msgCreateStorageProvider, user)
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
