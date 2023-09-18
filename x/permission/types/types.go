package types

import (
	"regexp"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gnfd "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	"github.com/bnb-chain/greenfield/types/resource"
)

type (
	Int  = math.Int
	Uint = math.Uint
)

type VerifyOptions struct {
	Resource   string
	WantedSize *uint64
}

var (
	BucketAllowedActions = map[ActionType]bool{
		ACTION_UPDATE_BUCKET_INFO: true,
		ACTION_DELETE_BUCKET:      true,

		ACTION_CREATE_OBJECT:  true,
		ACTION_DELETE_OBJECT:  true,
		ACTION_GET_OBJECT:     true,
		ACTION_COPY_OBJECT:    true,
		ACTION_EXECUTE_OBJECT: true,
		ACTION_LIST_OBJECT:    true,

		ACTION_TYPE_ALL: true,
	}
	BucketAllowedActionsAfterXxxxx = map[ActionType]bool{
		ACTION_UPDATE_BUCKET_INFO: true,
		ACTION_DELETE_BUCKET:      true,

		ACTION_CREATE_OBJECT:      true,
		ACTION_DELETE_OBJECT:      true,
		ACTION_GET_OBJECT:         true,
		ACTION_COPY_OBJECT:        true,
		ACTION_EXECUTE_OBJECT:     true,
		ACTION_LIST_OBJECT:        true,
		ACTION_UPDATE_OBJECT_INFO: true,

		ACTION_TYPE_ALL: true,
	}
	ObjectAllowedActions = map[ActionType]bool{
		ACTION_UPDATE_OBJECT_INFO: true,
		ACTION_CREATE_OBJECT:      true,
		ACTION_DELETE_OBJECT:      true,
		ACTION_GET_OBJECT:         true,
		ACTION_COPY_OBJECT:        true,
		ACTION_EXECUTE_OBJECT:     true,
		ACTION_LIST_OBJECT:        true,

		ACTION_TYPE_ALL: true,
	}
	GroupAllowedActions = map[ActionType]bool{
		ACTION_UPDATE_GROUP_MEMBER: true,
		ACTION_UPDATE_GROUP_EXTRA:  true,
		ACTION_DELETE_GROUP:        true,

		ACTION_TYPE_ALL: true,
	}
)

// Eval is used to evaluate the execution results of permission policies.
// First, each policy has an expiration time. If it has expired, EFFECT_UNSPECIFIED will be returned, indicating that it cannot be evaluated and further verification is required.
// Next, each statement in the policy needs to be checked, which includes verifying:
// 1. Whether the statement has expired,
// 2. Whether the limit size has been exceeded,
// 3. Whether the resource in the statement matches the input resource name,
// 4. Whether the action in the statement matches the input action.
// Finally, in the verification process, based on the effect check
// 1. if there is an explicit Deny, return EFFECT_DENY;
// 2. if there is an explicit Allowed, record the flag and continue execution;
// 3. after all statements have been checked, if the flag is true, return EFFECT_ALLOW; otherwise return EFFECT_UNSPECIFIED.
func (p *Policy) Eval(action ActionType, blockTime time.Time, opts *VerifyOptions) (Effect, *Policy) {
	// 1. the policy is expired, need delete
	if p.ExpirationTime != nil && p.ExpirationTime.Before(blockTime) {
		// Notice: We do not actively delete policies that expire for users.
		return EFFECT_UNSPECIFIED, nil
	}
	allowed := false
	updated := false
	// 2. check all the statements
	for i, s := range p.Statements {
		if s.ExpirationTime != nil && s.ExpirationTime.Before(blockTime) {
			continue
		}
		e, updatedStatement := s.Eval(action, opts)
		// statement need to be updated
		if updatedStatement != nil {
			updated = true
			p.Statements[i] = updatedStatement
		}
		if e == EFFECT_DENY {
			return EFFECT_DENY, nil
		} else if e == EFFECT_ALLOW {
			allowed = true
		}
	}
	if allowed {
		if updated {
			return EFFECT_ALLOW, p
		} else {
			return EFFECT_ALLOW, nil
		}
	}
	return EFFECT_UNSPECIFIED, nil
}

func NewMemberStatement() *Statement {
	return &Statement{
		Effect:    EFFECT_ALLOW,
		Resources: nil,
		Actions:   nil,
	}

}
func (s *Statement) Eval(action ActionType, opts *VerifyOptions) (Effect, *Statement) {
	// If 'resource' is not nil, it implies that the user intends to access a sub-resource, which would
	// be specified in 's.Resources'. Therefore, if the sub-resource in the statement is nil, we will ignore this statement.
	if opts != nil && opts.Resource != "" && s.Resources == nil {
		return EFFECT_UNSPECIFIED, nil
	}
	// If 'resource' is not nil, and 's.Resource' is also not nil, it indicates that we should verify whether
	// the resource that the user intends to access matches any items in 's.Resource'
	if opts != nil && opts.Resource != "" && s.Resources != nil {
		isMatch := false
		for _, res := range s.Resources {
			reg := regexp.MustCompile(res)
			if reg == nil {
				continue
			}
			matchRes := reg.MatchString(opts.Resource)
			if matchRes {
				isMatch = matchRes
				break
			}
		}
		if !isMatch {
			return EFFECT_UNSPECIFIED, nil
		}
	}

	for _, act := range s.Actions {
		if act == action || act == ACTION_TYPE_ALL {
			// Action matched, if effect is deny, then return deny
			if s.Effect == EFFECT_DENY {
				return EFFECT_DENY, nil
			}
			// There is special handling for ACTION_CREATE_OBJECT.
			// userA grant CreateObject permission to userB, but only allows him to create a limit size of object.
			// If exceeded, rejected
			if action == ACTION_CREATE_OBJECT && s.LimitSize != nil && opts != nil && opts.WantedSize != nil {
				if s.LimitSize.GetValue() >= *opts.WantedSize {
					s.LimitSize = &common.UInt64Value{Value: s.LimitSize.GetValue() - *opts.WantedSize}
					return EFFECT_ALLOW, s
				} else {
					return EFFECT_DENY, nil
				}
			}
			return s.Effect, nil
		}
	}

	return EFFECT_UNSPECIFIED, nil
}

