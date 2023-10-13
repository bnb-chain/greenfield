# Changelog

## v1.0.1
This release contains 2 chores.

Chores:
* [#500](https://github.com/bnb-chain/greenfield/pull/500) chore: add mainnet assets
* [#501](https://github.com/bnb-chain/greenfield/pull/501) chore: add sp price related cli and refine cli and comments


## v1.0.0
This release contains 1 chore.

Chores:
* [#496](https://github.com/bnb-chain/greenfield/pull/496) chore: add events for parameter updates

## v0.2.6-hf.1
This is a hotfix version to enable the chain to emit deposit events for user accounts and add local virtual group id for related events.

Chores:
* [#489](https://github.com/bnb-chain/greenfield/pull/489) chore: add deposit event for user account
* [#490](https://github.com/bnb-chain/greenfield/pull/490) chore: add lvg id to related events

## v0.2.6
This is a maintenance release to update the cosmos-sdk dependency to the latest version.

## v0.2.6-alpha.1
This release contains 3 new bugfixes.

Bugfixes:
* [#478](https://github.com/bnb-chain/greenfield/pull/478) fix: wrong event emitted for leaving group
* [#481](https://github.com/bnb-chain/greenfield/pull/481) fix: policy id not emit in deletepolicy event
* [#482](https://github.com/bnb-chain/greenfield/pull/482) fix: use versioned segment size parameter

Chores:
* [#480](https://github.com/bnb-chain/greenfield/pull/480) chore: improve the validations of parameters

## v0.2.5
This release contains all the changes for the v0.2.5 alpha versions.    

Features:
* [#435](https://github.com/bnb-chain/greenfield/pull/435) feat: deposit the balance of the bank account to the payment account
* [#448](https://github.com/bnb-chain/greenfield/pull/448) feat: timelock for large amount withdraw from payment
* [#449](https://github.com/bnb-chain/greenfield/pull/449) feat: complete cdc register
* [#457](https://github.com/bnb-chain/greenfield/pull/457) feat: add api for querying last quota update time

Bugfixes:
* [#471](https://github.com/bnb-chain/greenfield/pull/471) fix: some issues in challenge module
* [#470](https://github.com/bnb-chain/greenfield/pull/470) fix: testnet block sync issue
* [#451](https://github.com/bnb-chain/greenfield/pull/451) fix: audit issues by verichain
* [#456](https://github.com/bnb-chain/greenfield/pull/456) fix: fix parameter init issue
* [#458](https://github.com/bnb-chain/greenfield/pull/458) fix: correct emit event filed
* [#462](https://github.com/bnb-chain/greenfield/pull/462) fix: fix app hash mismatch for genesis block

Chore: 
* [#474](https://github.com/bnb-chain/greenfield/pull/474) chore: cleaning up ci jobs

## v0.2.5-alpha.3
This release is a bugfix release.

Bugfixes:
* [#470](https://github.com/bnb-chain/greenfield/pull/470) fix: testnet block sync issue
* [#471](https://github.com/bnb-chain/greenfield/pull/471) fix: some issues in challenge module

## v0.2.5-alpha.2
This release contains 1 bugfix.
* [#465](https://github.com/bnb-chain/greenfield/pull/465) fix: remaining policies need re-persistence

## v0.2.5-alpha.1
This release contains 4 features and 4 bugfixes.

Features:
* [#435](https://github.com/bnb-chain/greenfield/pull/435) feat: deposit the balance of the bank account to the payment account
* [#448](https://github.com/bnb-chain/greenfield/pull/448) feat: timelock for large amount withdraw from payment
* [#449](https://github.com/bnb-chain/greenfield/pull/449) feat: complete cdc register
* [#457](https://github.com/bnb-chain/greenfield/pull/457) feat: add api for querying last quota update time

Bugfixes:
* [#451](https://github.com/bnb-chain/greenfield/pull/451) fix: audit issues by verichain
* [#456](https://github.com/bnb-chain/greenfield/pull/456) fix: fix parameter init issue
* [#458](https://github.com/bnb-chain/greenfield/pull/458) fix: correct emit event filed
* [#462](https://github.com/bnb-chain/greenfield/pull/462) fix: fix app hash mismatch for genesis block

## v0.2.4
This release contains all the changes in the v0.2.4 alpha versions and 5 new bugfixes.

Bugfixes:
* [#433](https://github.com/bnb-chain/greenfield/pull/433) fix: add sp exit status check when craete gvg
* [#437](https://github.com/bnb-chain/greenfield/pull/437) fix: fix the last block time is empty issue
* [#443](https://github.com/bnb-chain/greenfield/pull/443) fix: fix some issues in payment module
* [#432](https://github.com/bnb-chain/greenfield/pull/432) fix: grant permission with wildcard objectname
* [#440](https://github.com/bnb-chain/greenfield/pull/440) fix: check secondary sp in gvg

Chores:
* [#431](https://github.com/bnb-chain/greenfield/pull/431) chore: fix sp e2e test sometimes fail issue
* [#442](https://github.com/bnb-chain/greenfield/pull/442) chore: switch the order of create sp event and update price event

## v0.2.4-alpha.3
This release contains 1 bugfix.

Bugfixes:
* [#428](https://github.com/bnb-chain/greenfield/pull/428) fix: fix early deletion fee calculation

## v0.2.4-alpha.2
This release contains 4 features and 4 bugfixes.      

Features:
* [#411](https://github.com/bnb-chain/greenfield/pull/411) feat: add RemoveExpiredPolicies to add method to remove expired data from kvstore
* [#413](https://github.com/bnb-chain/greenfield/pull/413) feat: implement cross-chain mechanism between op and greenfield
* [#415](https://github.com/bnb-chain/greenfield/pull/415) feat: enable plain store for full node
* [#420](https://github.com/bnb-chain/greenfield/pull/420) feat: skip sig verification on genesis block

Bugfixes:
* [#416](https://github.com/bnb-chain/greenfield/pull/416) fix: update the dependencies to the latest develop branch
* [#419](https://github.com/bnb-chain/greenfield/pull/419) fix: validate bls key and proof before submitting proposal
* [#410](https://github.com/bnb-chain/greenfield/pull/410) fix: fix group member key collision issues
* [#422](https://github.com/bnb-chain/greenfield/pull/422) fix: fix the SettleTimestamp calculation

Chores:
* [#408](https://github.com/bnb-chain/greenfield/pull/408) chore: payment refactor to use global store prices for billing
* [#409](https://github.com/bnb-chain/greenfield/pull/409) chore: add issue template
* [#414](https://github.com/bnb-chain/greenfield/pull/414) chore: modify default gas
* [#417](https://github.com/bnb-chain/greenfield/pull/417) chore: adapt to cross chain token mintable version
* [#421](https://github.com/bnb-chain/greenfield/pull/421) chore: update go version to 1.20


## v0.2.4-alpha.1
This release includes 4 features, 9 bugfixes and 2 documentation updates.

Features:
* [#374](https://github.com/bnb-chain/greenfield/pull/374) feat: group member expiration
* [#390](https://github.com/bnb-chain/greenfield/pull/390) feat: add flag to enable/disable heavy queries and refactor apis
* [#399](https://github.com/bnb-chain/greenfield/pull/399) feat: add new query APIs for group and group member  
* [#403](https://github.com/bnb-chain/greenfield/pull/403) feat: sp maintenance mode

Bugfixes:
* [#377](https://github.com/bnb-chain/greenfield/pull/377) fix: improve e2e tests to include more coverage from server side
* [#379](https://github.com/bnb-chain/greenfield/pull/379) fix: error member name in transferInRefundPackageType
* [#383](https://github.com/bnb-chain/greenfield/pull/383) fix: fix lock balance not updated for frozen payment account
* [#385](https://github.com/bnb-chain/greenfield/pull/385) fix: fix returned operation type in group cross chain app
* [#391](https://github.com/bnb-chain/greenfield/pull/391) fix: add cancel mb event for discontinue and delete
* [#398](https://github.com/bnb-chain/greenfield/pull/398) fix: register gov cross-chain app
* [#400](https://github.com/bnb-chain/greenfield/pull/400) fix: allow edit-sp by cmd without blskey
* [#401](https://github.com/bnb-chain/greenfield/pull/401) fix: add group existence check when verify permission
* [#375](https://github.com/bnb-chain/greenfield/pull/375) fix: fix defining err for ErrInvalidBlsPubKey  

Chores:
* [#376](https://github.com/bnb-chain/greenfield/pull/376) chore: add unit tests for the storage module
* [#378](https://github.com/bnb-chain/greenfield/pull/378) chore: add unit test cases for challenge module
* [#380](https://github.com/bnb-chain/greenfield/pull/380) chore: add unit test cases for payment module
* [#381](https://github.com/bnb-chain/greenfield/pull/381) chore: add tests for bridge module
* [#387](https://github.com/bnb-chain/greenfield/pull/387) chore: add more e2e test cases for payment module
* [#388](https://github.com/bnb-chain/greenfield/pull/388) chore: add cli tests
* [#389](https://github.com/bnb-chain/greenfield/pull/389) chore: add more test cases for cross chain apps

Documentation:  
* [#402](https://github.com/bnb-chain/greenfield/pull/402) docs: update testnet asset to v0.2.3
* [#404](https://github.com/bnb-chain/greenfield/pull/404) docs: update document site link

## 0.2.3
This is a official release for v0.2.3, includes all the changes since v0.2.2.

Bugfixes:
* [#375](https://github.com/bnb-chain/greenfield/pull/375) fix: defining err
* [#379](https://github.com/bnb-chain/greenfield/pull/379) fix: error member name in transferInRefundPackageType
* [#383](https://github.com/bnb-chain/greenfield/pull/383) fix: lock balance not updated for frozen payment account
* [#385](https://github.com/bnb-chain/greenfield/pull/385) fix: returned operation type in group cross chain app

Chores:
* [#376](https://github.com/bnb-chain/greenfield/pull/376) chore: add unit tests for the storage module
* [#377](https://github.com/bnb-chain/greenfield/pull/377) chore: improve e2e tests to include more coverage from server side
* [#380](https://github.com/bnb-chain/greenfield/pull/380) chore: add unit test cases for payment module
* [#381](https://github.com/bnb-chain/greenfield/pull/381) chore: add tests for bridge module
* [#383](https://github.com/bnb-chain/greenfield/pull/383) chore: add unit test cases for challenge module

## 0.2.3-alpha.7
This release includes 2 features and 3 bugfixes.

Features:
* [#366](https://github.com/bnb-chain/greenfield/pull/366) feat: add strategy for event emit
* [#368](https://github.com/bnb-chain/greenfield/pull/368) feat: limit sp slash amount and add query api

Bugfixes:
* [#369](https://github.com/bnb-chain/greenfield/pull/369) fix: parse failed when object name contains /
* [#370](https://github.com/bnb-chain/greenfield/pull/370) fix: fix the precision issue of storage bill
* [#371](https://github.com/bnb-chain/greenfield/pull/371) fix: add src dst sp check when migrate bucket

Chores:
* [#365](https://github.com/bnb-chain/greenfield/pull/365) ci: add e2e test coverage report

## 0.2.3-alpha.6
This release includes 6 features and 5 bugfixes.

Features:
* [#346](https://github.com/bnb-chain/greenfield/pull/346) feat: Enable websocket client as a option in Greenfield sdk
* [#349](https://github.com/bnb-chain/greenfield/pull/349) feat: add UpdateChannelPermissions tx for crosschain module
* [#350](https://github.com/bnb-chain/greenfield/pull/350) feat: add create storage provider command
* [#357](https://github.com/bnb-chain/greenfield/pull/357) feat: add api to filter virtual group families qualification
* [#352](https://github.com/bnb-chain/greenfield/pull/352) feat: add query params for virtual group
* [#348](https://github.com/bnb-chain/greenfield/pull/348) feat: fix issues and add test cases for payment

Bugfixes:
* [#351](https://github.com/bnb-chain/greenfield/pull/351) fix: update local virtual group event
* [#353](https://github.com/bnb-chain/greenfield/pull/353) fix: panic when delete unsealed object from lvg
* [#354](https://github.com/bnb-chain/greenfield/pull/354) fix: incorrect authority for keepers
* [#342](https://github.com/bnb-chain/greenfield/pull/342) fix: remove primary sp id from bucket info
* [#356](https://github.com/bnb-chain/greenfield/pull/356) fix: add validation of extra field when creating group

Chores:
* [#359](https://github.com/bnb-chain/greenfield/pull/359) chore: add swagger check
* [#358](https://github.com/bnb-chain/greenfield/pull/358) chore: Refine event for sp exit and bucket migration

## v0.2.3-alpha.5
This release adds 6 new features and 2 bugfixes.

Features
* [#315](https://github.com/bnb-chain/greenfield/pull/315) feat: add api for querying lock fee
* [#290](https://github.com/bnb-chain/greenfield/pull/290) feat: replace rlp with abi.encode in cross-chain transfer
* [#323](https://github.com/bnb-chain/greenfield/pull/323) feat: enable asset reconciliation
* [#326](https://github.com/bnb-chain/greenfield/pull/326) feat: add bls verification
* [#336](https://github.com/bnb-chain/greenfield/pull/336) feat: add tendermint to sdk
* [#341](https://github.com/bnb-chain/greenfield/pull/341) feat: support cross chain for multiple blockchains
* [#328](https://github.com/bnb-chain/greenfield/pull/328) feat: refactor for sp exit and bucket migrate 

Bugfixes
* [#307](https://github.com/bnb-chain/greenfield/pull/307) fix DefaultMaxPayloadSize from 2GB to 64GB 
* [#312](https://github.com/bnb-chain/greenfield/pull/312) fix: add chainid to sign bytes to prevent replay attack

Documentation
* [#316](https://github.com/bnb-chain/greenfield/pull/316) Update readme.md 
* [#282](https://github.com/bnb-chain/greenfield/pull/282) update readme for helm deployment

Chores
* [#324](https://github.com/bnb-chain/greenfield/pull/324) chore: update greenfield-cometbft-db version 

## v0.2.3-alpha.2
This release enables 2 new features:  

Features  
* [#301](https://github.com/bnb-chain/greenfield/pull/301) feat: add support of group name in policy
* [#304](https://github.com/bnb-chain/greenfield/pull/304) feat: allow simulate create bucket without approval

## v0.2.3-alpha.1
This release enables several features and bugfixes:

Features
* [#281](https://github.com/bnb-chain/greenfield/pull/281) feat: add versioned parameters to payment module 
* [#287](https://github.com/bnb-chain/greenfield/pull/287) feat: use median store price for secondary sp price 
* [#292](https://github.com/bnb-chain/greenfield/pull/292) feat: allows for setting a custom http client when NewGreenfieldClient 
* [#288](https://github.com/bnb-chain/greenfield/pull/288) feat: limit the interval for updating quota
* [#297](https://github.com/bnb-chain/greenfield/pull/297) feat: refine payment and update default parameter     

Bugfixes  
* [#279](https://github.com/bnb-chain/greenfield/pull/279) fix: fix the security issues 
* [#280](https://github.com/bnb-chain/greenfield/pull/280) fix: update go.mod to be compatible with ignite 
* [#286](https://github.com/bnb-chain/greenfield/pull/286) fix: update storage discontinue param's default value 
* [#295](https://github.com/bnb-chain/greenfield/pull/295) add missing field to event 
* [#285](https://github.com/bnb-chain/greenfield/pull/285) fix: ACTION_UPDATE_OBJECT_INFO not allowed to be used on object's bug 

## v0.2.2
This release enables several features and some bugfix:

Features
* [#249](https://github.com/bnb-chain/greenfield/pull/249) feat: support multiple messages in single tx for EIP712
* [#250](https://github.com/bnb-chain/greenfield/pull/250) feat: allow mirror bucket/object/group using name
* [#268](https://github.com/bnb-chain/greenfield/pull/268) feat: record challenge attestation result
* [#276](https://github.com/bnb-chain/greenfield/pull/276) feat: allow user to pass keyManager into Txopt

Bugfix
* [#248](https://github.com/bnb-chain/greenfield/pull/248) fix: add versioned params e2e's test
* [#252](https://github.com/bnb-chain/greenfield/pull/252) fix: remove paramsclient from sdk and swagger
* [#254](https://github.com/bnb-chain/greenfield/pull/254) fix: fix potential int64 multiplication overflow
* [#255](https://github.com/bnb-chain/greenfield/pull/255) fix: verify permission openapi params
* [#263](https://github.com/bnb-chain/greenfield/pull/263) fix: QueryGetSecondarySpStorePriceByTime may wrong data
* [#267](https://github.com/bnb-chain/greenfield/pull/267) chore: update swagger
* [#271](https://github.com/bnb-chain/greenfield/pull/271) fix: check every module's Msg
* [#270](https://github.com/bnb-chain/greenfield/pull/270) fix: sp check when reject seal object
* [#269](https://github.com/bnb-chain/greenfield/pull/269) fix: fix wrong link in readme
* [#274](https://github.com/bnb-chain/greenfield/pull/274) fix: add sp address check when deposit

## v0.2.2-alpha.2

This release enables 2 features:
- record challenge attestation result
- allow user to pass keyManager into Txopt  

* [#267](https://github.com/bnb-chain/greenfield/pull/267) chore: update swagger 
* [#268](https://github.com/bnb-chain/greenfield/pull/268) feat: record challenge attestation result 
* [#271](https://github.com/bnb-chain/greenfield/pull/271) fix: check every module's Msg 
* [#270](https://github.com/bnb-chain/greenfield/pull/270) fix: sp check when reject seal object 
* [#269](https://github.com/bnb-chain/greenfield/pull/269) fix: fix wrong link in readme 
* [#274](https://github.com/bnb-chain/greenfield/pull/274) fix: add sp address check when deposit 
* [#276](https://github.com/bnb-chain/greenfield/pull/276) feat: allow user to pass keyManager into Txopt 

## v0.2.1-alpha.1

This release[CHANGELOG.md](CHANGELOG.md) enable two features:
- support multiple messages in single tx
- allow mirror bucket/object/group using name

Features
* [#250](https://github.com/bnb-chain/greenfield/pull/250) feat: allow mirror bucket/object/group using name
* [#249](https://github.com/bnb-chain/greenfield/pull/249) feat: support multiple messages in single tx for EIP712

Fix
* [#248](https://github.com/bnb-chain/greenfield/pull/248) fix: add versioned params e2e's test
* [#252](https://github.com/bnb-chain/greenfield/pull/252) fix: remove paramsclient from sdk and swagger
* [#255](https://github.com/bnb-chain/greenfield/pull/255) fix: verify permission openapi params
* [#254](https://github.com/bnb-chain/greenfield/pull/254) fix: fix potential int64 multiplication overflow
* [#263](https://github.com/bnb-chain/greenfield/pull/263) fix: QueryGetSecondarySpStorePriceByTime may wrong data


## v0.2.1
This is a hot fix release for v0.2.0
* [#251](https://github.com/bnb-chain/greenfield/pull/251) fix: correct the counting of deleted objects
* [#256](https://github.com/bnb-chain/greenfield/pull/256) dep: bump cosmos-sdk to v0.2.1

## v0.2.0
This release updates the greenfield-cosmos-sdk dependency. Please refer to the [greenfield-cosmos-sdk repository](https://github.com/bnb-chain/greenfield-cosmos-sdk/) for the update details.
* [#188](https://github.com/bnb-chain/greenfield/pull/188) feat: integrate greenfield with cosmos sdk v0.47
* [#190](https://github.com/bnb-chain/greenfield/pull/190) feat: add more fields to sp events
* [#191](https://github.com/bnb-chain/greenfield/pull/191) feat: define the turn for submitting attestation
* [#197](https://github.com/bnb-chain/greenfield/pull/197) fix: fix e2e issues due to km refactor
* [#199](https://github.com/bnb-chain/greenfield/pull/199) feat: migrate challenge module to cosmos sdk v0.47
* [#194](https://github.com/bnb-chain/greenfield/pull/194) feat: fix the issues of commands
* [#200](https://github.com/bnb-chain/greenfield/pull/200) feat: migrate challenge e2e tests
* [#203](https://github.com/bnb-chain/greenfield/pull/203) chore: fix ante test
* [#205](https://github.com/bnb-chain/greenfield/pull/205) ci: run ci jobs for every pull request
* [#206](https://github.com/bnb-chain/greenfield/pull/206) fix: migrate sdk fix to v0.47_latest
* [#207](https://github.com/bnb-chain/greenfield/pull/207) fix: init app with upgrade handlers
* [#208](https://github.com/bnb-chain/greenfield/pull/208) docs: fix localup scripts in document
* [#210](https://github.com/bnb-chain/greenfield/pull/210) feat: remove amino dependencies for GetSignBytes
* [#212](https://github.com/bnb-chain/greenfield/pull/212) feat: add export key for localup script
* [#196](https://github.com/bnb-chain/greenfield/pull/196) feat: modify sp module & storage module & permission module to adapt cosmos sdk v0.47
* [#214](https://github.com/bnb-chain/greenfield/pull/214) fix: fix e2e test for gashub
* [#216](https://github.com/bnb-chain/greenfield/pull/216) feat: payment adapt to cosmos-sdk v0.47
* [#215](https://github.com/bnb-chain/greenfield/pull/215) feat: add update-object-info for updateobject's visibility (cherry pick #138)
* [#220](https://github.com/bnb-chain/greenfield/pull/220) feat: support empty operator for verifypermission
* [#219](https://github.com/bnb-chain/greenfield/pull/219) fix: sp & storage & permission module's cli bug
* [#217](https://github.com/bnb-chain/greenfield/pull/217) feat: remove dependency for params module
* [#221](https://github.com/bnb-chain/greenfield/pull/221) fix: bring back the swagger server
* [#225](https://github.com/bnb-chain/greenfield/pull/225) fix: fix the banner issue and sync a tiny pr
* [#224](https://github.com/bnb-chain/greenfield/pull/224) feat: add support for EVM json-rpc request
* [#226](https://github.com/bnb-chain/greenfield/pull/222) fix nil pointer panic
* [#231](https://github.com/bnb-chain/greenfield/pull/231) feat: update cosmos sdk to v0.47.2  
* [#232](https://github.com/bnb-chain/greenfield/pull/232) fix: fix challenge random issue
* [#218](https://github.com/bnb-chain/greenfield/pull/218) feat: support multi version params for storage module
* [#234](https://github.com/bnb-chain/greenfield/pull/234) fix: sp staking ledger error when slash
* [#223](https://github.com/bnb-chain/greenfield/pull/223) feat: enable stale permission GC
* [#235](https://github.com/bnb-chain/greenfield/pull/235) feat: update dependency for the cosmos-sdk
* [#236](https://github.com/bnb-chain/greenfield/pull/236) fix: update swagger file based on the latest cosmos-sdk
* [#237](https://github.com/bnb-chain/greenfield/pull/237) swagger: replace gov v1beta1 by v1
* [#242](https://github.com/bnb-chain/greenfield/pull/242) fix: replace github.com/gogo/protobuf with github.com/cosmos/gogoproto 

## v0.1.2
* [\#195](https://github.com/bnb-chain/greenfield/pull/195) feat: make sp receive storage fee with funding addr
* [\#167](https://github.com/bnb-chain/greenfield/pull/167) chore: change default sp price
* [\#164](https://github.com/bnb-chain/greenfield/pull/164) feat: update relayer fee for mirror transactions
* [\#168](https://github.com/bnb-chain/greenfield/pull/168) fix: list group error
* [\#170](https://github.com/bnb-chain/greenfield/pull/170) chore: rename the name of buf buf.yaml
* [\#171](https://github.com/bnb-chain/greenfield/pull/171) ci: add testnet_config to release page and fix issues of release flow
* [\#172](https://github.com/bnb-chain/greenfield/pull/172) fix: unify property field names of events
* [\#152](https://github.com/bnb-chain/greenfield/pull/152) feat: add empty object feature for chain
* [\#137](https://github.com/bnb-chain/greenfield/pull/137) feat: allow sp to stop serving objects
* [\#175](https://github.com/bnb-chain/greenfield/pull/175) fix: add ErrInvalidApproval errorcode for sp’s approval invalid
* [\#150](https://github.com/bnb-chain/greenfield/pull/150) feat: refactor key manager to hide private key.
* [\#176](https://github.com/bnb-chain/greenfield/pull/176) fix: init tmclient
* [\#177](https://github.com/bnb-chain/greenfield/pull/177) feat: add more fields to sp events
* [\#165](https://github.com/bnb-chain/greenfield/pull/165) feat: define the turn for submitting attestation
* [\#179](https://github.com/bnb-chain/greenfield/pull/179) fix: fix e2e issues of challenge
* [\#182](https://github.com/bnb-chain/greenfield/pull/182) docs: fix localup scripts in document
* [\#138](https://github.com/bnb-chain/greenfield/pull/138) feat: add update-object-info for updateobject’s visibility (#c)
* [\#180](https://github.com/bnb-chain/greenfield/pull/180) feat: add export key for localup script
* [\#184](https://github.com/bnb-chain/greenfield/pull/184) feat: implement queries and events for frontend
* [\#183](https://github.com/bnb-chain/greenfield/pull/183) fix: remove randomized params from challenge module
* [\#178](https://github.com/bnb-chain/greenfield/pull/178) feat: remove amino dependencies for GetSignBytes
* [\#185](https://github.com/bnb-chain/greenfield/pull/185) fix: sp & storage & permission module’s cli bug
* [\#187](https://github.com/bnb-chain/greenfield/pull/187) feat: support empty operator for verifypermission

## v0.1.1
* [\#166](https://github.com/bnb-chain/greenfield/pull/166) fix: gashub causes state sync to fail to synchronize

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
