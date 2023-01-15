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

bfsd="/Users/owen/go/bin/bfsd --home $HOME/.bfs"
#bfsd="./build/bin/bfsd --home $HOME/.bfs"
#$bfsd keys list

#alice_addr=$($bfsd keys list --output json | jq -r '.[0].address')
sp0_addr=$($bfsd keys list --output json | jq -r '.[2].address')
sp1_addr=$($bfsd keys list --output json | jq -r '.[3].address')
sp2_addr=$($bfsd keys list --output json | jq -r '.[4].address')
sp3_addr=$($bfsd keys list --output json | jq -r '.[5].address')
sp4_addr=$($bfsd keys list --output json | jq -r '.[6].address')
sp5_addr=$($bfsd keys list --output json | jq -r '.[7].address')
sp6_addr=$($bfsd keys list --output json | jq -r '.[8].address')
user_addr=$($bfsd keys list --output json | jq -r '.[9].address')

#$bfsd q bank balances "${alice_addr}"
#$bfsd q payment params

#$bfsd tx payment create-payment-account --from alice -y
#payment_account=$($bfsd q payment get-payment-accounts-by-owner "${alice_addr}" --output json | jq -r '.paymentAccounts[0]')
#$bfsd tx payment deposit "${alice_addr}" 10000000 --from alice -y
#$bfsd tx payment deposit "${payment_account}" 1 --from alice -y
#$bfsd tx payment sponse "$payment_account" 1 --from alice -y
#$bfsd q payment dynamic-balance "$alice_addr"
#$bfsd q payment dynamic-balance "$payment_account"
#sleep 5
#$bfsd q payment dynamic-balance "$alice_addr"
#$bfsd q payment dynamic-balance "$payment_account"
#
## disable payment account refund
#$bfsd tx payment disable-refund "$payment_account" --from alice -y
#refundable=$($bfsd q payment show-payment-account "$payment_account" -o json | jq '.paymentAccount.refundable')
#check_operation "disable refund" "$refundable" "false"

# mock create bucket
bucket_name="test-bucket"
object_name="test-object"
$bfsd tx payment mock-create-bucket "$bucket_name" "" "" "$sp0_addr" 1 --from user -y
$bfsd q payment dynamic-balance "$sp0_addr"
$bfsd q payment dynamic-balance "$user_addr"
$bfsd tx payment mock-put-object "$bucket_name" "$object_name" 30 --from user -y
$bfsd q payment dynamic-balance "$user_addr"
$bfsd tx payment mock-seal-object "$bucket_name" "$object_name" "$sp1_addr,$sp2_addr,$sp3_addr,$sp4_addr,$sp5_addr,$sp6_addr" --from user -y
$bfsd q payment dynamic-balance "$user_addr"
$bfsd q payment dynamic-balance "$sp0_addr"
$bfsd q payment dynamic-balance "$sp1_addr"
$bfsd q payment list-flow
$bfsd tx payment mock-delete-object "$bucket_name" "$object_name" --from user -y
$bfsd q payment dynamic-balance "$user_addr"
$bfsd q payment dynamic-balance "$sp0_addr"
$bfsd q payment dynamic-balance "$sp1_addr"
$bfsd q payment list-flow
$bfsd tx payment mock-update-bucket-read-packet "$bucket_name" 0 --from user -y
$bfsd q payment dynamic-balance "$user_addr"
$bfsd q payment dynamic-balance "$sp0_addr"
