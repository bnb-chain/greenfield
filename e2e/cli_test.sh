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
denom=bnb

# keys
$gnfd keys list
validator_addr=$($gnfd keys show validator0 --output json | jq -r ".address")
$gnfd keys add user
user_addr=$($gnfd keys show user --output json | jq -r ".address")
$gnfd keys add sp0
$gnfd keys add sp1
$gnfd keys add sp2
$gnfd keys add sp3
$gnfd keys add sp4
$gnfd keys add sp5
$gnfd keys add sp6
sp0_addr=$($gnfd keys show sp0 --output json | jq -r ".address")
sp1_addr=$($gnfd keys show sp1 --output json | jq -r ".address")
sp2_addr=$($gnfd keys show sp2 --output json | jq -r ".address")
sp3_addr=$($gnfd keys show sp3 --output json | jq -r ".address")
sp4_addr=$($gnfd keys show sp4 --output json | jq -r ".address")
sp5_addr=$($gnfd keys show sp5 --output json | jq -r ".address")
sp6_addr=$($gnfd keys show sp6 --output json | jq -r ".address")

# balance
$gnfd tx bank send validator0 "$user_addr" "100000$denom" -y
$gnfd q bank balances "$user_addr"

# ----- payment account test -----
# $gnfd q payment params
# create payment account
# $gnfd tx payment create-payment-account --from user -y
# payment_account=$($gnfd q payment get-payment-accounts-by-owner "$user_addr" --output json | jq -r '.paymentAccounts[0]')
# disable payment account refund
# refundable=$($gnfd q payment show-payment-account "$payment_account" -o json | jq '.paymentAccount.refundable')
# check_operation "disable refund" "$refundable" "true"
# $gnfd tx payment disable-refund "$payment_account" --from user -y
# refundable=$($gnfd q payment show-payment-account "$payment_account" -o json | jq '.paymentAccount.refundable')
# check_operation "disable refund" "$refundable" "false"
# deposit
# $gnfd tx payment deposit "${payment_account}" 100 --from user -y
