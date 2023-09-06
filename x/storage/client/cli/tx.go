package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	cmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/x/storage/types"
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

	cmd.AddCommand(
		CmdCreateBucket(),
		CmdDeleteBucket(),
		CmdUpdateBucketInfo(),
		CmdMirrorBucket(),
		CmdDiscontinueBucket(),
	)

	cmd.AddCommand(
		CmdCreateObject(),
		CmdDeleteObject(),
		CmdCancelCreateObject(),
		CmdCopyObject(),
		CmdMirrorObject(),
		CmdDiscontinueObject(),
		CmdUpdateObjectInfo(),
	)

	cmd.AddCommand(
		CmdCreateGroup(),
		CmdDeleteGroup(),
		CmdUpdateGroupMember(),
		CmdUpdateGroupExtra(),
		CmdRenewGroupMember(),
		CmdLeaveGroup(),
		CmdMirrorGroup(),
	)

	cmd.AddCommand(
		CmdPutPolicy(),
		CmdDeletePolicy(),
	)

	cmd.AddCommand(CmdCancelMigrateBucket())
	// this line is used by starport scaffolding # 1

	return cmd
}

// CmdCreateBucket returns a CLI command handler for creating a MsgCreateBucket transaction.
func CmdCreateBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-bucket [bucket-name]",
		Short: "create a new bucket which associate to a primary sp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			argBucketName := args[0]

			visibility, err := cmd.Flags().GetString(FlagVisibility)
			if err != nil {
				return err
			}
			visibilityType, err := GetVisibilityType(visibility)
			if err != nil {
				return err
			}

			chargedReadQuota, err := cmd.Flags().GetUint64(FlagChargedReadQuota)
			if err != nil {
				return err
			}

			payment, _ := cmd.Flags().GetString(FlagPaymentAccount)
			paymentAcc, _, _, err := GetPaymentAccountField(clientCtx.Keyring, payment)
			if err != nil {
				return err
			}

			primarySP, _ := cmd.Flags().GetString(FlagPrimarySP)
			primarySPAcc, _, _, err := GetPrimarySPField(clientCtx.Keyring, primarySP)
			if err != nil {
				return err
			}

			approveSignature, _ := cmd.Flags().GetString(FlagApproveSignature)
			approveTimeoutHeight, _ := cmd.Flags().GetUint64(FlagApproveTimeoutHeight)

			approveSignatureBytes, err := hex.DecodeString(approveSignature)
			if err != nil {
				return err
			}
			msg := types.NewMsgCreateBucket(
				clientCtx.GetFromAddress(),
				argBucketName,
				visibilityType,
				primarySPAcc,
				paymentAcc,
				approveTimeoutHeight,
				approveSignatureBytes,
				chargedReadQuota,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetVisibility())
	cmd.Flags().AddFlagSet(FlagSetApproval())
	cmd.Flags().String(FlagPaymentAccount, "", "The address of the account used to pay for the read fee. The default is the sender account.")
	cmd.Flags().String(FlagPrimarySP, "", "The operator account address of primarySp")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-bucket [bucket-name]",
		Short: "delete an existing bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteBucket(
				clientCtx.GetFromAddress(),
				argBucketName,
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

func CmdUpdateBucketInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-bucket-info [bucket-name] [charged-read-quota]",
		Short: "Update the meta of bucket, E.g ChargedReadQuota, PaymentAccount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argChargedReadQuota, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			visibility, err := cmd.Flags().GetString(FlagVisibility)
			if err != nil {
				return err
			}
			visibilityType, err := GetVisibilityType(visibility)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateBucketInfo(
				clientCtx.GetFromAddress(),
				argBucketName,
				&argChargedReadQuota,
				nil,
				visibilityType,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetVisibility())

	return cmd
}

func CmdCancelCreateObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-create-object [bucket-name] [object-name]",
		Short: "Broadcast message cancel_create_object",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelCreateObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
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

func CmdCreateObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-object [bucket-name] [object-name] [payload-size] [content-type]",
		Short: "Create a new object in the bucket, checksums split by ','",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]
			argPayloadSize := args[2]
			argContentType := args[3]

			payloadSize, err := strconv.ParseUint(argPayloadSize, 10, 64)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			visibility, err := cmd.Flags().GetString(FlagVisibility)
			if err != nil {
				return err
			}
			visibilityType, err := GetVisibilityType(visibility)
			if err != nil {
				return err
			}

			checksums, _ := cmd.Flags().GetString(FlagExpectChecksums)
			redundancyTypeFlag, _ := cmd.Flags().GetString(FlagRedundancyType)
			approveSignature, _ := cmd.Flags().GetString(FlagApproveSignature)
			approveTimeoutHeight, _ := cmd.Flags().GetUint64(FlagApproveTimeoutHeight)

			approveSignatureBytes, err := hex.DecodeString(approveSignature)
			if err != nil {
				return err
			}

			checksumsStr := strings.Split(checksums, ",")
			if checksumsStr == nil {
				return gnfderrors.ErrInvalidChecksum
			}
			var expectChecksums [][]byte
			for _, checksum := range checksumsStr {
				tmp, err := hex.DecodeString(checksum)
				if err != nil {
					return err
				}
				expectChecksums = append(expectChecksums, tmp)
			}

			var redundancyType types.RedundancyType
			if redundancyTypeFlag == "EC" {
				redundancyType = types.REDUNDANCY_EC_TYPE
			} else if redundancyTypeFlag == "Replica" {
				redundancyType = types.REDUNDANCY_REPLICA_TYPE
			} else {
				return types.ErrInvalidRedundancyType
			}

			msg := types.NewMsgCreateObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
				payloadSize,
				visibilityType,
				expectChecksums,
				argContentType,
				redundancyType,
				approveTimeoutHeight,
				approveSignatureBytes,
			)
			primarySP, err := cmd.Flags().GetString(FlagPrimarySP)
			if err != nil {
				return err
			}
			_, spKeyName, _, err := GetPrimarySPField(clientCtx.Keyring, primarySP)
			if err != nil {
				return err
			}
			sig, _, err := clientCtx.Keyring.Sign(spKeyName, msg.GetApprovalBytes())
			if err != nil {
				return err
			}
			msg.PrimarySpApproval.Sig = sig

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetVisibility())
	cmd.Flags().AddFlagSet(FlagSetApproval())
	cmd.Flags().String(FlagPrimarySP, "", "The operator account address of primarySp")
	cmd.Flags().String(FlagExpectChecksums, "", "The checksums that calculate by redundancy algorithm")
	cmd.Flags().String(FlagRedundancyType, "", "The redundancy type, EC or Replica ")
	return cmd
}

func CmdCopyObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy-object [src-bucket-name] [dst-bucket-name] [src-object-name] [dst-object-name]",
		Short: "Copy an existing object in a bucket to another bucket",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSrcBucketName := args[0]
			argDstBucketName := args[1]
			argSrcObjectName := args[2]
			argDstObjectName := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			approveSignature, _ := cmd.Flags().GetString(FlagApproveSignature)
			approveTimeoutHeight, _ := cmd.Flags().GetUint64(FlagApproveTimeoutHeight)

			approveSignatureBytes, err := hex.DecodeString(approveSignature)
			if err != nil {
				return err
			}
			msg := types.NewMsgCopyObject(
				clientCtx.GetFromAddress(),
				argSrcBucketName,
				argDstBucketName,
				argSrcObjectName,
				argDstObjectName,
				approveTimeoutHeight,
				approveSignatureBytes,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetApproval())

	return cmd
}

func CmdDeleteObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-object [bucket-name] [object-name]",
		Short: "Delete an existing object",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
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

func CmdUpdateObjectInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-object-info [bucket-name] [object-name] [visibility]",
		Short: "Update the meta of object, Currently only support: Visibility",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]

			visibility, err := cmd.Flags().GetString(FlagVisibility)
			if err != nil {
				return err
			}
			visibilityType, err := GetVisibilityType(visibility)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateObjectInfo(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
				visibilityType,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetVisibility())

	return cmd
}

func CmdDiscontinueObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discontinue-object [bucket-name] [object-ids] [reason]",
		Short: "Discontinue to store objects",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectIds := args[1]
			argObjectReason := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			objectIds := make([]cmath.Uint, 0)
			splitIds := strings.Split(argObjectIds, ",")
			for _, split := range splitIds {
				id, ok := big.NewInt(0).SetString(split, 10)
				if !ok {
					return fmt.Errorf("invalid object id: %s", id)
				}
				if id.Cmp(big.NewInt(0)) < 0 {
					return fmt.Errorf("object id should not be negative")
				}

				objectIds = append(objectIds, cmath.NewUintFromBigInt(id))
			}

			msg := types.NewMsgDiscontinueObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				objectIds,
				argObjectReason,
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

func CmdCreateGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-group [group-name]",
		Short: "Create a new group with optional members, split member addresses by ','",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			extra, _ := cmd.Flags().GetString(FlagExtra)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateGroup(
				clientCtx.GetFromAddress(),
				argGroupName,
				extra,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExtra, "", "extra info for the group")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-group [group-name]",
		Short: "Delete an existing group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteGroup(
				clientCtx.GetFromAddress(),
				argGroupName,
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

func CmdLeaveGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave-group [group-owner] [group-name]",
		Short: "Leave the group you're a member of",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupOwner := args[0]
			argGroupName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupOwner, err := sdk.AccAddressFromHexUnsafe(argGroupOwner)
			if err != nil {
				return err
			}

			msg := types.NewMsgLeaveGroup(
				clientCtx.GetFromAddress(),
				groupOwner,
				argGroupName,
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

func CmdUpdateGroupMember() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-member [group-name] [member-to-add] [member-expiration-to-add] [member-to-delete]",
		Short: "Update the member of the group you own, split member addresses and expiration(UNIX timestamp) by ,",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			argMemberToAdd := args[1]
			argMemberExpirationToAdd := args[2]
			argMemberToDelete := args[3]
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			membersToAdd := strings.Split(argMemberToAdd, ",")
			memberExpirationStr := strings.Split(argMemberExpirationToAdd, ",")
			if len(memberExpirationStr) != len(membersToAdd) {
				return errors.New("[member-to-add] and [member-expiration-to-add] should have the same length")
			}

			msgGroupMemberToAdd := make([]*types.MsgGroupMember, 0, len(argMemberToAdd))
			if len(membersToAdd) > 0 {
				for i := range membersToAdd {
					if len(membersToAdd[i]) > 0 {
						_, err := sdk.AccAddressFromHexUnsafe(membersToAdd[i])
						if err != nil {
							return err
						}
						member := types.MsgGroupMember{
							Member: membersToAdd[i],
						}
						if len(memberExpirationStr[i]) > 0 {
							unix, err := strconv.ParseInt(memberExpirationStr[i], 10, 64)
							if err != nil {
								return err
							}
							expiration := time.Unix(unix, 0)
							member.ExpirationTime = &expiration
						}

						msgGroupMemberToAdd = append(msgGroupMemberToAdd, &member)
					}
				}
			}

			var memberAddrsToDelete []sdk.AccAddress
			if len(argMemberToDelete) == 0 {
				membersToDelete := strings.Split(argMemberToDelete, ",")
				for _, member := range membersToDelete {
					if len(member) > 0 {
						memberAddr, err := sdk.AccAddressFromHexUnsafe(member)
						if err != nil {
							return err
						}
						memberAddrsToDelete = append(memberAddrsToDelete, memberAddr)
					}
				}
			}
			msg := types.NewMsgUpdateGroupMember(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				argGroupName,
				msgGroupMemberToAdd,
				memberAddrsToDelete,
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

func CmdRenewGroupMember() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "renew-group-member [group-name] [member] [member-expiration]",
		Short: "renew the member of the group you own, split member-addresses and member-expiration(UNIX timestamp) by ,",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			argMember := args[1]
			argMemberExpiration := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			memberExpirationStr := strings.Split(argMemberExpiration, ",")
			members := strings.Split(argMember, ",")

			if len(memberExpirationStr) != len(members) {
				return errors.New("member and member-expiration should have the same length")
			}

			msgGroupMember := make([]*types.MsgGroupMember, 0, len(argMember))
			for i := range members {
				if len(members[i]) > 0 {
					_, err := sdk.AccAddressFromHexUnsafe(members[i])
					if err != nil {
						return err
					}
					member := types.MsgGroupMember{
						Member: members[i],
					}
					if len(memberExpirationStr[i]) > 0 {
						unix, err := strconv.ParseInt(memberExpirationStr[i], 10, 64)
						if err != nil {
							return err
						}
						expiration := time.Unix(unix, 0)
						member.ExpirationTime = &expiration
					}

					msgGroupMember = append(msgGroupMember, &member)
				}
			}

			msg := types.NewMsgRenewGroupMember(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				argGroupName,
				msgGroupMember,
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

func CmdUpdateGroupExtra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-extra [group-name] [extra]",
		Short: "Update the extra info of the group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			argExtra := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgUpdateGroupExtra(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				argGroupName,
				argExtra,
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

func CmdPutPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put-policy [principle-value] [resource]",
		Short: "put a policy to bucket/object/group which can grant permission to others",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPrincipalValue := args[0]
			argResource := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			principal, err := GetPrincipal(argPrincipalValue)
			if err != nil {
				return err
			}

			msg := types.NewMsgPutPolicy(
				clientCtx.GetFromAddress(),
				argResource,
				&principal,
				nil,
				nil,
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

func CmdDeletePolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-policy [principle-value] [resource]",
		Short: "Delete policy with specify principle",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delete the policy, the principle-value can be account or group id.

Example:
$ %s tx storage delete-policy 0xffffffffffffffffffffff
$ %s tx delete-policy 3
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPrincipalValue := args[0]
			argResource := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			principal, err := GetPrincipal(argPrincipalValue)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeletePolicy(
				clientCtx.GetFromAddress(),
				argResource,
				&principal,
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

func CmdMirrorBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror-bucket",
		Short: "Mirror an existing bucket to the destination chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketId, _ := cmd.Flags().GetString(FlagBucketId)
			argBucketName, _ := cmd.Flags().GetString(FlagBucketName)
			argDestChainId, _ := cmd.Flags().GetString(FlagDestChainId)

			bucketId := big.NewInt(0)
			if argBucketId == "" && argBucketName == "" {
				return fmt.Errorf("bucket id or bucket name should be provided")
			} else if argBucketId != "" && argBucketName != "" {
				return fmt.Errorf("bucket id and bucket name should not be provided together")
			} else if argBucketId != "" {
				ok := false
				bucketId, ok = big.NewInt(0).SetString(argBucketId, 10)
				if !ok {
					return fmt.Errorf("invalid bucket id: %s", argBucketId)
				}
				if bucketId.Cmp(big.NewInt(0)) <= 0 {
					return fmt.Errorf("bucket id should be positive")
				}
			}

			if argDestChainId == "" {
				return fmt.Errorf("destination chain id should be provided")
			}
			destChainId, err := strconv.ParseUint(argDestChainId, 10, 16)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorBucket(
				clientCtx.GetFromAddress(),
				sdk.ChainID(destChainId),
				cmath.NewUintFromBigInt(bucketId),
				argBucketName,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagBucketId, "", "Id of the bucket to mirror")
	cmd.Flags().String(FlagBucketName, "", "Name of the bucket to mirror")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDiscontinueBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discontinue-bucket [bucket-name] [reason]",
		Short: "Discontinue to store bucket",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectReason := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDiscontinueBucket(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectReason,
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

func CmdMirrorObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror-object",
		Short: "Mirror the object to the destination chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argObjectId, _ := cmd.Flags().GetString(FlagObjectId)
			argBucketName, _ := cmd.Flags().GetString(FlagBucketName)
			argObjectName, _ := cmd.Flags().GetString(FlagObjectName)
			argDestChainId, _ := cmd.Flags().GetString(FlagDestChainId)

			objectId := big.NewInt(0)
			if argObjectId == "" && argObjectName == "" {
				return fmt.Errorf("object id or object name should be provided")
			} else if argObjectId != "" && argObjectName != "" {
				return fmt.Errorf("object id and object name should not be provided together")
			} else if argObjectId != "" {
				ok := false
				objectId, ok = big.NewInt(0).SetString(argObjectId, 10)
				if !ok {
					return fmt.Errorf("invalid object id: %s", argObjectId)
				}
				if objectId.Cmp(big.NewInt(0)) <= 0 {
					return fmt.Errorf("object id should be positive")
				}
			} else if argObjectName != "" && argBucketName == "" {
				return fmt.Errorf("object name and bucket name should not be provided together")
			}

			if argDestChainId == "" {
				return fmt.Errorf("destination chain id should be provided")
			}
			destChainId, err := strconv.ParseUint(argDestChainId, 10, 16)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorObject(
				clientCtx.GetFromAddress(),
				sdk.ChainID(destChainId),
				cmath.NewUintFromBigInt(objectId),
				argBucketName,
				argObjectName,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagObjectId, "", "Id of the object to mirror")
	cmd.Flags().String(FlagObjectName, "", "Name of the object to mirror")
	cmd.Flags().String(FlagBucketName, "", "Name of the bucket that the object belongs to")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdMirrorGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror-group",
		Short: "Mirror an existing group to the destination chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupId, _ := cmd.Flags().GetString(FlagGroupId)
			argGroupName, _ := cmd.Flags().GetString(FlagGroupName)
			argDestChainId, _ := cmd.Flags().GetString(FlagDestChainId)

			groupId := big.NewInt(0)
			if argGroupId == "" && argGroupName == "" {
				return fmt.Errorf("group id or group name should be provided")
			} else if argGroupId != "" && argGroupName != "" {
				return fmt.Errorf("group id and group name should not be provided together")
			} else if argGroupId != "" {
				ok := false
				groupId, ok = big.NewInt(0).SetString(argGroupId, 10)
				if !ok {
					return fmt.Errorf("invalid groupd id: %s", argGroupId)
				}
				if groupId.Cmp(big.NewInt(0)) <= 0 {
					return fmt.Errorf("groupd id should be positive")
				}
			}

			if argDestChainId == "" {
				return fmt.Errorf("destination chain id should be provided")
			}
			destChainId, err := strconv.ParseUint(argDestChainId, 10, 16)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorGroup(
				clientCtx.GetFromAddress(),
				sdk.ChainID(destChainId),
				cmath.NewUintFromBigInt(groupId),
				argGroupName,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagGroupId, "", "Id of the group to mirror")
	cmd.Flags().String(FlagGroupName, "", "Name of the group to mirror")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
