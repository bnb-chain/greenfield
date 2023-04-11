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

			// seal address
			sealAddressStr, err := cmd.Flags().GetString(FlagSealAddress)
			if err != nil {
				return err
			}
			sealAddress := sdk.AccAddress{}
			if sealAddressStr != "" {
				sealAddress, err = sdk.AccAddressFromHexUnsafe(sealAddressStr)
				if err != nil {
					return err
				}
			}

			// approval address
			approvalAddressStr, err := cmd.Flags().GetString(FlagApprovalAddress)
			if err != nil {
				return err
			}
			approvalAddress := sdk.AccAddress{}
			if approvalAddressStr != "" {
				approvalAddress, err = sdk.AccAddressFromHexUnsafe(approvalAddressStr)
				if err != nil {
					return err
				}
			}

			// gc address
			gcAddressStr, err := cmd.Flags().GetString(FlagGcAddress)
			if err != nil {
				return err
			}
			gcAddress := sdk.AccAddress{}
			if gcAddressStr != "" {
				gcAddress, err = sdk.AccAddressFromHexUnsafe(gcAddressStr)
				if err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgEditStorageProvider(
				spAddress,
				endpoint,
				&description,
				sealAddress,
				approvalAddress,
				gcAddress,
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

	cmd.Flags().String(FlagSealAddress, "", "The seal address of storage provider")
	cmd.Flags().String(FlagApprovalAddress, "", "The approval address of storage provider")
	cmd.Flags().String(FlagGcAddress, "", "The gc address of storage provider")

	return cmd
}

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sp-address] [fund-address] [value]",
		Short: "Broadcast message deposit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			spAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}

			fundAddress, err := sdk.AccAddressFromHexUnsafe(args[1])
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgDeposit(
				fundAddress,
				spAddress,
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

			authorization = types.NewDepositAuthorization(spAddress, depositLimit)

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
	fsCreateStorageProvider.String(FlagGcAddress, "", "The gc address of storage provider")

	fsCreateStorageProvider.String(FlagEndpoint, "", "The storage provider's endpoint")

	// payment
	fsCreateStorageProvider.String(FlagReadPrice, "100", "The storage provider's read price, in bnb wei per charge byte")
	fsCreateStorageProvider.Uint64(FlagFreeReadQuota, 10000, "The storage provider's free read quota, in byte")
	fsCreateStorageProvider.String(FlagStorePrice, "10000", "The storage provider's store price, in bnb wei per charge byte")

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
	GcAddress       sdk.AccAddress

	Endpoint string
	Deposit  string

	ReadPrice     sdk.Dec
	FreeReadQuota uint64
	StorePrice    sdk.Dec
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
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(approvalAddress)
	if err != nil {
		return c, err
	}
	c.ApprovalAddress = addr

	// gc address
	gcAddress, err := flagSet.GetString(FlagGcAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(gcAddress)
	if err != nil {
		return c, err
	}
	c.GcAddress = addr

	// Endpoint
	endpoint, err := flagSet.GetString(FlagEndpoint)
	if err != nil {
		return c, err
	}
	c.Endpoint = endpoint

	// payment
	readPriceStr, _ := flagSet.GetString(FlagReadPrice)
	if readPriceStr != "" {
		readPrice, err := sdk.NewDecFromStr(readPriceStr)
		if err != nil {
			return c, err
		}

		c.ReadPrice = readPrice
	}
	freeReadQuota, err := flagSet.GetUint64(FlagFreeReadQuota)
	if err != nil {
		return c, err
	}
	c.FreeReadQuota = freeReadQuota

	storePriceStr, _ := flagSet.GetString(FlagStorePrice)
	if storePriceStr != "" {
		storePrice, err := sdk.NewDecFromStr(storePriceStr)
		if err != nil {
			return c, err
		}

		c.StorePrice = storePrice
	}

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

	msg, err := types.NewMsgCreateStorageProvider(
		config.Creator, config.SpAddress, config.FundingAddress,
		config.SealAddress, config.ApprovalAddress, config.GcAddress, description,
		config.Endpoint, deposit, config.ReadPrice, config.FreeReadQuota, config.StorePrice,
	)
	if err != nil {
		return txBldr, msg, err
	}

	return txBldr, msg, nil
}
