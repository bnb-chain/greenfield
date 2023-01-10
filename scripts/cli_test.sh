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
$bfsd keys list

alice_addr=$($bfsd keys list --output json | jq -r '.[0].address')
bob_addr=$($bfsd keys list --output json | jq -r '.[1].address')
$bfsd q bank balances "${alice_addr}"
$bfsd q payment params

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
$bfsd tx payment mock-create-bucket bucket1 '' '' "$bob_addr" 1 --from alice -y
$bfsd q payment dynamic-balance "$bob_addr"
