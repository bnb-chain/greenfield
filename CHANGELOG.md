# Changelog

## v0.0.6
This release includes following features:
1. Support cross chain governance;
2. Storage module improvement;
3. Add e2e test framework and swagger API script;

* [\#76](https://github.com/bnb-chain/greenfield/pull/76) feat: update tx simulation in sdk
* [\#63](https://github.com/bnb-chain/greenfield/pull/63) feat: enable params change proposal for cross chain governance
* [\#74](https://github.com/bnb-chain/greenfield/pull/74) docs: local network
* [\#72](https://github.com/bnb-chain/greenfield/pull/72) doc: add overview and tutorial doc for governance
* [\#73](https://github.com/bnb-chain/greenfield/pull/73) doc: add doc for the bridge module
* [\#71](https://github.com/bnb-chain/greenfield/pull/71) docs: add consensus and staking part
* [\#69](https://github.com/bnb-chain/greenfield/pull/69) ci: add gosec checker
* [\#65](https://github.com/bnb-chain/greenfield/pull/65) docs: add the key management docs and keyring module tutorial
* [\#66](https://github.com/bnb-chain/greenfield/pull/66) feat: add decimals of BNB and gweiBNB and e2e test of gashub module
* [\#67](https://github.com/bnb-chain/greenfield/pull/67) docs: add the overview of architecture && cross chain
* [\#64](https://github.com/bnb-chain/greenfield/pull/64) feat: sdk gas price simulation
* [\#57](https://github.com/bnb-chain/greenfield/pull/57) feat: storage module improvement
* [\#53](https://github.com/bnb-chain/greenfield/pull/53) feat: add init function for gashub module
* [\#60](https://github.com/bnb-chain/greenfield/pull/60) fix: unify denom as BNB
* [\#62](https://github.com/bnb-chain/greenfield/pull/62) refactor: rewrite the NewGreenfieldClient function by option pattern
* [\#58](https://github.com/bnb-chain/greenfield/pull/58) docs: add the token economics docs and bank module tutorial
* [\#59](https://github.com/bnb-chain/greenfield/pull/59) feat: gov RegisterInterfaces for sdk
* [\#54](https://github.com/bnb-chain/greenfield/pull/54) feat: Add deploy scripts for sp
* [\#48](https://github.com/bnb-chain/greenfield/pull/48) feat: add go-sdk and e2e test framework
* [\#56](https://github.com/bnb-chain/greenfield/pull/56) docs: build the framework of docs
* [\#49](https://github.com/bnb-chain/greenfield/pull/49) feat: Add event for storage module
* [\#55](https://github.com/bnb-chain/greenfield/pull/55) feat: enable swagger api for local deployment
* [\#51](https://github.com/bnb-chain/greenfield/pull/51) feat: added proto-gen-swagger for both greenfield and cosmos-sdk
* [\#44](https://github.com/bnb-chain/greenfield/pull/44) feat: add payment to storage module
* [\#47](https://github.com/bnb-chain/greenfield/pull/47) feat: add config for cross chain in env


## v0.0.5
This release includes features, mainly:
1. Implement payment module;
2. Implement storage provider module;
3. Implement storage management module.

* [\#42](https://github.com/bnb-chain/greenfield/pull/42) chore: run goimportssort over the repo
* [\#18](https://github.com/bnb-chain/greenfield/pull/18) feat: add storage module
* [\#5](https://github.com/bnb-chain/greenfield/pull/5) feats: init payment module
* [\#39](https://github.com/bnb-chain/greenfield/pull/39) doc: add events doc for the whole project
* [\#41](https://github.com/bnb-chain/greenfield/pull/41) feat: add more field of storage event
* [\#40](https://github.com/bnb-chain/greenfield/pull/40) feat: add comments for the events of bridge module
* [\#38](https://github.com/bnb-chain/greenfield/pull/38) ci: fix Dockerfile and add docker image release job
* [\#35](https://github.com/bnb-chain/greenfield/pull/35) deploy: update deployment scripts
* [\#46](https://github.com/bnb-chain/greenfield/pull/36) deployment: add bls env to setup script
* [\#34](https://github.com/bnb-chain/greenfield/pull/34) feat: add gashub module
* [\#6](https://github.com/bnb-chain/greenfield/pull/6) feat: add sp module
* [\#32](https://github.com/bnb-chain/greenfield/pull/32) feat: add support for EVM jsonrpc
* [\#33](https://github.com/bnb-chain/greenfield/pull/33) fix: revert gashub module and fix build error

## v0.0.4
This release is for rebranding from `inscription` to `greenfield`, renaming is applied to all packages, files.

* [\#30](https://github.com/bnb-chain/greenfield/pull/30) feat: rebrand from inscription to greenfield

## v0.0.3
This release includes features and bug fixes, mainly:
1. Implement the cross chain communication layer;
2. Implement the cross chain token transfer;
3. Add scripts to set up full nodes using state sync.

* [\#8](https://github.com/bnb-chain/greenfield/pull/8) feat: implement bridge module
* [\#27](https://github.com/bnb-chain/greenfield/pull/27) feat: remove ValAddress and update EIP712 related functions
* [\#26](https://github.com/bnb-chain/greenfield/pull/26) fix: init viper before getting configs
* [\#25](https://github.com/bnb-chain/greenfield/pull/25) deployment: add scripts to set up full nodes using state sync

## v0.0.2
This release includes features and bug fixes, mainly:
1. Customized upgrade module;
2. Customized Tendermint with vote pool;
3. EIP712 bug fix;
4. Deployment scripts fix.

* [\#17](https://github.com/bnb-chain/greenfield/pull/17) feat: custom upgrade module
* [\#20](https://github.com/bnb-chain/greenfield/pull/20) ci: fix release flow
* [\#21](https://github.com/bnb-chain/greenfield/pull/21) feat: init balance of relayers in genesis state
* [\#19](https://github.com/bnb-chain/greenfield/pull/19) deployment: fix relayer key generation
* [\#16](https://github.com/bnb-chain/greenfield/pull/16) feat: pass config to app when creating new app
* [\#14](https://github.com/bnb-chain/greenfield/pull/16) disable unnecessary modules


## v0.0.1
This is the first release of the greenfield.

It includes three key features:
1. EIP721 signing schema support
2. New staking mechanism
3. Local network set scripts


FEATURES
* [\#11](https://github.com/bnb-chain/greenfield/pull/11) feat: customize staking module for greenfield 
* [\#10](https://github.com/bnb-chain/greenfield/pull/10) deployment: local setup scripts
* [\#2](https://github.com/bnb-chain/greenfield/pull/2) feat: add support for EIP712 and eth address standard r4r

DEP
* [\#3](https://github.com/bnb-chain/greenfield/pull/3) dep: replace cosmos-sdk to greenfield-cosmos-sdk v0.0.1(v0.46.4)

BUG FIX
* [\#9](https://github.com/bnb-chain/greenfield/pull/9) fix: add coin type to init cmd

DOCS
* [\#1](https://github.com/bnb-chain/greenfield/pull/1) docs: refine the readme with detailed introduction documentation
