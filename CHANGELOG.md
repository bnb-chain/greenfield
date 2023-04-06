# Changelog
## v0.1.0
* [\#141](https://github.com/bnb-chain/greenfield/pull/141) fix wrong comments for events.proto in storage (created_at field shows block timestamp instead of block number)
* [\#149](https://github.com/bnb-chain/greenfield/pull/149) fix: get price boundary logic, remove useless query, enhance ci
* [\#148](https://github.com/bnb-chain/greenfield/pull/148) fix: gas simulation issue  
* [\#146](https://github.com/bnb-chain/greenfield/pull/146) fix: incorrect assignment of the CreateAt field in EventCreateObject  
* [\#145](https://github.com/bnb-chain/greenfield/pull/145) feat: add expiration time to event  
* [\#151](https://github.com/bnb-chain/greenfield/pull/151) feat: Refine events and bugfix
* [\#153](https://github.com/bnb-chain/greenfield/pull/153) fix: Revert event field changes  
* [\#154](https://github.com/bnb-chain/greenfield/pull/154) fix: Revert event field changes 
* [\#155](https://github.com/bnb-chain/greenfield/pull/155) docs: remove docs
* [\#144](https://github.com/bnb-chain/greenfield/pull/144) fix: get price boundary logic, remove useless query, enhance ci
* [\#157](https://github.com/bnb-chain/greenfield/pull/157) fix: occasionally timeout in auto-settle e2e test
* [\#156](https://github.com/bnb-chain/greenfield/pull/156) patch v0.0.11 fix to main branch
* [\#160](https://github.com/bnb-chain/greenfield/pull/160) feat: Only the creator and owner are allowed to cancel create objects 
* [\#161](https://github.com/bnb-chain/greenfield/pull/161) fix: update license from GPL to AGPL
* [\#158](https://github.com/bnb-chain/greenfield/pull/158) fix: sdk gas simulation  

## v0.0.11
* [\#140](https://github.com/bnb-chain/greenfield/pull/140) fix: `Visibility` type in `CreateBucketSynPackage`
* [\#139](https://github.com/bnb-chain/greenfield/pull/139) fix: payment queries

## v0.0.10
* [\#136](https://github.com/bnb-chain/greenfield/pull/136) fix: forbid a stream account with negative static balance to pay
* [\#135](https://github.com/bnb-chain/greenfield/pull/135) modify default to inherit
* [\#132](https://github.com/bnb-chain/greenfield/pull/132) feat: allow unordered attestations
* [\#131](https://github.com/bnb-chain/greenfield/pull/131) feat: support delete bucket with non-zero charged read quota
* [\#108](https://github.com/bnb-chain/greenfield/pull/108) chore: refine storage module
* [\#126](https://github.com/bnb-chain/greenfield/pull/126) chore: refine permission module
* [\#125](https://github.com/bnb-chain/greenfield/pull/125) chore: refine bridge module
* [\#124](https://github.com/bnb-chain/greenfield/pull/124) feat: The visibility property of the Bucket&Object can be modified
* [\#112](https://github.com/bnb-chain/greenfield/pull/115) chore: refine sp module
* [\#129](https://github.com/bnb-chain/greenfield/pull/129) chore: refine payment module
* [\#117](https://github.com/bnb-chain/greenfield/pull/117) feat: implement validator tax in storage payment
* [\#116](https://github.com/bnb-chain/greenfield/pull/116) feat: implement challenge module
* [\#130](https://github.com/bnb-chain/greenfield/pull/130) fix: check status of object before mirroring
* [\#122](https://github.com/bnb-chain/greenfield/pull/122) chore: refine permission module for code quality 
* [\#128](https://github.com/bnb-chain/greenfield/pull/128) docs: add the banner of gnfd
* [\#121](https://github.com/bnb-chain/greenfield/pull/121) chore: code quality self-review of ante handler
* [\#103](https://github.com/bnb-chain/greenfield/pull/103) feat: add challenger address to validators
* [\#123](https://github.com/bnb-chain/greenfield/pull/123) feat: add max buckets per account to params
* [\#110](https://github.com/bnb-chain/greenfield/pull/110) feat: add expiration and limit size for policy
* [\#118](https://github.com/bnb-chain/greenfield/pull/118) chore: refine the code of the storage mode


## v0.0.9
The resource mirror is introduced in this release.

* [\#109](https://github.com/bnb-chain/greenfield/pull/109) feat: add events for permission module
* [\#107](https://github.com/bnb-chain/greenfield/pull/107) fix: update license and readme
* [\#104](https://github.com/bnb-chain/greenfield/pull/104) fix: refine the bridge module
* [\#101](https://github.com/bnb-chain/greenfield/pull/101) doc: add detail doc for permission module
* [\#50](https://github.com/bnb-chain/greenfield/pull/50) feat: add cross chain txs for storage resources
* [\#114](https://github.com/bnb-chain/greenfield/pull/114) fix: potential attack risks in storage module


## v0.0.8

This release includes following features:
1. Add enhancements to storage module;
2. Revert the related changes about the callbackGasprice;

* [\#100](https://github.com/bnb-chain/greenfield/pull/100) revert: revert the related changes about the callbackGasprice
* [\#102](https://github.com/bnb-chain/greenfield/pull/102) feat: Enhancement storage module

## v0.0.7
This release includes following features:
1. Implement permission module;
2. Implement challenge module;
3. Refactor payment module;
4. Storage module improvement;
5. SP module improvement;

* [\#70](https://github.com/bnb-chain/greenfield/pull/70) fix: Storage Provider account check
* [\#81](https://github.com/bnb-chain/greenfield/pull/81) doc: add payment module doc
* [\#80](https://github.com/bnb-chain/greenfield/pull/80) feat: add index by id for storage module
* [\#82](https://github.com/bnb-chain/greenfield/pull/82) bugfix list object/bucket error
* [\#85](https://github.com/bnb-chain/greenfield/pull/85) fix is non-empty bucket error
* [\#79](https://github.com/bnb-chain/greenfield/pull/79) doc: add data availability challenge doc
* [\#90](https://github.com/bnb-chain/greenfield/pull/90) feat: update default cross-chain transfer out fee
* [\#83](https://github.com/bnb-chain/greenfield/pull/83) feat: enable querying bucket, object and group by id
* [\#91](https://github.com/bnb-chain/greenfield/pull/91) complete acc address best practice
* [\#92](https://github.com/bnb-chain/greenfield/pull/92) fix: update gas price and consensus params
* [\#94](https://github.com/bnb-chain/greenfield/pull/94) feat: support customized nonce
* [\#75](https://github.com/bnb-chain/greenfield/pull/75) doc: add SP and storage meta doc
* [\#43](https://github.com/bnb-chain/greenfield/pull/43) feat: implement challenge module
* [\#96](https://github.com/bnb-chain/greenfield/pull/96) docs: refine the document structure
* [\#88](https://github.com/bnb-chain/greenfield/pull/88) feat: improve sp module
* [\#95](https://github.com/bnb-chain/greenfield/pull/95) feat: update the crosschain keeper in bridge module
* [\#87](https://github.com/bnb-chain/greenfield/pull/87) feat: refactor payment module
* [\#97](https://github.com/bnb-chain/greenfield/pull/97) feat: update default parameters for challenge module
* [\#93](https://github.com/bnb-chain/greenfield/pull/93) refactor: change addr in payment module from string to AccAccount
* [\#68](https://github.com/bnb-chain/greenfield/pull/68) feat: Implement permission module
* [\#89](https://github.com/bnb-chain/greenfield/pull/89) feat: Create storage provider in genesis by genesis transaction

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
