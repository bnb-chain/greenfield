# Changelog

## v0.0.3
This release includes features and bug fixes, mainly:
1. Implement the cross chain communication layer;
2. Implement the cross chain token transfer;
3. Add scripts to set up full nodes using state sync.

* [\#8](https://github.com/bnb-chain/inscription/pull/8) feat: implement bridge module
* [\#27](https://github.com/bnb-chain/inscription/pull/27) feat: remove ValAddress and update EIP712 related functions
* [\#26](https://github.com/bnb-chain/inscription/pull/26) fix: init viper before getting configs
* [\#25](https://github.com/bnb-chain/inscription/pull/25) deployment: add scripts to set up full nodes using state sync

## v0.0.2
This release includes features and bug fixes, mainly:
1. Customized upgrade module;
2. Customized Tendermint with vote pool;
3. EIP712 bug fix;
4. Deployment scripts fix.

* [\#17](https://github.com/bnb-chain/inscription/pull/17) feat: custom upgrade module
* [\#20](https://github.com/bnb-chain/inscription/pull/20) ci: fix release flow
* [\#21](https://github.com/bnb-chain/inscription/pull/21) feat: init balance of relayers in genesis state
* [\#19](https://github.com/bnb-chain/inscription/pull/19) deployment: fix relayer key generation
* [\#16](https://github.com/bnb-chain/inscription/pull/16) feat: pass config to app when creating new app
* [\#14](https://github.com/bnb-chain/inscription/pull/16) disable unnecessary modules


## v0.0.1
This is the first release of the inscription.

It includes three key features:
1. EIP721 signing schema support
2. New staking mechanism
3. Local network set scripts


FEATURES
* [\#11](https://github.com/bnb-chain/inscription/pull/11) feat: customize staking module for inscription 
* [\#10](https://github.com/bnb-chain/inscription/pull/10) deployment: local setup scripts
* [\#2](https://github.com/bnb-chain/inscription/pull/2) feat: add support for EIP712 and eth address standard r4r

DEP
* [\#3](https://github.com/bnb-chain/inscription/pull/3) dep: replace cosmos-sdk to inscription-cosmos-sdk v0.0.1(v0.46.4)

BUG FIX
* [\#9](https://github.com/bnb-chain/inscription/pull/9) fix: add coin type to init cmd

DOCS
* [\#1](https://github.com/bnb-chain/inscription/pull/1) docs: refine the readme with detailed introduction documentation
