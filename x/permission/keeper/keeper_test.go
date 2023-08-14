package keeper_test

import (
	"math/rand"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/permission/types"
)

func (s *TestSuite) TestPruneAccountPolicies() {
	now := s.ctx.BlockTime()
	oneDayAfter := now.AddDate(0, 0, 1)

	resourceIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}
	policyIds := make([]math.Uint, 3)

	// policy without expiry
	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceIds[0],
		Statements:     nil,
		ExpirationTime: nil,
	}
	policyId, err := s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[0] = policyId

	policy.ResourceId = resourceIds[2]
	policyId, err = s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[2] = policyId

	// policy with expiry
	policy.ResourceId = resourceIds[1]
	policy.ExpirationTime = &oneDayAfter
	policyId, err = s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[1] = policyId

	testCases := []struct {
		name       string
		ctx        sdk.Context
		resourceId math.Uint
		policyId   math.Uint
		found      bool
		preRun     func()
		postRun    func()
	}{
		{
			name:       "no expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter),
			resourceId: resourceIds[0],
			policyId:   policyIds[0],
			found:      true,
		},
		{
			name:       "expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter),
			resourceId: resourceIds[1],
			policyId:   policyIds[1],
			found:      true,
		},
		{
			name:       "expiry and prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[1],
			policyId:   policyIds[1],
		},
		{
			name:       "update from no expiry to expiry and prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[0],
			policyId:   policyIds[0],
			preRun: func() {
				oldPolicy, found := s.permissionKeeper.GetPolicyByID(s.ctx, policyIds[0])
				s.True(found)
				oldPolicy.ExpirationTime = &oneDayAfter
				newId, err := s.permissionKeeper.PutPolicy(s.ctx, oldPolicy)
				s.NoError(err)
				s.Equal(policyIds[0], newId)
			},
		},
		{
			name:       "update from expiry to no expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[2],
			policyId:   policyIds[2],
			found:      true,
			preRun: func() {
				oldPolicy, found := s.permissionKeeper.GetPolicyByID(s.ctx, policyIds[2])
				s.True(found)
				oldPolicy.ExpirationTime = nil
				newId, err := s.permissionKeeper.PutPolicy(s.ctx, oldPolicy)
				s.NoError(err)
				s.Equal(policyIds[2], newId)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, found := s.permissionKeeper.GetPolicyByID(tc.ctx, tc.policyId)
			s.True(found)
			s.permissionKeeper.RemoveExpiredPolicies(tc.ctx)
			_, found = s.permissionKeeper.GetPolicyByID(tc.ctx, tc.policyId)
			s.Equal(tc.found, found)
			if tc.postRun != nil {
				tc.postRun()
			}
		})
	}
}

func (s *TestSuite) TestPruneGroupPolicies() {
	now := s.ctx.BlockTime()
	oneDayAfter := now.AddDate(0, 0, 1)

	resourceIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}
	policyIds := make([]math.Uint, 3)

	// member without expiry
	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_GROUP,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceIds[0],
		Statements:     nil,
		ExpirationTime: nil,
	}
	policyId, err := s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[0] = policyId

	policy.ResourceId = resourceIds[2]
	policyId, err = s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[2] = policyId

	// member with expiry
	policy.ResourceId = resourceIds[1]
	policy.ExpirationTime = &oneDayAfter
	policyId, err = s.permissionKeeper.PutPolicy(s.ctx, &policy)
	s.NoError(err)
	policyIds[1] = policyId

	testCases := []struct {
		name       string
		ctx        sdk.Context
		resourceId math.Uint
		policyId   math.Uint
		found      bool
		preRun     func()
		postRun    func()
	}{
		{
			name:       "no expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter),
			resourceId: resourceIds[0],
			policyId:   policyIds[0],
			found:      true,
		},
		{
			name:       "expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter),
			resourceId: resourceIds[1],
			policyId:   policyIds[1],
			found:      true,
		},
		{
			name:       "expiry and prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[1],
			policyId:   policyIds[1],
		},
		{
			name:       "update from no expiry to expiry and prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[0],
			policyId:   policyIds[0],
			preRun: func() {
				oldPolicy, found := s.permissionKeeper.GetPolicyByID(s.ctx, policyIds[0])
				s.True(found)
				oldPolicy.ExpirationTime = &oneDayAfter
				newId, err := s.permissionKeeper.PutPolicy(s.ctx, oldPolicy)
				s.NoError(err)
				s.Equal(policyIds[0], newId)
			},
		},
		{
			name:       "update from expiry to no expiry and no prune",
			ctx:        s.ctx.WithBlockTime(oneDayAfter.Add(time.Second)),
			resourceId: resourceIds[2],
			policyId:   policyIds[2],
			found:      true,
			preRun: func() {
				oldPolicy, found := s.permissionKeeper.GetPolicyByID(s.ctx, policyIds[2])
				s.True(found)
				oldPolicy.ExpirationTime = nil
				newId, err := s.permissionKeeper.PutPolicy(s.ctx, oldPolicy)
				s.NoError(err)
				s.Equal(policyIds[2], newId)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, found := s.permissionKeeper.GetPolicyByID(tc.ctx, tc.policyId)
			s.True(found)
			s.permissionKeeper.RemoveExpiredPolicies(tc.ctx)
			_, found = s.permissionKeeper.GetPolicyByID(tc.ctx, tc.policyId)
			s.Equal(tc.found, found)
			if tc.postRun != nil {
				tc.postRun()
			}
		})
	}
}
