package cli_test

import (
	"fmt"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/bfs/testutil/network"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment/client/cli"
    "github.com/bnb-chain/bfs/x/payment/types"
)

func networkWithBnbPriceObjects(t *testing.T) (*network.Network, types.BnbPrice) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
    require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	bnbPrice := &types.BnbPrice{}
	nullify.Fill(&bnbPrice)
	state.BnbPrice = bnbPrice
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), *state.BnbPrice
}

func TestShowBnbPrice(t *testing.T) {
	net, obj := networkWithBnbPriceObjects(t)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  types.BnbPrice
	}{
		{
			desc: "get",
			args: common,
			obj:  obj,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var args []string
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowBnbPrice(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetBnbPriceResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.BnbPrice)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.BnbPrice),
				)
			}
		})
	}
}

