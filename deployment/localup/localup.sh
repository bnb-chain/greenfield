#!/usr/bin/env bash
basedir=$(cd `dirname $0`; pwd)
workspace=${basedir}
source ${workspace}/.env
source ${workspace}/utils.sh

bin_name=gnfd
bin=${workspace}/../../build/bin/${bin_name}

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
        ${bin} keys add validator${i} --keyring-backend test --home ${workspace}/.local/validator${i} > ${workspace}/.local/validator${i}/info 2>&1
        ${bin} keys add relayer${i} --keyring-backend test --home ${workspace}/.local/relayer${i} > ${workspace}/.local/relayer${i}/relayer_info 2>&1
        ${bin} keys add relayer_bls${i} --keyring-backend test --home ${workspace}/.local/relayer${i} --algo eth_bls > ${workspace}/.local/relayer${i}/relayer_bls_info 2>&1
    done

    # add sp account
    sp_size=1
    if [ $# -eq 2 ];then
      sp_size=$2
    fi
    for ((i=0;i<${sp_size};i++));do
      #create sp and sp fund account
      mkdir -p ${workspace}/.local/sp${i}
      ${bin} keys add sp${i} --keyring-backend test --home ${workspace}/.local/sp${i} > ${workspace}/.local/sp${i}/info 2>&1
      ${bin} keys add sp${i}_fund --keyring-backend test --home ${workspace}/.local/sp${i} > ${workspace}/.local/sp${i}/fund_info 2>&1
    done

}

