package core

import (
	"bufio"
	"fmt"
	"os"
)

type Config struct {
	GrpcAddr string `yaml:"GrpcAddr"`
	ChainId  string `yaml:"ChainId"`
	Mnemonic string `yaml:"Mnemonic"` // test account mnemonic with enough balance
	Denom    string `yaml:"Denom"`
}

func InitConfig() *Config {
	// todo: support qa and testnet config
	return InitE2eConfig()
}

func InitE2eConfig() *Config {
	return &Config{
		GrpcAddr: "localhost:9090",
		ChainId:  "greenfield_9000-121",
		Denom:    "bnb",
		Mnemonic: ParseValidatorMnemonic(0),
	}
}

// ParseValidatorMnemonic read a file and return the last non-empty line
func ParseValidatorMnemonic(i int) string {
	file, err := os.Open(fmt.Sprintf("../../deployment/localup/.local/validator%d/info", i))
	if err != nil {
		panic(err)
	}
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
