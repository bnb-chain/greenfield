package types

import (
	"regexp"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gnfd "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/resource"
)

type (
	Int  = math.Int
	Uint = math.Uint
)

func NewDefaultPolicyForGroupMember(groupID math.Uint, member sdk.AccAddress) *Policy {
	return &Policy{
		ResourceType:    resource.RESOURCE_TYPE_GROUP,
		ResourceId:      groupID,
		MemberStatement: NewMemberStatement(),
	}
}

func (p *Policy) Eval(action ActionType, resource *string) Effect {
	allowed := false
	for _, s := range p.Statements {
		e := s.Eval(action, resource)
		if e == EFFECT_DENY {
			return e
		} else if e == EFFECT_ALLOW {
			allowed = true
		}
	}
	if allowed {
		return EFFECT_ALLOW
	}
	return EFFECT_PASS
}

func (p *Policy) GetGroupMemberStatement() (*Statement, bool) {
	for _, s := range p.Statements {
		for _, act := range s.Actions {
			if act == ACTION_GROUP_MEMBER {
				return s, true
			}
		}
	}
	return nil, false
}

func NewMemberStatement() *Statement {
	return &Statement{
		Effect:    EFFECT_ALLOW,
		Resources: nil,
		Actions:   []ActionType{ACTION_GROUP_MEMBER},
	}

}
func (s *Statement) Eval(action ActionType, resource *string) Effect {
	if resource != nil && s.Resources == nil {
		return EFFECT_PASS
	}

	if s.Resources != nil && resource != nil {
		isMatch := false
		for _, res := range s.Resources {
			reg := regexp.MustCompile(res)
			if reg == nil {
				continue
			}
			matchRes := reg.MatchString(*resource)
			if matchRes {
				isMatch = matchRes
				break
			}
		}
		if !isMatch {
			return EFFECT_PASS
		}
	}

	for _, act := range s.Actions {
		if act == action || act == ACTION_TYPE_ALL {
			return s.Effect
		}
	}

	return EFFECT_PASS
}

func (s *Statement) ValidateBasic() error {
	for _, r := range s.Resources {
		var grn gnfd.GRN
		err := grn.ParseFromString(r, true)
		if err != nil {
			return err
		}
	}
	return nil
}