func (s *Statement) ValidateBasic(resType resource.ResourceType) error {
	if s.Effect == EFFECT_UNSPECIFIED {
		return ErrInvalidStatement.Wrap("Please specify the Effect explicitly. Not allowed set EFFECT_UNSPECIFIED")
	}
	switch resType {
	case resource.RESOURCE_TYPE_UNSPECIFIED:
		return ErrInvalidStatement.Wrap("Please specify the ResourceType explicitly. Not allowed set RESOURCE_TYPE_UNSPECIFIED")
	case resource.RESOURCE_TYPE_BUCKET:
		//containsCreateObject := false
		//for _, a := range s.Actions {
		//	if !BucketAllowedActions[a] {
		//		return ErrInvalidStatement.Wrapf("%s not allowed to be used on bucket.", a.String())
		//	}
		//	if a == ACTION_CREATE_OBJECT {
		//		containsCreateObject = true
		//	}
		//}
		//if !containsCreateObject && s.LimitSize != nil {
		//	return ErrInvalidStatement.Wrap("The LimitSize option can only be used with CreateObject actions at the bucket level. .")
		//}
		for _, r := range s.Resources {
			var grn gnfd.GRN
			err := grn.ParseFromString(r, true)
			if err != nil {
				return ErrInvalidStatement.Wrapf("GRN parse from string failed, err: %s", err)
			}
		}
	case resource.RESOURCE_TYPE_OBJECT:
		for _, a := range s.Actions {
			if !ObjectAllowedActions[a] {
				return ErrInvalidStatement.Wrapf("%s not allowed to be used on object.", a.String())
			}
		}
		if s.LimitSize != nil {
			return ErrInvalidStatement.Wrap("The LimitSize option can only be used with CreateObject actions at the bucket level. ")
		}
	case resource.RESOURCE_TYPE_GROUP:
		for _, a := range s.Actions {
			if !GroupAllowedActions[a] {
				return ErrInvalidStatement.Wrapf("%s not allowed to be used on group.", a.String())
			}
		}
		if s.LimitSize != nil {
			return ErrInvalidStatement.Wrap("The LimitSize option can only be used with CreateObject actions at the bucket level. ")
		}
	default:
		return ErrInvalidStatement.Wrap("unknown resource type.")
	}
	return nil
}

func (s *Statement) ValidateAfterNagqu(resType resource.ResourceType) error {
	if s.Effect == EFFECT_UNSPECIFIED {
		return ErrInvalidStatement.Wrap("Please specify the Effect explicitly. Not allowed set EFFECT_UNSPECIFIED")
	}
	switch resType {
	case resource.RESOURCE_TYPE_UNSPECIFIED:
		return ErrInvalidStatement.Wrap("Please specify the ResourceType explicitly. Not allowed set RESOURCE_TYPE_UNSPECIFIED")
	case resource.RESOURCE_TYPE_BUCKET:
		for _, r := range s.Resources {
			_, err := regexp.Compile(r)
			if err != nil {
				return ErrInvalidStatement.Wrapf("The Resources regexp compile failed, err: %s", err)
			}
		}
	case resource.RESOURCE_TYPE_OBJECT:
		if s.Resources != nil {
			return ErrInvalidStatement.Wrap("The Resources option can only be used at the bucket level. ")
		}
	case resource.RESOURCE_TYPE_GROUP:
		if s.Resources != nil {
			return ErrInvalidStatement.Wrap("The Resources option can only be used at the bucket level. ")
		}
	default:
		return ErrInvalidStatement.Wrap("unknown resource type.")
	}
	return nil
}

func (s *Statement) ValidateRuntime(ctx sdk.Context, resType resource.ResourceType) error {
	var bucketAllowedActions map[ActionType]bool
	if ctx.IsUpgraded(upgradetypes.Xxxxx) {
		bucketAllowedActions = BucketAllowedActionsAfterXxxxx
	} else {
		bucketAllowedActions = BucketAllowedActions
	}
	if resType == resource.RESOURCE_TYPE_BUCKET {
		containsCreateObject := false
		for _, a := range s.Actions {
			if !bucketAllowedActions[a] {
				return ErrInvalidStatement.Wrapf("%s not allowed to be used on bucket.", a.String())
			}
			if a == ACTION_CREATE_OBJECT {
				containsCreateObject = true
			}
		}
		if !containsCreateObject && s.LimitSize != nil {
			return ErrInvalidStatement.Wrap("The LimitSize option can only be used with CreateObject actions at the bucket level. .")
		}
	}
	return nil
}
