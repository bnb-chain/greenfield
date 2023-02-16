package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
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

	cmd.AddCommand(CmdCreateBucket())
	cmd.AddCommand(CmdDeleteBucket())
	cmd.AddCommand(CmdCreateObject())
	cmd.AddCommand(CmdSealObject())
	cmd.AddCommand(CmdRejectSealObject())
	cmd.AddCommand(CmdDeleteObject())
	cmd.AddCommand(CmdCreateGroup())
	cmd.AddCommand(CmdDeleteGroup())
	cmd.AddCommand(CmdUpdateGroupMember())
	cmd.AddCommand(CmdLeaveGroup())
	cmd.AddCommand(CmdCopyObject())
	cmd.AddCommand(CmdUpdateBucketInfo())
	cmd.AddCommand(CmdCancelCreateObject())
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

			isPublic, err := cmd.Flags().GetBool(FlagPublic)
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

			approver, _ := cmd.Flags().GetString(FlagApprover)
			_, approverName, _, err := GetApproverField(clientCtx.Keyring, approver)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateBucket(
				clientCtx.GetFromAddress(),
				argBucketName,
				isPublic,
				primarySPAcc,
				paymentAcc,
				nil,
			)
			approval, _, err := clientCtx.Keyring.Sign(approverName, msg.GetApprovalBytes())
			if err != nil {
				return err
			}

			msg.PrimarySpApprovalSignature = approval
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagPublic, false, "If true(by default), only owner and grantee can access it. Otherwise, every one have permission to access it.")
	cmd.Flags().String(FlagPaymentAccount, "", "The address of the account used to pay for the read fee. The default is the sender account.")
	cmd.Flags().String(FlagPrimarySP, "", "The operator account address of primarySp")
	cmd.Flags().String(FlagApprover, "", "The approval account address of primarySp")
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
		Use:   "update-read-quota [bucket-name]",
		Short: "Update the meta of bucket, E.g ReadQuota, PaymentAccount",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateBucketInfo(
				clientCtx.GetFromAddress(),
				argBucketName,
				types.READ_QUOTA_FREE,
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
		Use:   "create-object [bucket-name] [object-name]",
		Short: "create a new object in the bucket",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]
			argObjectPath := args[2]
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// read file
			f, err := os.OpenFile(argObjectPath, os.O_RDONLY, 0644)
			if err != nil {
				return err
			}

			// TODO(fynn): calc redundancy hashes. hard code here.
			expectChecksum := make([][]byte, 7)
			buf, _ := io.ReadAll(f)
			h := sha256.New()
			h.Write(buf)
			sum := h.Sum(nil)
			expectChecksum[0] = sum
			expectChecksum[1] = sum
			expectChecksum[2] = sum
			expectChecksum[3] = sum
			expectChecksum[4] = sum
			expectChecksum[5] = sum
			expectChecksum[6] = sum

			contentType := http.DetectContentType(buf)

			isPublic, _ := cmd.Flags().GetBool(FlagPublic)

			msg := types.NewMsgCreateObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
				uint64(len(buf)),
				isPublic,
				expectChecksum,
				contentType,
				[]byte("for-test"),
				nil, // NOTE(fynn): Not specified here.
			)
			primarySP, err := cmd.Flags().GetString(FlagPrimarySP)
			if err != nil {
				return err
			}
			_, spKeyName, _, err := GetPrimarySPField(clientCtx.Keyring, primarySP)
			if err != nil {
				return err
			}
			approval, _, err := clientCtx.Keyring.Sign(spKeyName, msg.GetApprovalBytes())
			if err != nil {
				return err
			}
			msg.PrimarySpApprovalSignature = approval

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagPublic, true, "If true(by default), only owner and grantee can access it. Otherwise, every one have permission to access it.")
	cmd.Flags().String(FlagPrimarySP, "", "The operator account address of primarySp")
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

			msg := types.NewMsgCopyObject(
				clientCtx.GetFromAddress(),
				argSrcBucketName,
				argDstBucketName,
				argSrcObjectName,
				argDstObjectName,
				nil, // TODO: Refine the cli parameters
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

func CmdSealObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seal-object [bucket-name] [object-name]",
		Short: "Seal the object",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO(fynn): hardcode here, impl after signature ready
			spSignatures := make([][]byte, 7)
			for i := 0; i < len(spSignatures); i++ {
				spSignatures[i] = []byte("for-test")
			}
			msg := types.NewMsgSealObject(
				clientCtx.GetFromAddress(),
				argBucketName,
				argObjectName,
				nil,
				spSignatures,
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

func CmdRejectSealObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject-seal-object [bucket-name] [object-name]",
		Short: "Reject to seal the object",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRejectUnsealedObject(
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

func CmdCreateGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-group [group-name]",
		Short: "Create a new group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateGroup(
				clientCtx.GetFromAddress(),
				argGroupName,
				nil, // TODO: Refine the cli parameters
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
		Use:   "leave-group [group-name]",
		Short: "Leave the group you're a member of",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgLeaveGroup(
				clientCtx.GetFromAddress(),
				sdk.AccAddress{}, // TODO: add group owner acc
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
		Use:   "update-group-member [group-name]",
		Short: "Update the member of the group you own",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argGroupName := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateGroupMember(
				clientCtx.GetFromAddress(),
				argGroupName,
				nil, // TODO: Refine the cli parameters
				nil, // TODO: Refine the cli parameters
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
