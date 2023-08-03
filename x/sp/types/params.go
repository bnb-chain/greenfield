package types

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// SP params default values
const (
	// Default deposit denom
	DefaultDepositDenom = "BNB"
	// DefaultNumOfHistoricalBlocksForMaintenanceRecords The oldest block from current will be kept of for SP maintenance records
	DefaultNumOfHistoricalBlocksForMaintenanceRecords = 864000
	// DefaultMaintenanceDurationQuota is the total allowed time for a SP to be in Maintenance mode within DefaultNumOfHistoricalBlocksForMaintenanceRecords
	DefaultMaintenanceDurationQuota = 21600 // in second
	// DefaultNumOfLockUpBlocksForMaintenance defines blocks difference which Sp update itself to Maintenance mode is allowed
	DefaultNumOfLockUpBlocksForMaintenance = 21600
)

var (
	// DefaultMinDeposit defines the minimum deposit amount for all storage provider
	DefaultMinDeposit = math.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)))
	// DefaultSecondarySpStorePriceRatio is 12%
	DefaultSecondarySpStorePriceRatio = sdk.NewDecFromIntWithPrec(sdk.NewInt(12), 2)
)

var (
	KeyDepositDenom                               = []byte("DepositDenom")
	KeyMinDeposit                                 = []byte("MinDeposit")
	KeySecondarySpStorePriceRatio                 = []byte("SecondarySpStorePriceRatio")
	KeyNumOfHistoricalBlocksForMaintenanceRecords = []byte("NumOfHistoricalBlocksForMaintenanceRecords")
	KeyMaintenanceDurationQuota                   = []byte("MaintenanceDurationQuota")
	KeyNumOfLockUpBlocksForMaintenance            = []byte("NumOfLockUpBlocksForMaintenance")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(depositDenom string, minDeposit math.Int, secondarySpStorePriceRatio sdk.Dec,
	historicalBlocksForMaintenanceRecords, maintenanceDurationQuota, lockUpBlocksForMaintenance int64) Params {
	return Params{
		DepositDenom:               depositDenom,
		MinDeposit:                 minDeposit,
		SecondarySpStorePriceRatio: secondarySpStorePriceRatio,
		NumOfHistoricalBlocksForMaintenanceRecords: historicalBlocksForMaintenanceRecords,
		MaintenanceDurationQuota:                   maintenanceDurationQuota,
		NumOfLockupBlocksForMaintenance:            lockUpBlocksForMaintenance,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDepositDenom, DefaultMinDeposit, DefaultSecondarySpStorePriceRatio,
		DefaultNumOfHistoricalBlocksForMaintenanceRecords, DefaultMaintenanceDurationQuota, DefaultNumOfLockUpBlocksForMaintenance)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositDenom, &p.DepositDenom, validateDepositDenom),
		paramtypes.NewParamSetPair(KeyMinDeposit, &p.MinDeposit, validateMinDeposit),
		paramtypes.NewParamSetPair(KeySecondarySpStorePriceRatio, &p.SecondarySpStorePriceRatio, validateSecondarySpStorePriceRatio),
		paramtypes.NewParamSetPair(KeyNumOfHistoricalBlocksForMaintenanceRecords, &p.NumOfHistoricalBlocksForMaintenanceRecords, validateHistoricalBlocksForMaintenanceRecords),
		paramtypes.NewParamSetPair(KeyMaintenanceDurationQuota, &p.MaintenanceDurationQuota, validateMaintenanceDurationQuota),
		paramtypes.NewParamSetPair(KeyNumOfLockUpBlocksForMaintenance, &p.NumOfLockupBlocksForMaintenance, validateLockUpBlocksForMaintenance),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDepositDenom(p.DepositDenom); err != nil {
		return err
	}

	if err := validateMinDeposit(p.MinDeposit); err != nil {
		return err
	}

	if err := validateSecondarySpStorePriceRatio(p.SecondarySpStorePriceRatio); err != nil {
		return err
	}
	if err := validateHistoricalBlocksForMaintenanceRecords(p.NumOfHistoricalBlocksForMaintenanceRecords); err != nil {
		return err
	}
	if err := validateMaintenanceDurationQuota(p.MaintenanceDurationQuota); err != nil {
		return err
	}
	if err := validateLockUpBlocksForMaintenance(p.NumOfLockupBlocksForMaintenance); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateDepositDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("deposit denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateMinDeposit(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("minimum deposit amount cannot be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum deposit amount cannot be lower than 0")
	}

	return nil
}

func validateSecondarySpStorePriceRatio(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() || !v.IsPositive() || v.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid secondary sp store price ratio")
	}
	return nil
}

func validateHistoricalBlocksForMaintenanceRecords(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return errors.New("HistoricalBlocksForMaintenanceRecords cannot be zero")
	}
	return nil
}

func validateMaintenanceDurationQuota(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return errors.New("MaintenanceDurationQuota cannot be zero")
	}
	return nil
}
func validateLockUpBlocksForMaintenance(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return errors.New("LockUpBlocksForMaintenance cannot be zero")
	}
	return nil
}
