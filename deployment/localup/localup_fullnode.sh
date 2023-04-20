#!/usr/bin/env bash
basedir=$(cd `dirname $0`; pwd)
workspace=${basedir}
source ${workspace}/.env
source ${workspace}/utils.sh

bin_name=gnfd
bin=${workspace}/../../build/bin/${bin_name}

function init_fullnode() {
    size=$1
    rm -rf ${workspace}/.local/dataseed*
    
    trust_height=$(curl -s http://localhost:${VALIDATOR_RPC_PORT_START}/block | jq -r '.result.block.header.height')
    trust_hash=$(curl -s http://localhost:${VALIDATOR_RPC_PORT_START}/block | jq -r '.result.block_id.hash')
    tmp=""
    n=0
    for v in ${workspace}/.local/validator*;do
        tmp="$tmp localhost:$((${VALIDATOR_RPC_PORT_START}+${n}))"
        ((n++))
    done
    rpc_servers=$(joinByString ',' ${tmp})

    for ((i=0;i<${size};i++));do
        # init chain
        ${bin} init dataseed${i} --chain-id ${CHAIN_ID} --default-denom ${STAKING_BOND_DENOM} --home ${workspace}/.local/dataseed${i}
        # remove unused files
        rm -rf ${workspace}/.local/dataseed${i}/priv_validator_key.json
        # copy configs from validator
        cp ${workspace}/.local/validator0/config/genesis.json ${workspace}/.local/dataseed${i}/config/genesis.json
        cp ${workspace}/.local/validator0/config/app.toml ${workspace}/.local/dataseed${i}/config/app.toml
        cp ${workspace}/.local/validator0/config/client.toml ${workspace}/.local/dataseed${i}/config/client.toml
        cp ${workspace}/.local/validator0/config/config.toml ${workspace}/.local/dataseed${i}/config/config.toml

        # set state sync info
        sed -i 'N;s/\# starting from the height of the snapshot\.\nenable = false/\# starting from the height of the snapshot\.\nenable = true/Mg'sed -i -e 'N;s/\# starting from the height of the snapshot\.\nenable = false/\# starting from the height of the snapshot\.\nenable = true/g' ${workspace}/.local/dataseed${i}/config/config.toml
        sed -i -e "s/trust_height = 0/trust_height = ${trust_height}/g" ${workspace}/.local/dataseed${i}/config/config.toml
        sed -i -e "s/trust_hash = \"\"/trust_hash = \"${trust_hash}\"/g" ${workspace}/.local/dataseed${i}/config/config.toml
        sed -i -e "s/rpc_servers = \"\"/rpc_servers = \"${rpc_servers}\"/g" ${workspace}/.local/dataseed${i}/config/config.toml
    done
}

function start_fullnode() {
    size=$1
    for ((i=0;i<${size};i++));do
        mkdir -p ${workspace}/.local/dataseed${i}/logs
        nohup ${bin} start --home ${workspace}/.local/dataseed${i} \
            --address 0.0.0.0:$((${DATASEED_ADDRESS_PORT_START}+${i})) \
            --api.address 0.0.0.0:$((${DATASEED_GRPC_WEB_PORT_START}+${i})) \
            --grpc.address 0.0.0.0:$((${DATASEED_GRPC_PORT_START}+${i})) \
            --p2p.laddr tcp://0.0.0.0:$((${DATASEED_P2P_PORT_START}+${i})) \
            --p2p.external-address 127.0.0.1:$((${DATASEED_P2P_PORT_START}+${i})) \
            --rpc.laddr tcp://0.0.0.0:$((${DATASEED_RPC_PORT_START}+${i})) \
            --json-rpc.address 127.0.0.1:$((${DATASEED_JSONRPC_PORT_START}+${i}+${i})) \
            --json-rpc.ws-address 127.0.0.1:$((${DATASEED_JSONRPC_PORT_START}+${i}+${i}+1)) \
            --log_format json > ${workspace}/.local/dataseed${i}/logs/node.log &
    done
}

function stop_fullnode() {
    ps -ef | grep ${bin_name} | grep dataseed | awk '{print $2}' | xargs kill
}

CMD=$1
SIZE=3
if [ ! -z $2 ] && [ "$2" -gt "0" ]; then
    SIZE=$2
fi

case ${CMD} in
init)
    echo "===== init ===="
    init_fullnode $SIZE
    echo "===== end ===="
    ;;
start)
    echo "===== start ===="
    start_fullnode $SIZE
    echo "===== end ===="
    ;;
stop)
    echo "===== stop ===="
    stop_fullnode
    echo "===== end ===="
    ;;
all)
    echo "===== stop ===="
    stop_fullnode
    echo "===== init ===="
    init_fullnode $SIZE
    echo "===== start ===="
    start_fullnode $SIZE
    echo "===== end ===="
    ;;
*)
    echo "Usage: localup_fullnode.sh all | init | start | stop"
    ;;
esac
