package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	types3 "github.com/bnb-chain/greenfield/types"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/types/resource"
)

func TestGRNBasic(t *testing.T) {
	var grn types3.GRN
	testBucketName := storageutils.GenRandomBucketName()
	testObjectName := storageutils.GenRandomObjectName()
	testAcc := sample.AccAddress()
	testGroupName := storageutils.GenRandomGroupName()

	err := grn.ParseFromString("grn:b::"+testBucketName, false)
	require.NoError(t, err)
	require.Equal(t, grn.MustGetBucketName(), testBucketName)
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_BUCKET)

	err = grn.ParseFromString("grn:o::"+testBucketName+"/"+testObjectName, false)
	require.NoError(t, err)
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_OBJECT)
	bucketName, objectName := grn.MustGetBucketAndObjectName()
	require.Equal(t, bucketName, testBucketName)
	require.Equal(t, objectName, testObjectName)

	err = grn.ParseFromString("grn:g:"+testAcc+":"+testGroupName, false)
	require.NoError(t, err)
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_GROUP)
	groupOwner, groupName := grn.MustGetGroupOwnerAndAccount()
	require.Equal(t, groupOwner.String(), testAcc)
	require.Equal(t, groupName, testGroupName)
}

func TestGRNAbnormal(t *testing.T) {
	var grn types3.GRN
	err := grn.ParseFromString("arn:b::test-bucket", false)
	require.Error(t, err)
	require.ErrorIs(t, err, gnfderrors.ErrInvalidGRN)
	require.True(t, strings.Contains(err.Error(), "regex match error"))

	err = grn.ParseFromString("grn:a::test-bucket", false)
	require.Error(t, err)
	require.ErrorIs(t, err, gnfderrors.ErrInvalidGRN)
	require.True(t, strings.Contains(err.Error(), "regex match error"))

	err = grn.ParseFromString("grn:o:1dd2:test-bucket", false)
	require.Error(t, err)
	require.ErrorIs(t, err, gnfderrors.ErrInvalidGRN)
	require.True(t, strings.Contains(err.Error(), "Not allowed acc non-empty in bucket resource name"))
}

func TestGRNWildcard(t *testing.T) {
	var grn types3.GRN
	testBucketName := storageutils.GenRandomBucketName()

	err := grn.ParseFromString("grn:b::*", false)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "regex match error"))

	err = grn.ParseFromString("grn:b::*", true)
	require.NoError(t, err)
	require.Equal(t, grn.MustGetBucketName(), "*")
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_BUCKET)

	err = grn.ParseFromString("grn:o::"+testBucketName+"/test*", false)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "regex match error"))

	err = grn.ParseFromString("grn:o::"+testBucketName+"/test*", true)
	require.NoError(t, err)
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_OBJECT)
	bucketName, objectName := grn.MustGetBucketAndObjectName()
	require.Equal(t, bucketName, testBucketName)
	require.Equal(t, objectName, "test*")
}

func TestGRNBasicNew(t *testing.T) {
	ownerAcc := sample.RandAccAddress()

	require.Equal(t, "grn:b::testbucket", types3.NewBucketGRN("testbucket").String())
	require.Equal(t, "grn:o::testbucket/testobject", types3.NewObjectGRN("testbucket", "testobject").String())
	groupGRNString := "grn:g:" + ownerAcc.String() + ":testgroup"
	require.Equal(t, groupGRNString, types3.NewGroupGRN(ownerAcc, "testgroup").String())
}

func TestGRNWithSlash(t *testing.T) {
	var grn types3.GRN

	err := grn.ParseFromString("grn:o::"+"testbucket"+"/"+"test/object", false)
	require.NoError(t, err)
	require.Equal(t, grn.ResourceType(), resource.RESOURCE_TYPE_OBJECT)
	bucketName, objectName := grn.MustGetBucketAndObjectName()
	require.Equal(t, bucketName, "testbucket")
	require.Equal(t, objectName, "test/object")
}
