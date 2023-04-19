package cli

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	cmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	storageTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	storageTxCmd.AddCommand(
		CmdCreateBucket(),
		CmdDeleteBucket(),
		CmdUpdateBucketInfo(),
		CmdMirrorBucket(),
		CmdDiscontinueBucket(),
	)

	storageTxCmd.AddCommand(
		CmdCreateObject(),
		CmdDeleteObject(),
		CmdCancelCreateObject(),
		CmdCopyObject(),
		CmdMirrorObject(),
		CmdDiscontinueObject(),
		CmdUpdateObjectInfo(),
	)

	storageTxCmd.AddCommand(
		CmdCreateGroup(),
		CmdDeleteGroup(),
		CmdUpdateGroupMember(),
		CmdLeaveGroup(),
		CmdMirrorGroup(),
	)

	storageTxCmd.AddCommand(
		CmdPutPolicy(),
		CmdDeletePolicy(),
	)

	// this line is used by starport scaffolding # 1

	return storageTxCmd
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
		Short: "create a new object in the bucket, checksums split by ','",
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
				nil,
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
		Use:   "create-group [group-name] [member-list]",
		Short: "Create a new group with several initial members, split member addresses by ','",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			argMemberList := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var memberAddrs []sdk.AccAddress
			members := strings.Split(argMemberList, ",")
			for _, member := range members {
				memberAddr, err := sdk.AccAddressFromHexUnsafe(member)
				if err != nil {
					return err
				}
				memberAddrs = append(memberAddrs, memberAddr)
			}
			msg := types.NewMsgCreateGroup(
				clientCtx.GetFromAddress(),
				argGroupName,
				memberAddrs,
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
		Use:   "update-group-member [group-name] [member-to-add] [member-to-delete]",
		Short: "Update the member of the group you own, split member addresses by ,",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]
			argMemberToAdd := args[1]
			argMemberToDelete := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			var memberAddrsToAdd []sdk.AccAddress
			membersToAdd := strings.Split(argMemberToAdd, ",")
			for _, member := range membersToAdd {
				memberAddr, err := sdk.AccAddressFromHexUnsafe(member)
				if err != nil {
					return err
				}
				memberAddrsToAdd = append(memberAddrsToAdd, memberAddr)
			}
			var memberAddrsToDelete []sdk.AccAddress
			membersToDelete := strings.Split(argMemberToDelete, ",")
			for _, member := range membersToDelete {
				memberAddr, err := sdk.AccAddressFromHexUnsafe(member)
				if err != nil {
					return err
				}
				memberAddrsToDelete = append(memberAddrsToDelete, memberAddr)
			}
			msg := types.NewMsgUpdateGroupMember(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				argGroupName,
				memberAddrsToAdd,
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

func CmdPutPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put-policy [operator] [principal-type] [principle-value] [resource]",
		Short: "put a policy to bucket/object/group which can grant permission to others",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			//argOperator := args[0]
			argPrincipalType := args[1]
			argPrincipalValue := args[2]
			argResource := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			principalType, err := GetPrincipalType(argPrincipalType)
			if err != nil {
				return err
			}

			principal := permtypes.Principal{
				Type:  principalType,
				Value: argPrincipalValue,
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
		Use:   "delete-policy [principal-type] [principle-value] [resource]",
		Short: "Broadcast message delete-policy",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPrincipalType := args[0]
			argPrincipalValue := args[1]
			argResource := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			principalType, err := GetPrincipalType(argPrincipalType)
			if err != nil {
				return err
			}

			principal := permtypes.Principal{
				Type:  principalType,
				Value: argPrincipalValue,
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
		Use:   "mirror-bucket [bucket-id]",
		Short: "Mirror an existing bucket to the destination chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketId := args[0]

			bucketId, ok := big.NewInt(0).SetString(argBucketId, 10)
			if !ok {
				return fmt.Errorf("invalid bucket id: %s", argBucketId)
			}
			if bucketId.Cmp(big.NewInt(0)) < 0 {
				return fmt.Errorf("bucket id should not be negative")
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorBucket(
				clientCtx.GetFromAddress(),
				cmath.NewUintFromBigInt(bucketId),
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
		Use:   "mirror-object [object-id]",
		Short: "Mirror the object to the destination chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argObjectId := args[0]

			objectId, ok := big.NewInt(0).SetString(argObjectId, 10)
			if !ok {
				return fmt.Errorf("invalid object id: %s", argObjectId)
			}
			if objectId.Cmp(big.NewInt(0)) < 0 {
				return fmt.Errorf("object id should not be negative")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorObject(
				clientCtx.GetFromAddress(),
				cmath.NewUintFromBigInt(objectId),
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

func CmdMirrorGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror-group [group-id]",
		Short: "Mirror an existing group to the destination chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupId := args[0]

			groupId, ok := big.NewInt(0).SetString(argGroupId, 10)
			if !ok {
				return fmt.Errorf("invalid groupd id: %s", argGroupId)
			}
			if groupId.Cmp(big.NewInt(0)) < 0 {
				return fmt.Errorf("groupd id should not be negative")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMirrorGroup(
				clientCtx.GetFromAddress(),
				cmath.NewUintFromBigInt(groupId),
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
