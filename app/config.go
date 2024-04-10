package app

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

type AppConfig struct {
	serverconfig.Config

	CrossChain CrossChainConfig `mapstructure:"cross-chain"`

	PaymentCheck PaymentCheckConfig `mapstructure:"payment-check"`
}

type CrossChainConfig struct {
	SrcChainId uint32 `mapstructure:"src-chain-id"`

	DestBscChainId uint32 `mapstructure:"dest-bsc-chain-id"`

	DestOpChainId uint32 `mapstructure:"dest-op-chain-id"`
}

type PaymentCheckConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Interval uint32 `mapstructure:"interval"`
}

var CustomAppTemplate = serverconfig.DefaultConfigTemplate + `
###############################################################################
###                           CrossChain Config                             ###
###############################################################################
[cross-chain]
# chain-id for current chain
src-chain-id = {{ .CrossChain.SrcChainId }}
# chain-id for bsc destination chain
dest-bsc-chain-id = {{ .CrossChain.DestBscChainId }}
# chain-id for op bnb destination chain
dest-op-chain-id = {{ .CrossChain.DestOpChainId }}

###############################################################################
###                           PaymentCheck Config                           ###
###############################################################################
[payment-check]
# enabled - the flag to enable/disable payment check
enabled = {{ .PaymentCheck.Enabled }}
# interval - the block interval run check payment
interval = {{ .PaymentCheck.Interval }}
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
	srvCfg.MinGasPrices = "5000000000BNB" // 5gei

	return &AppConfig{
		Config: *srvCfg,
		CrossChain: CrossChainConfig{
			SrcChainId:     1,
			DestBscChainId: 2,
			DestOpChainId:  3,
		},
		PaymentCheck: PaymentCheckConfig{
			Enabled:  false,
			Interval: 100,
		},
	}
}
