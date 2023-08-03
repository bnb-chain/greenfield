package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type SPMnemonics struct {
	OperatorMnemonic string `yaml:"OperatorMnemonic"` // operator account mnemonic with enough balance
	SealMnemonic     string `yaml:"SealMnemonic"`     // seal account mnemonic with enough balance
	FundingMnemonic  string `yaml:"FundingMnemonic"`  // funding account mnemonic with enough balance
	ApprovalMnemonic string `yaml:"ApprovalMnemonic"` // approval account mnemonic with enough balance
	GcMnemonic       string `yaml:"GcMnemonic"`       // gc account mnemonic with enough balance
	TestMnemonic     string `yaml:"TestMnemonic"`     // test account mnemonic with enough balance
}

type Config struct {
	GrpcAddr             string        `yaml:"GrpcAddr"`
	TendermintAddr       string        `yaml:"TendermintAddr"`
	ChainId              string        `yaml:"ChainId"`
	ValidatorMnemonic    string        `yaml:"Mnemonic"`           // validator operator account mnemonic with enough balance
	ValidatorBlsMnemonic string        `yaml:"BLSMnemonic"`        // validator's mnemonic for bls key
	RelayerMnemonic      string        `yaml:"RelayerMnemonic"`    // relayer mnemonic
	ChallengerMnemonic   string        `yaml:"ChallengerMnemonic"` // challenger mnemonic
	SPMnemonics          []SPMnemonics `yaml:"SPMnemonics"`
	SPBLSMnemonic        []string      `yaml:"SPBLSMnemonic"`
	Denom                string        `yaml:"Denom"`
	ValidatorHomeDir     string        `yaml:"ValidatorHomeDir"`
	ValidatorTmRPCAddr   string        `yaml:"ValidatorTmRPCAddr"`
}

func InitConfig() *Config {
	// TODO: support qa and testnet config
	return InitE2eConfig()
}

func InitE2eConfig() *Config {
	config := &Config{
		GrpcAddr:             "localhost:9090",
		TendermintAddr:       "http://127.0.0.1:26750",
		ChainId:              "greenfield_9000-121",
		Denom:                "BNB",
		ValidatorMnemonic:    ParseValidatorMnemonic(0),
		ValidatorBlsMnemonic: ParseValidatorBlsMnemonic(0),
		RelayerMnemonic:      ParseRelayerMnemonic(0),
		ChallengerMnemonic:   ParseChallengerMnemonic(0),
		ValidatorHomeDir:     ParseValidatorHomeDir(0),
		ValidatorTmRPCAddr:   ParseValidatorTmRPCAddrDir(0),
	}
	for i := 0; i < 7; i++ {
		config.SPMnemonics = append(config.SPMnemonics, ParseSPMnemonics(i))
		config.SPBLSMnemonic = append(config.SPBLSMnemonic, ParseSPBLSMnemonics(i))
	}
	return config
}

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/validator%d/info", i))
}

// ParseValidatorBlsMnemonic read the validator mnemonic of bls from file
func ParseValidatorBlsMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/validator%d/bls_info", i))
}

// ParseRelayerMnemonic read the relayer mnemonic from file
func ParseRelayerMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/relayer%d/relayer_info", i))
}

// ParseChallengerMnemonic read the challenger mnemonic from file
func ParseChallengerMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/challenger%d/challenger_info", i))
}

// ParseSPMnemonics read the sp mnemonics from file
func ParseSPMnemonics(i int) SPMnemonics {
	return SPMnemonics{
		OperatorMnemonic: ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/info", i)),
		SealMnemonic:     ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/seal_info", i)),
		FundingMnemonic:  ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/fund_info", i)),
		ApprovalMnemonic: ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/approval_info", i)),
		GcMnemonic:       ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/gc_info", i)),
		TestMnemonic:     ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/test_info", i)),
	}
}

// ParseSPBLSMnemonics read the sp bls mnemonics from file
func ParseSPBLSMnemonics(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../deployment/localup/.local/sp%d/bls_info", i))
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

// ParseValidatorHomeDir returns the home dir of the validator
func ParseValidatorHomeDir(i int) string {
	return fmt.Sprintf("../../deployment/localup/.local/validator%d", i)
}

// ParseValidatorTmRPCAddrDir returns the home dir of the validator
func ParseValidatorTmRPCAddrDir(i int) string {
	return fmt.Sprintf("tcp://0.0.0.0:%d", 26750+i)
}
