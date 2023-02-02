#!/usr/bin/env bash
set -ex

function check_operation() {
	printf "\n=================== Checking $1 ===================\n"
	echo "$2"

	echo "$2" | grep -q $3
	if [ $? -ne 0 ]; then
		echo "Checking $1 Failed"
		exit 1
	fi
}

# dirs
repo_root_dir="$(cd "$(dirname "$0")/.."; pwd)"
gnfd_path="$repo_root_dir/build/bin/gnfd"
e2e_test_dir="$repo_root_dir/build/e2e"
validator_home_dir="$repo_root_dir/deployment/localup/.local/validator0"

# reset integration test dir
rm -rf "$e2e_test_dir"
mkdir -p "$e2e_test_dir/config"
mkdir -p "$e2e_test_dir/keyring-test"
cp "$repo_root_dir/e2e/client.toml" "$e2e_test_dir/config/"
cp "$validator_home_dir/keyring-test/validator0.info" "$e2e_test_dir/keyring-test"

gnfd="$gnfd_path --home $e2e_test_dir"

# keys
$gnfd keys list
validator_addr=$($gnfd keys show validator0 --output json | jq -r ".address")

# payment account test
$gnfd q bank balances "$validator_addr"
$gnfd q payment params
$gnfd tx payment create-payment-account --from validator0 -y
payment_account=$($gnfd q payment get-payment-accounts-by-owner "$validator_addr" --output json | jq -r '.paymentAccounts[0]')
$gnfd tx payment deposit "${payment_account}" 100 --from validator0 -y
# disable payment account refund
##$gnfd tx payment disable-refund "$payment_account" --from alice -y
##refundable=$($gnfd q payment show-payment-account "$payment_account" -o json | jq '.paymentAccount.refundable')
##check_operation "disable refund" "$refundable" "false"
#
## mock create bucket
#bucket_name="test-bucket"
#object_name="test-object"
#$gnfd tx payment mock-create-bucket "$bucket_name" "" "" "$sp0_addr" 1 --from user -y
#$gnfd q payment dynamic-balance "$sp0_addr"
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd tx payment mock-put-object "$bucket_name" "$object_name" 30 --from user -y
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd tx payment mock-seal-object "$bucket_name" "$object_name" "$sp1_addr,$sp2_addr,$sp3_addr,$sp4_addr,$sp5_addr,$sp6_addr" --from user -y
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd q payment dynamic-balance "$sp0_addr"
#$gnfd q payment dynamic-balance "$sp1_addr"
#$gnfd q payment list-flow
#$gnfd q payment list-mock-bucket-meta
## mock-update-bucket-read-packet
## todo: 0 will raise Error: failed to pack and hash typedData primary type: invalid integer value <nil>/<nil> for type int32
#$gnfd tx payment mock-update-bucket-read-packet "$bucket_name" 2 --from user -y
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd q payment dynamic-balance "$sp0_addr"
## mock-set-bucket-payment-account
#$gnfd tx payment mock-set-bucket-payment-account "$bucket_name" "$payment_account" "$payment_account" --from user -y
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd q payment dynamic-balance "$payment_account"
## mock-delete-object
#$gnfd tx payment mock-delete-object "$bucket_name" "$object_name" --from user -y
#$gnfd q payment dynamic-balance "$user_addr"
#$gnfd q payment dynamic-balance "$sp0_addr"
#$gnfd q payment dynamic-balance "$sp1_addr"
#$gnfd q payment list-flow
