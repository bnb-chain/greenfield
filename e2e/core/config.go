package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type SPMnemonics struct {
	OperatorMnemonic string `yaml:"Mnemonic"` // operator account mnemonic with enough balance
	SealMnemonic     string `yaml:"Mnemonic"` // seal account mnemonic with enough balance
	FundingMnemonic  string `yaml:"Mnemonic"` // funding account mnemonic with enough balance
	ApprovalMnemonic string `yaml:"Mnemonic"` // approval account mnemonic with enough balance
}

type Config struct {
	GrpcAddr          string      `yaml:"GrpcAddr"`
	ChainId           string      `yaml:"ChainId"`
	ValidatorMnemonic string      `yaml:"Mnemonic"` // validator operator account mnemonic with enough balance
	SPMnemonics       SPMnemonics `yaml:"SPMnemonics"`
	Denom             string      `yaml:"Denom"`
}

func InitConfig() *Config {
	// TODO: support qa and testnet config
	return InitE2eConfig()
}

func InitE2eConfig() *Config {
	return &Config{
		GrpcAddr:          "localhost:9090",
		ChainId:           "greenfield_9000-121",
		Denom:             "BNB",
		ValidatorMnemonic: ParseValidatorMnemonic(0),
		SPMnemonics:       ParseSPMnemonics(0),
	}
}

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/validator%d/info", i))
}

// ParseSPMnemonics read the sp mnemonics from file
func ParseSPMnemonics(i int) SPMnemonics {
	return SPMnemonics{
		OperatorMnemonic: ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/info", i)),
		SealMnemonic:     ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/seal_info", i)),
		FundingMnemonic:  ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/fund_info", i)),
		ApprovalMnemonic: ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/approval_info", i)),
	}

}

func ParseMnemonicFromFile(fileName string) string {
	fileName = filepath.Clean(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	// #nosec
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		if scanner.Text() != "" {
			line = scanner.Text()
		}
	}
	return line
}
