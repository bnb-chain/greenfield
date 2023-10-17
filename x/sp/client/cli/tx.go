package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	spTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	spTxCmd.AddCommand(
		CmdCreateStorageProvider(),
		CmdDeposit(),
		CmdEditStorageProvider(),
		CmdGrantDepositAuthorization(),
		CmdUpdateStorageProviderStatus(),
		CmdUpdateStorageProviderStoragePrice(),
	)

	// this line is used by starport scaffolding # 1

	return spTxCmd
}

func CmdCreateStorageProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-storage-provider [path/to/create_storage_provider_proposal.json]",
		Short: "submit a create new storage provider proposal",
		Args:  cobra.ExactArgs(1),
		Long:  `Submit a create new storage provider proposal by submitting a JSON file with the new storage provider details, once the proposal has been passed, create a new storage provider initialized with a self deposit.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx sp create-storage-provider path/to/create_sp_proposal.json --from keyname
Where create_storagep_provider.json contains:
{
  "messages": [
    {
      "@type": "/greenfield.sp.MsgCreateStorageProvider",
      "description": {
        "moniker": "sp0",
        "identity": "",
        "website": "",
        "security_contact": "",
        "details": ""
      },
      "sp_address": "0x012Eadb23D670db68Ba8e67e6F34DE6ACE55b547",
      "funding_address": "0x84b3307313e253eF5787b55616BB1F6F7139C2c0",
      "seal_address": "0xbBD6cD73Cd376c3Dda20de0c4CBD8Fb1Bca2410D",
      "approval_address": "0xdCE01bfaBc7c9c0865bCCeF872493B4BE3b343E8",
      "gc_address": "0x0a1C8982C619B93bA7100411Fc58382306ab431b",
      "maintenance_address": "0xbE03316B1D7c3FCB69136e47e02442d6Fb3396dB",
      "endpoint": "https://sp0.greenfield.io",
      "deposit": {
        "denom": "BNB",
        "amount": "1000000000000000000000"
      },
      "read_price": "0.108", 
      "store_price": "0.016",
      "free_read_quota": 1073741824,
      "creator": "0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2",
      "bls_key": "af8c586885a490a1775bcbef95e6162de1904777f3fb91e3bfd0ffd690fe0d477d0984f11852c64dc77d4583c99f34cb",
      "bls_proof": "8bbce5330c5a46416ec41bfb93d938e8fb2e01d0a4035bd7b87efb98762e5e71faf00427d991003680325b7f97b362640f8e58e69bf774cd59e2267bdfe5a2e6578194b6834531893a39253c718edae2511977991895cdc8dd9e1136e43d721c"
    }
  ],
  "title": "create sp for test",
  "summary": "test",
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "1000000000000000000BNB"
}
modify the related configrations as you need. Example:
1) read_price = $0.09/1024/1024/1024/300(bnb price)*10^18/30/24/60/60 = 0.108 wei/bytes/s
2) store_price = $0.023*(1-6*0.07)/1024/1024/1024/300(bnb price)*10^18/30/24/60/60 = 0.016 wei/bytes/s. (0.07 division for each secondary SP)
3) free_read_quota defines free read quota for each bucket, uint bytes.
`, version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msgs, metadata, title, summary, deposit, err := govcli.ParseSubmitProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			govMsg, err := v1.NewMsgSubmitProposal(msgs, deposit, clientCtx.GetFromAddress().String(), metadata, title, summary)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			if len(msgs) != 1 {
				return fmt.Errorf("invalid message length: %d", len(msgs))
			}

			spMsg, ok := msgs[0].(*types.MsgCreateStorageProvider)
			if !ok || spMsg.ValidateBasic() != nil {
				return fmt.Errorf("invalid create storage provider message")
			}

			fundingAddr, err := sdk.AccAddressFromHexUnsafe(spMsg.FundingAddress)
			if err != nil {
				return err
			}
			if !fundingAddr.Equals(clientCtx.GetFromAddress()) {
				return fmt.Errorf("the from address should be the funding address: %s", fundingAddr.String())
			}

			spAddr, err := sdk.AccAddressFromHexUnsafe(spMsg.SpAddress)
			if err != nil {
				return err
			}

			grantee := authtypes.NewModuleAddress(govtypes.ModuleName)
			authorization := types.NewDepositAuthorization(spAddr, &spMsg.Deposit)
			authzMsg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, nil)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), authzMsg, govMsg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func CmdEditStorageProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-storage-provider [sp-address]",
		Short: "Edit an existing storage provider account",
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

			// maintenance address
			maintenanceAddressStr, err := cmd.Flags().GetString(FlagMaintenanceAddress)
			if err != nil {
				return err
			}
			maintenanceAddress := sdk.AccAddress{}
			if maintenanceAddressStr != "" {
				maintenanceAddress, err = sdk.AccAddressFromHexUnsafe(maintenanceAddressStr)
				if err != nil {
					return err
				}
			}

			// bls key
			blsPubKey, err := cmd.Flags().GetString(FlagBlsPubKey)
			if err != nil {
				return err
			}
			if len(blsPubKey) > 0 && len(blsPubKey) != 2*sdk.BLSPubKeyLength {
				return fmt.Errorf("invalid bls pubkey")
			}

			// bls proof
			blsProof, _ := cmd.Flags().GetString(FlagBlsProof)

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
				maintenanceAddress,
				blsPubKey,
				blsProof,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagEndpoint, "", "The storage provider's endpoint")
	// DescriptionEdit
	cmd.Flags().String(FlagEditMoniker, types.DoNotModifyDesc, "The storage provider's name")
	cmd.Flags().String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	cmd.Flags().String(FlagWebsite, types.DoNotModifyDesc, "The storage provider's (optional) website")
	cmd.Flags().String(FlagDetails, types.DoNotModifyDesc, "The storage provider's (optional) details")

	cmd.Flags().String(FlagSealAddress, "", "The seal address of storage provider")
	cmd.Flags().String(FlagBlsPubKey, "", "The Bls public key of storage provider")
	cmd.Flags().String(FlagBlsProof, "", "The Bls signature of storage provider signing the bls pub key")
	cmd.Flags().String(FlagApprovalAddress, "", "The approval address of storage provider")
	cmd.Flags().String(FlagGcAddress, "", "The gc address of storage provider")
	cmd.Flags().String(FlagMaintenanceAddress, "", "The maintenance address of storage provider")

	return cmd
}

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sp-address] [fund-address] [value]",
		Short: "SP stake tokens from funding account",
		Args:  cobra.ExactArgs(3),
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
		Short: "Grant authorization to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %s tx %s grant [grantee address] send --spend-limit=1000bnb --SPAddress [sp address] --from=sp0_fund
	`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
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
	fsCreateStorageProvider.String(FlagWebsite, "", "The storage provider's (optional) website")
	fsCreateStorageProvider.String(FlagSecurityContact, "", "The storage provider's (optional) security contact email")
	fsCreateStorageProvider.String(FlagDetails, "", "The storage provider's (optional) details")
	fsCreateStorageProvider.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")

	fsCreateStorageProvider.String(FlagCreator, "", "The creator address of storage provider")
	fsCreateStorageProvider.String(FlagSpAddress, "", "The account address of storage provider")
	fsCreateStorageProvider.String(FlagOperatorAddress, "", "The operator address of storage provider")
	fsCreateStorageProvider.String(FlagFundingAddress, "", "The funding address of storage provider")
	fsCreateStorageProvider.String(FlagSealAddress, "", "The seal address of storage provider")
	fsCreateStorageProvider.String(FlagBlsPubKey, "", "The bls public key of storage provider")
	fsCreateStorageProvider.String(FlagBlsProof, "", "The Bls signature of storage provider signing the bls pub key")
	fsCreateStorageProvider.String(FlagApprovalAddress, "", "The approval address of storage provider")
	fsCreateStorageProvider.String(FlagGcAddress, "", "The gc address of storage provider")
	fsCreateStorageProvider.String(FlagMaintenanceAddress, "", "The maintenance address of storage provider")

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

	SpAddress          sdk.AccAddress
	FundingAddress     sdk.AccAddress
	SealAddress        sdk.AccAddress
	BlsPubKey          string
	BlsProof           string
	ApprovalAddress    sdk.AccAddress
	GcAddress          sdk.AccAddress
	MaintenanceAddress sdk.AccAddress

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

	// bls key
	blsPubKey, err := flagSet.GetString(FlagBlsPubKey)
	if err != nil {
		return c, err
	}
	if len(blsPubKey) != 2*sdk.BLSPubKeyLength {
		return c, fmt.Errorf("invalid bls pubkey")
	}
	c.BlsPubKey = blsPubKey

	// bls proof
	blsProof, err := flagSet.GetString(FlagBlsProof)
	if err != nil {
		return c, err
	}
	c.BlsProof = blsProof

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

	// maintenance address
	maintenanceAddress, err := flagSet.GetString(FlagMaintenanceAddress)
	if err != nil {
		return c, err
	}
	addr, err = sdk.AccAddressFromHexUnsafe(maintenanceAddress)
	if err != nil {
		return c, err
	}
	c.MaintenanceAddress = addr

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
		config.SealAddress, config.ApprovalAddress, config.GcAddress, config.MaintenanceAddress, description,
		config.Endpoint, deposit, config.ReadPrice, config.FreeReadQuota, config.StorePrice,
		config.BlsPubKey, config.BlsProof,
	)
	if err != nil {
		return txBldr, msg, err
	}

	return txBldr, msg, nil
}

func CmdUpdateStorageProviderStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status [sp-address] [new-status] [flags]",
		Short: "Update status of a storage provider",
		Long: strings.TrimSpace(
			fmt.Sprintf(`update the storage provider status between STATUS_IN_SERVICE and STATUS_IN_MAINTENANCE, need to provide the maintenance duration in second if status is to STATUS_IN_MAINTENANCE.

