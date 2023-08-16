package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tmos "github.com/cometbft/cometbft/libs/os"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/sp/client/cli"
)

// SPGenTxCmd builds the application's gentx command.
func SPGenTxCmd(mbm module.BasicManager, txEncCfg client.TxEncodingConfig, genBalIterator types.GenesisBalancesIterator, defaultNodeHome string) *cobra.Command {
	ipDefault, _ := server.ExternalIP()
	fsCreateStorageProvider, defaultsDesc := cli.CreateStorageProviderMsgFlagSet(ipDefault)

	cmd := &cobra.Command{
		Use:   "spgentx [amount]",
		Short: "Generate a genesis tx that creates a storage provider",
		Args:  cobra.ExactArgs(1),
		Long: fmt.Sprintf(`Generate a genesis transaction that creates a storage provider,
that is signed by the key in the Keyring referenced by a given name. A node ID and consensus
pubkey may optionally be provided. If they are omitted, they will be retrieved from the priv_validator.json
file. The following default parameters are included:
    %s

Example:
$ %s gentx 10000000000000000000000000BNB --home ./deployment/localup/.local/sp0 \
	--creator=0x76330E9C31D8B91a8247a9bbA2959815D3008417 \
	--operator-address=0x76330E9C31D8B91a8247a9bbA2959815D3008417 \
	--funding-address=0x52C30AA52788ec9C8F36C3774C1F50702BCa59b9 \
	--seal-address=0x419D46b3aA67Dc9075c4FEC4c456fd29697CB897 \
    --bls-pub-key=84a22236c7859ba4e52f43412801b30d3ad1d2f23324a26b001646f30c299357cacb6ccfc017854d9830db8e63639ce6 \
    --bls-proof=987406f0dd2e5f5f3e9e3e447f832dbf73809479ea552fe9b23e06e65651c01a5893e92a797d12078ef9b0c9eb72b8d90faaee65f738e222793d94b80a58cd0f329375ee04e8460f12f4772cf42859d30dd6444ed3350c3335aedf1dbde3bb68 \
	--approval-address=0x68a60866C1e98e277a7389c9Ad90c10cb56debc9 \
	--keyring-backend=test --chain-id=greenfield_9000-121 \
	--moniker=sp0 --details=sp0 --website=http://website --endpoint="http://127.0.0.1:9033" \
	--node tcp://localhost:26752 --node-id sp0 --ip 127.0.0.1 \
	--gas '' --output-document=./deployment/localup/.local/gensptx/gentx-sp0.json
`, defaultsDesc, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			nodeID, _, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}
			if nodeIDString, _ := cmd.Flags().GetString(cli.FlagNodeID); nodeIDString != "" {
				nodeID = nodeIDString
			}

			genDoc, err := tmtypes.GenesisDocFromFile(config.GenesisFile())
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis doc file %s", config.GenesisFile())
			}

			var genesisState map[string]json.RawMessage
			if err = json.Unmarshal(genDoc.AppState, &genesisState); err != nil {
				return errors.Wrap(err, "failed to unmarshal genesis state")
			}

			if err = mbm.ValidateGenesis(cdc, txEncCfg, genesisState); err != nil {
				return errors.Wrap(err, "failed to validate genesis state")
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			// set flags for creating a gensptx
			createSpCfg, err := cli.PrepareConfigForTxCreateStorageProvider(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "error creating configuration to create storage provider msg")
			}

			amount := args[0]
			coins, err := sdk.ParseCoinsNormalized(amount)
			if err != nil {
				return errors.Wrap(err, "failed to parse coins")
			}

			err = genutil.ValidateAccountInGenesis(genesisState, genBalIterator, createSpCfg.FundingAddress, coins, cdc)
			if err != nil {
				return errors.Wrap(err, "failed to validate account in genesis")
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			clientCtx = clientCtx.WithInput(inBuf).WithFromAddress(createSpCfg.FundingAddress)

			createSpCfg.Deposit = amount

			// create a 'create-storage provider' message
			txBldr, msg, err := cli.BuildCreateStorageProviderMsg(createSpCfg, txFactory)
			if err != nil {
				return errors.Wrap(err, "failed to build create-validator message")
			}

			// write the unsigned transaction to the buffer
			w := bytes.NewBuffer([]byte{})
			clientCtx = clientCtx.WithOutput(w)

			if err = txBldr.PrintUnsignedTx(clientCtx, msg); err != nil {
				return errors.Wrap(err, "failed to print unsigned std tx")
			}

			// read the transaction
			stdTx, err := readUnsignedGenTxFile(clientCtx, w)
			if err != nil {
				return errors.Wrap(err, "failed to read unsigned gen tx file")
			}

			// sig verification will skip in the genesis block,
			// but still need a data to be set in Tx to skip the basic validation.
			underlyingTx := authTx.UnWrapTx(stdTx)
			underlyingTx.Signatures = [][]byte{[]byte(fmt.Sprintf("genesis create sp [%s]", createSpCfg.Moniker))}
			stdTx = underlyingTx

			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return errors.Wrap(err, "failed to create output file path")
				}
			}

			if err := writeSignedGenTx(clientCtx, outputDocument, stdTx); err != nil {
				return errors.Wrap(err, "failed to write signed gen tx")
			}

			cmd.PrintErrf("Genesis transaction written to %q\n", outputDocument)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagOutputDocument, "", "Write the genesis transaction JSON document to the given file instead of the default location")

	cmd.Flags().AddFlagSet(fsCreateStorageProvider)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := tmos.EnsureDir(writePath, 0o700); err != nil {
		return "", err
	}

	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(clientCtx client.Context, r io.Reader) (sdk.Tx, error) {
	bz, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	aTx, err := clientCtx.TxConfig.TxJSONDecoder()(bz)
	if err != nil {
		return nil, err
	}

	return aTx, err
}

func writeSignedGenTx(clientCtx client.Context, outputDocument string, tx sdk.Tx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	json, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(outputFile, "%s\n", json)

	return err
}