function generate_genesis() {
    # create sp address in genesis
    sp_size=1
    if [ $# -eq 2 ];then
      sp_size=$2
    fi
    for ((i=0;i<${sp_size};i++));do
      #create sp and sp fund account
      sp_addrs=("$(${bin} keys show sp${i} -a --keyring-backend test --home ${workspace}/.local/sp${i})")
      spfund_addrs=("$(${bin} keys show sp${i}_fund -a --keyring-backend test --home ${workspace}/.local/sp${i})")
      ${bin} add-genesis-account $sp_addrs ${GENESIS_ACCOUNT_BALANCE}${STAKING_BOND_DENOM} --home ${workspace}/.local/validator0
      ${bin} add-genesis-account $spfund_addrs ${GENESIS_ACCOUNT_BALANCE}${STAKING_BOND_DENOM} --home ${workspace}/.local/validator0
    done

    size=$1
    declare -a validator_addrs=()
    for ((i=0;i<${size};i++));do
        # export validator addresses
        validator_addrs+=("$(${bin} keys show validator${i} -a --keyring-backend test --home ${workspace}/.local/validator${i})")
    done

    declare -a relayer_addrs=()
    for ((i=0;i<${size};i++));do
        # export validator addresses
        relayer_addrs+=("$(${bin} keys show relayer${i} -a --keyring-backend test --home ${workspace}/.local/relayer${i})")
    done

    mkdir -p ${workspace}/.local/gentx
    for ((i=0;i<${size};i++));do
        for validator_addr in "${validator_addrs[@]}";do
            # init genesis account in genesis state
            ${bin} add-genesis-account $validator_addr ${GENESIS_ACCOUNT_BALANCE}${STAKING_BOND_DENOM} --home ${workspace}/.local/validator${i}
        done

        for relayer_addr in "${relayer_addrs[@]}";do
            # init genesis account in genesis state
            ${bin} add-genesis-account $relayer_addr ${GENESIS_ACCOUNT_BALANCE}${STAKING_BOND_DENOM} --home ${workspace}/.local/validator${i}
        done

        rm -rf ${workspace}/.local/validator${i}/config/gentx/

        validatorAddr=${validator_addrs[$i]}
        relayerAddr="$(${bin} keys show relayer${i} -a --keyring-backend test --home ${workspace}/.local/relayer${i})"
        relayerBLSKey="$(${bin} keys show relayer_bls${i} --keyring-backend test --home ${workspace}/.local/relayer${i} --output json | jq -r .pubkey_hex)"

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
            --node tcp://localhost:$((${VALIDATOR_P2P_PORT_START}+${i})) \
            --node-id "validator${i}" \
            --ip 127.0.0.1
        cp ${workspace}/.local/validator${i}/config/gentx/gentx-validator${i}.json ${workspace}/.local/gentx/
    done

    node_ids=""
    # bond validator tx in genesis state
    for ((i=0;i<${size};i++));do
        cp ${workspace}/.local/gentx/* ${workspace}/.local/validator${i}/config/gentx/
        ${bin} collect-gentxs --home ${workspace}/.local/validator${i}
        node_ids="$(${bin} tendermint show-node-id --home ${workspace}/.local/validator${i})@127.0.0.1:$((${VALIDATOR_P2P_PORT_START}+${i})) ${node_ids}"
    done

    persistent_peers=$(joinByString ',' ${node_ids})
    for ((i=0;i<${size};i++));do
        cp ${workspace}/.local/validator0/config/genesis.json ${workspace}/.local/validator${i}/config/
        sed -i -e "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"0${BASIC_DENOM}\"/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/denom-to-suggest = \"uatom\"/denom-to-suggest = \"${BASIC_DENOM}\"/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/\"stake\"/\"${BASIC_DENOM}\"/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/\"denom_metadata\": \[\]/\"denom_metadata\": \[${NATIVE_COIN_DESC}\]/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/persistent_peers = \".*\"/persistent_peers = \"${persistent_peers}\"/g" ${workspace}/.local/validator${i}/config/config.toml
        sed -i -e "s/addr_book_strict = true/addr_book_strict = false/g" ${workspace}/.local/validator${i}/config/config.toml
        sed -i -e "s/allow_duplicate_ip = false/allow_duplicate_ip = true/g" ${workspace}/.local/validator${i}/config/config.toml
        sed -i -e "s/snapshot-interval = 0/snapshot-interval = ${SNAPSHOT_INTERVAL}/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/src-chain-id = 1/src-chain-id = ${SRC_CHAIN_ID}/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/dest-chain-id = 2/dest-chain-id = ${DEST_CHAIN_ID}/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/snapshot-keep-recent = 2/snapshot-keep-recent = ${SNAPSHOT_KEEP_RECENT}/g" ${workspace}/.local/validator${i}/config/app.toml
        sed -i -e "s/\"reserve_time\": \"15552000\"/\"reserve_time\": \"600\"/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/\"forced_settle_time\": \"86400\"/\"forced_settle_time\": \"100\"/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/172800s/${DEPOSIT_VOTE_PERIOD}/g" ${workspace}/.local/validator${i}/config/genesis.json
        sed -i -e "s/\"10000000\"/\"${MIN_DEPOSIT_AMOUNT}\"/g" ${workspace}/.local/validator${i}/config/genesis.json
    done
}

function start() {
    size=$1
    for ((i=0;i<${size};i++));do
        mkdir -p ${workspace}/.local/validator${i}/logs
        nohup ${bin} start --home ${workspace}/.local/validator${i} \
            --address 0.0.0.0:$((${VALIDATOR_ADDRESS_PORT_START}+${i})) \
            --grpc-web.address 0.0.0.0:$((${VALIDATOR_GRPC_WEB_PORT_START}+${i})) \
            --grpc.address 0.0.0.0:$((${VALIDATOR_GRPC_PORT_START}+${i})) \
            --p2p.laddr tcp://0.0.0.0:$((${VALIDATOR_P2P_PORT_START}+${i})) \
            --p2p.external-address 127.0.0.1:$((${VALIDATOR_P2P_PORT_START}+${i})) \
            --rpc.laddr tcp://0.0.0.0:$((${VALIDATOR_RPC_PORT_START}+${i})) \
            --json-rpc.address 127.0.0.1:$((${VALIDATOR_JSONRPC_PORT_START}+${i}+${i})) \
            --json-rpc.ws-address 127.0.0.1:$((${VALIDATOR_JSONRPC_PORT_START}+${i}+${i}+1)) \
            --log_format json > ${workspace}/.local/validator${i}/logs/node.log &
    done
}

function stop() {
    ps -ef | grep ${bin_name} | grep validator | awk '{print $2}' | xargs kill
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
    echo "===== stop ===="
    stop
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
