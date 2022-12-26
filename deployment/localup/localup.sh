#!/usr/bin/env bash
basedir=$(cd `dirname $0`; pwd)
workspace=${basedir}
source ${workspace}/.env
bin_name=bfsd
bin=${workspace}/../../build/bin/${bin_name}
address_port_start=28750
p2p_port_start=27750
grpc_port_start=9090
grpc_web_port_start=9190
rpc_port_start=26750


function joinByString() {
  local separator="$1"
  shift
  local first="$1"
  shift
  printf "%s" "$first" "${@/#/$separator}"
}

function init() {
    size=$1
    rm -rf ${workspace}/.local
    mkdir -p ${workspace}/.local
    for ((i=0;i<${size};i++));do
        mkdir -p ${workspace}/.local/validator${i}
        mkdir -p ${workspace}/.local/relayer${i}

        # init chain
        ${bin} init validator${i} --chain-id ${CHAIN_ID} --staking-bond-denom ${STAKING_BOND_DENOM} --home ${workspace}/.local/validator${i}

        # create genesis accounts
        ${bin} keys add validator${i} --keyring-backend test --home ${workspace}/.local/validator${i} > ${workspace}/.local/validator${i}/info
        ${bin} keys add relayer${i} --keyring-backend test --home ${workspace}/.local/relayer${i} --algo eth_bls > ${workspace}/.local/relayer${i}/info
    done
}

function generate_genesis() {
    size=$1
    declare -a validator_addrs=()
    for ((i=0;i<${size};i++));do
        # export validator addresses
        validator_addrs+=("$(${bin} keys show validator${i} -a --keyring-backend test --home ${workspace}/.local/validator${i})")
    done

    mkdir -p ${workspace}/.local/gentx
    for ((i=0;i<${size};i++));do
        for validator_addr in "${validator_addrs[@]}";do
            # init genesis account in genesis state
            ${bin} add-genesis-account $validator_addr ${GENESIS_ACCOUNT_BALANCE}${STAKING_BOND_DENOM} --home ${workspace}/.local/validator${i}
        done

        rm -rf ${workspace}/.local/validator${i}/config/gentx/

        validatorAddr=${validator_addrs[$i]}
        relayerAddr="$(${bin} keys show relayer${i} -a --keyring-backend test --home ${workspace}/.local/relayer${i})"
        relayerBLSKey="$(${bin} keys show relayer${i} --keyring-backend test --home ${workspace}/.local/relayer${i} --output json | jq -r .pubkey_hex)"
        
        # create bond validator tx
        ${bin} gentx validator${i} ${STAKING_BOND_AMOUNT}${STAKING_BOND_DENOM} $validatorAddr $relayerAddr $relayerBLSKey \
            --home ${workspace}/.local/validator${i} \
            --keyring-backend=test \
            --chain-id=${CHAIN_ID} \
            --moniker="validator${i}" \
            --commission-max-change-rate=${COMMISSION_MAX_CHANGE_RATE} \
            --commission-max-rate=${COMMISSION_MAX_RATE} \
            --commission-rate=${COMMISSION_RATE} \
            --details="validator${i}" \
            --website="http://website" \
            --node tcp://localhost:$((${p2p_port_start}+${i})) \
            --node-id "validator${i}" \
            --ip 127.0.0.1
        cp ${workspace}/.local/validator${i}/config/gentx/gentx-validator${i}.json ${workspace}/.local/gentx/
    done

    node_ids=""
    # bond validator tx in genesis state
    for ((i=0;i<${size};i++));do
        cp ${workspace}/.local/gentx/* ${workspace}/.local/validator${i}/config/gentx/
        ${bin} collect-gentxs --home ${workspace}/.local/validator${i}
        node_ids="$(${bin} tendermint show-node-id --home ${workspace}/.local/validator${i})@127.0.0.1:$((${p2p_port_start}+${i})) ${node_ids}"
    done

    persistent_peers=$(joinByString ',' ${node_ids})
    for ((i=0;i<${size};i++));do
        cp ${workspace}/.local/validator0/config/genesis.json ${workspace}/.local/validator${i}/config/
        sed -i -e "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"0${BASIC_DENOM}\"/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/denom-to-suggest = \"uatom\"/denom-to-suggest = \"${BASIC_DENOM}\"/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/stake/${BASIC_DENOM}/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/\"denom_metadata\": \[\]/\"denom_metadata\": \[${NATIVE_COIN_DESC}\]/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/persistent_peers = \".*\"/persistent_peers = \"${persistent_peers}\"/g" ${workspace}/.local/validator${i}/config/config.toml
        sed -i -e "s/addr_book_strict = true/addr_book_strict = false/g" ${workspace}/.local/validator${i}/config/config.toml
        sed -i -e "s/allow_duplicate_ip = false/allow_duplicate_ip = true/g" ${workspace}/.local/validator${i}/config/config.toml
        
    done
}

function start() {
    size=$1
    for ((i=0;i<${size};i++));do
        mkdir -p ${workspace}/.local/validator${i}/logs
        nohup ${bin} start --home ${workspace}/.local/validator${i} \
            --address 0.0.0.0:$((${address_port_start}+${i})) \
            --grpc-web.address 0.0.0.0:$((${grpc_web_port_start}+${i})) \
            --grpc.address 0.0.0.0:$((${grpc_port_start}+${i})) \
            --p2p.laddr tcp://0.0.0.0:$((${p2p_port_start}+${i})) \
            --p2p.external-address 127.0.0.1:$((${p2p_port_start}+${i})) \
            --rpc.laddr tcp://0.0.0.0:$((${rpc_port_start}+${i})) \
            --log_format json > ${workspace}/.local/validator${i}/logs/node.log &
    done
}

function stop() {
    killall ${bin_name}
}


CMD=$1
SIZE=3
if [ ! -z $2 ] && [ "$2" -gt "0" ]; then
    SIZE=$2
fi

case ${CMD} in
init)
    echo "===== init ===="
    init $SIZE
    echo "===== end ===="
    ;;
generate)
    echo "===== generate genesis ===="
    generate_genesis $SIZE
    echo "===== end ===="
    ;;
start)
    echo "===== start ===="
    start $SIZE
    echo "===== end ===="
    ;;
stop)
    echo "===== stop ===="
    stop
    echo "===== end ===="
    ;;
all)
    echo "===== init ===="
    init $SIZE
    echo "===== generate genesis ===="
    generate_genesis $SIZE
    echo "===== start ===="
    start $SIZE
    echo "===== end ===="
    ;;
*)
    echo "Usage: localup.sh all | init | generate | start | stop"
    ;;
esac