#!/usr/bin/env bash
function checksum() {
    echo $(sha256sum $@ | awk '{print $1}')
}

declare change_log_file="./CHANGELOG.md"
declare version="## $@"
declare version_prefix="## v"
declare start=0
declare CHANGE_LOG=""

while read line; do
    if [[ $line == *"$version"* ]]; then
        start=1
        continue
    fi
    if [[ $line == *"$version_prefix"* ]] && [ $start == 1 ]; then
        break;
    fi
    if [ $start == 1 ]; then
        CHANGE_LOG+="$line\n"
    fi
done < "${change_log_file}"

LINUX_BIN_SUM="$(checksum ./linux/linux)"
MAC_BIN_SUM="$(checksum ./macos/macos)"
TESTNET_CONFIG_SUM="$(checksum ./testnet_config.zip)"
MAINNET_CONFIG_SUM="$(checksum ./mainnet_config.zip)"

OUTPUT=$(cat <<-END
## Changelog\n
${CHANGE_LOG}\n
## Assets\n
|    Assets    | Sha256 Checksum  |
| :-----------: |------------|
| linux | ${LINUX_BIN_SUM} |
| mac  | ${MAC_BIN_SUM} |
| testnet_config.zip  | ${TESTNET_CONFIG_SUM} |\n
| mainnet_config.zip  | ${MAINNET_CONFIG_SUM} |\n
END
)

echo -e "${OUTPUT}"