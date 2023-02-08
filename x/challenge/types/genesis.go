package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		OngoingChallenges: []Challenge{},
		RecentSlashes:     []Slash{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in ongoingChallenge
	ongoingChallengeIndexMap := make(map[string]struct{})

	for _, elem := range gs.OngoingChallenges {
		index := string(OngoingChallengeKey(elem.Id))
		if _, ok := ongoingChallengeIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for ongoingChallenge")
		}
		ongoingChallengeIndexMap[index] = struct{}{}
	}
	// Check for duplicated ID in recentSlash
	recentSlashIdMap := make(map[uint64]bool)
	recentSlashCount := gs.GetRecentSlashCount()
	for _, elem := range gs.RecentSlashes {
		if _, ok := recentSlashIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for recentSlash")
		}
		if elem.Id >= recentSlashCount {
			return fmt.Errorf("recentSlash id should be lower or equal than the last id")
		}
		recentSlashIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
