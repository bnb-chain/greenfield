#!/usr/bin/env bash
set -ex

alice_addr=$(bfsd keys list --output json | jq -r '.[0].address')
bfsd q bank balances "${alice_addr}"
bfsd q payment params

bfsd tx payment create-payment-account --from alice -y
payment_account=$(bfsd q payment get-payment-accounts-by-user "${alice_addr}" --output json | jq -r '.paymentAccounts[0]')
bfsd tx payment deposit "${alice_addr}" 1000000 --from alice -y
bfsd tx payment deposit "${payment_account}" 1 --from alice -y
bfsd tx payment sponse "$payment_account" 1 --from alice -y
bfsd q payment dynamic-balance "$alice_addr"
bfsd q payment dynamic-balance "$payment_account"
sleep 5
bfsd q payment dynamic-balance "$alice_addr"
bfsd q payment dynamic-balance "$payment_account"
