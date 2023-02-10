package app

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

type AppConfig struct {
	serverconfig.Config

	CrossChain CrossChainConfig `mapstructure:"cross_chain"`
}

type CrossChainConfig struct {
	SrcChainId uint32 `mapstructure:"src_chain_id"`

	DestChainId uint32 `mapstructure:"dest_chain_id"`
}

var CustomAppTemplate = serverconfig.DefaultConfigTemplate + `
###############################################################################
###                           CrossChain Config                             ###
###############################################################################
[cross_chain]
# chain-id for current chain
src_chain_id = {{ .CrossChain.SrcChainId }}
# chain-id for destination chain(bsc)
dest_chain_id = {{ .CrossChain.DestChainId }}
`

func NewDefaultAppConfig() *AppConfig {
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0stake"

	return &AppConfig{
		Config: *srvCfg,
		CrossChain: CrossChainConfig{
			SrcChainId:  1,
			DestChainId: 2,
		},
	}
}