Examples:
 $ %s tx %s update-status 0x.... STATUS_IN_MAINTENANCE --duration 21600
	`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			spAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}
			newStatus := args[1]
			duration, err := cmd.Flags().GetInt64(FlagDuration)
			if err != nil {
				return err
			}
			var msg sdk.Msg
			switch newStatus {
			case types.STATUS_IN_SERVICE.String():
				msg = types.NewMsgUpdateStorageProviderStatus(spAddress, types.STATUS_IN_SERVICE, 0)
			case types.STATUS_IN_MAINTENANCE.String():
				msg = types.NewMsgUpdateStorageProviderStatus(spAddress, types.STATUS_IN_MAINTENANCE, duration)
			default:
				return fmt.Errorf("status %s is not expected\n", newStatus)
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagDuration, 0, "maintenance duration requested by a SP")
	return cmd
}

func CmdUpdateStorageProviderStoragePrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-price [sp-address] [read-price] [store-price] [free-read-quota]",
		Short: "Update prices and free read quota of a storage provider, all prices in BNB wei",
		Long: strings.TrimSpace(
			fmt.Sprintf(`update the storage provider read, store price and free read quota, if there is no change to a specific value, the current value should also be provided.

The unit of price is a decimal, which indicates wei BNB per byte per second. 
E.g. the price is 0.02183945725, means approximately $0.018 / GB / Month. 
(0.02183945725 * (30 * 86400) * (1024 * 1024 * 1024) * 300 / 10 ** 18 ≈ 0.018, assume the BNB price is 300 USD)

The free-read-quota unit is bytes, for 1GB free quota, it is 1073741824.

Examples:
 $ %s tx %s update-price 0x... 0.1469890427 0.02183945725 1073741824
	`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			spAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}
			readPrice, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}
			storePrice, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return err
			}
			quota, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}
			msg := types.MsgUpdateSpStoragePrice{
				SpAddress:     spAddress.String(),
				ReadPrice:     readPrice,
				StorePrice:    storePrice,
				FreeReadQuota: quota,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
