package cli

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdDeposit())
	cmd.AddCommand(CmdEditStorageProvider())
	cmd.AddCommand(CmdGrantDepositAuthorization())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdEditStorageProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-storage-provider [sp-address]",
		Short: "Broadcast message editStorageProvider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSpAddress := args[0]

			spAddress, err := sdk.AccAddressFromHexUnsafe(argSpAddress)
			if err != nil {
				return err
			}

			endpoint, _ := cmd.Flags().GetString(FlagEndpoint)
			moniker, _ := cmd.Flags().GetString(FlagEditMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			details, _ := cmd.Flags().GetString(FlagDetails)
			description := types.NewDescription(moniker, identity, website, details)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgEditStorageProvider(
				spAddress,
				endpoint,
				&description,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagEndpoint, types.DoNotModifyDesc, "The storage provider's endpoint")
	// DescriptionEdit
	cmd.Flags().String(FlagEditMoniker, types.DoNotModifyDesc, "The storage provider's name")
	cmd.Flags().String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	cmd.Flags().String(FlagWebsite, types.DoNotModifyDesc, "The storage provider's (optional) website")
	cmd.Flags().String(FlagDetails, types.DoNotModifyDesc, "The storage provider's (optional) details")

	return cmd
}

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sp-address] [value]",
		Short: "Broadcast message deposit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argValue := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO:impl later
			coin := sdk.Coin{Denom: argValue}
			msg := types.NewMsgDeposit(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				coin,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdGrantDepositAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant <grantee> --from <granter>",
		Short: "Broadcast message deposit authorization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}
			var authorization authz.Authorization
			var depositLimit *sdk.Coin

			spAddressString, err := cmd.Flags().GetString(FlagSpAddress)
			if err != nil {
				return err
			}
			spAddress, err := sdk.AccAddressFromHexUnsafe(spAddressString)
			if err != nil {
				return err
			}

			limit, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}
			spendLimit, err := sdk.ParseCoinNormalized(limit)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			if spendLimit.Denom != res.Params.DepositDenom {
				return fmt.Errorf("invalid denom %s; coin denom should match the current bond denom %s", spendLimit.Denom, res.Params.DepositDenom)
			}

			if !spendLimit.IsPositive() {
				return fmt.Errorf("spend-limit should be greater than zero")
			}
			depositLimit = &spendLimit

			authorization, err = types.NewDepositAuthorization(spAddress, depositLimit)
			if err != nil {
				return err
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagSpendLimit, "", "SpendLimit for deposit Authorization, an array of Coins allowed spend")
	cmd.Flags().String(FlagSpAddress, "", "The account address of storage provider")
	cmd.Flags().Int64(FlagExpiration, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry. Default is 0.")
	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(FlagExpiration)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil
	}
	e := time.Unix(exp, 0)
	return &e, nil
}

// CreateStorageProviderMsgFlagSet Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateStorageProviderMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateStorageProvider := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateStorageProvider.String(FlagIP, ipDefault, "The node's public IP")
	fsCreateStorageProvider.String(FlagNodeID, "", "The node's NodeID")

	fsCreateStorageProvider.String(FlagMoniker, "", "The sp's name")
	fsCreateStorageProvider.String(FlagWebsite, "", "The validator's (optional) website")
	fsCreateStorageProvider.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsCreateStorageProvider.String(FlagDetails, "", "The validator's (optional) details")
	fsCreateStorageProvider.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")

	fsCreateStorageProvider.String(FlagCreator, "", "The creator address of storage provider")
	fsCreateStorageProvider.String(FlagSpAddress, "", "The account address of storage provider")
	fsCreateStorageProvider.String(FlagOperatorAddress, "", "The operator address of storage provider")
	fsCreateStorageProvider.String(FlagFundingAddress, "", "The funding address of storage provider")
	fsCreateStorageProvider.String(FlagSealAddress, "", "The seal address of storage provider")
	fsCreateStorageProvider.String(FlagApprovalAddress, "", "The approval address of storage provider")

	fsCreateStorageProvider.String(FlagEndpoint, "", "The storage provider's endpoint")

	return fsCreateStorageProvider, defaultsDesc
}

type TxCreateStorageProviderConfig struct {
	ChainID string
	NodeID  string

	Creator sdk.AccAddress

	Moniker         string
	Identity        string
	Website         string
	SecurityContact string
	Details         string

	SpAddress       sdk.AccAddress
	FundingAddress  sdk.AccAddress
	SealAddress     sdk.AccAddress
	ApprovalAddress sdk.AccAddress

	Endpoint string
	Deposit  string
}

func PrepareConfigForTxCreateStorageProvider(flagSet *flag.FlagSet) (TxCreateStorageProviderConfig, error) {
	c := TxCreateStorageProviderConfig{}

	// Creator
	creator, err := flagSet.GetString(FlagCreator)
	if err != nil {
		return c, err
	}
	addr, err := sdk.AccAddressFromHexUnsafe(creator)
	if err != nil {
		return c, err
	}
	c.Creator = addr

	// Description
	moniker, err := flagSet.GetString(FlagMoniker)
	if err != nil {
		return c, err
	}
	c.Moniker = moniker

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}
	c.Identity = identity

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}
	c.Website = website

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}
	c.SecurityContact = securityContact

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}
	c.Details = details

	// spAddress
	operatorAddress, err := flagSet.GetString(FlagOperatorAddress)
	fmt.Println(operatorAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(operatorAddress)
	if err != nil {
		return c, err
	}
	c.SpAddress = addr

	// funding address
	fundingAddress, err := flagSet.GetString(FlagFundingAddress)
	fmt.Println(fundingAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(fundingAddress)
	if err != nil {
		return c, err
	}
	c.FundingAddress = addr

	// seal address
	sealAddress, err := flagSet.GetString(FlagSealAddress)
	fmt.Println(fundingAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(sealAddress)
	if err != nil {
		return c, err
	}
	c.SealAddress = addr

	// approval address
	approvalAddress, err := flagSet.GetString(FlagApprovalAddress)
	fmt.Println(fundingAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(approvalAddress)
	if err != nil {
		return c, err
	}
	c.ApprovalAddress = addr

	// Endpoint
	endpoint, err := flagSet.GetString(FlagEndpoint)
	if err != nil {
		return c, err
	}
	c.Endpoint = endpoint

	return c, err
}

// BuildCreateStorageProviderMsg makes a new MsgCreateStorageProvider.
func BuildCreateStorageProviderMsg(config TxCreateStorageProviderConfig, txBldr tx.Factory) (tx.Factory, sdk.Msg, error) {
	depositStr := config.Deposit
	deposit, err := sdk.ParseCoinNormalized(depositStr)
	if err != nil {
		return txBldr, nil, err
	}

	description := types.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.Details,
	)

	// TODO may add new flags
	newReadPrice := sdk.NewDec(core.RandInt64(100, 200))
	newStorePrice := sdk.NewDec(core.RandInt64(10000, 20000))

	msg, err := types.NewMsgCreateStorageProvider(
		config.Creator, config.SpAddress, config.FundingAddress,
		config.SealAddress, config.ApprovalAddress, description,
		config.Endpoint, deposit, newReadPrice, 10000, newStorePrice,
	)
	if err != nil {
		return txBldr, msg, err
	}

	return txBldr, msg, nil
}
