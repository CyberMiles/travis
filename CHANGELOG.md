# Changelog

## v0.1.3-beta-hotfix2

_December 4th, 2018_

Fix an emergency bug.  

## v0.1.3-beta-hotfix1

_November 19th, 2018_

Fix a bug concerning the slashing mechanism.  

## v0.1.3-beta

_November 13th, 2018_

### FEATURES

- Fix a bug on voting power computation which allowed some validators to take disproportionately large amounts of block awards.
- Allow Validators to change its default compensation rate.
- Make backup validators participate in normal operations to ensure that the backup is always ready.

### IMPROVEMENTS

- Allow validators to temporally deactivate itself for emergencies.
- Support hot swap for validators to decrease slashing risk.

### FIXES
- Many stability improvements.

## v0.1.2-beta

_October 15th, 2018_

The Mainnet version is released. This version fixed all the bugs found in Travis v0.1.0 and Travis v0.1.1-beta.

## v0.1.0-rc.7

_September 14th, 2018_

### FEATURES

- Improve block award system and its effectiveness: Block reward will be released when a block is committed. 

### IMPROVEMENTS

- Complete the staking system:
  - Check time stamp when a block is proposed. The time stamp won’t be earlier than the previous block. 
  - The reward information of each block will be stored in leveldb.
  - Add an interface to check the reward information of each block. 
- Improve system flexibility: Configuration parameters are stored in a genesis.json file.
- Add test cases and modify the test scripts. 

### FIXES
- Fixed some small bugs 

## v0.1.0-rc.6

_August 31st, 2018_

### FEATURES

- Continue with the development of DPOS 1.4. Ranking and block rewards of Validator is directly correlated to their participation, contribution, loyalty and governance. For detailed algorithm, please refer to the DPOS document on https://www.cybermiles.io/validator/.
- Improve governance mechanism to incentivise worthy delegators. With an interface of setCompRate, Validator can reward an individual delegator by setting a higher compensation rate for him/her. 

### IMPROVEMENTS

- Make transaction more efficient: Use local client to replace rpc client to communicate with Tendermint Core.
- Make transaction more convenient: Support array style json in system parameters. 
- Improve System testability:
  - Add concurrent test and input check test cases 
  - Benchmark test supports KeepAlive mode

### FIXES
- Fixed some small bugs 

## v0.1.0-rc.5

_August 17th, 2018_

### FEATURES
- Improve system security: Verify Delegator's transaction signature in CMT Cube. Make sure that the transaction is initiated by CMT Cube.
- Revise the fault tolerance mechanism in deployment of libENI. 
  * If any Node fails to download the library, it won’t go through the panic program. To ensure the Node can run normally, the network records the failure status, and allows downloading manually.
  * To secure connectivity of global Nodes, more libENI downloading addresses are added. If the Nodes fail to download the library with the first URL, they with try with the rest in order.
  
### IMPROVEMENTS
- Improve system security: 
  * sendTransaction & sendRawTransaction no longer run through txpool API. This avoids bugs from using web3 or geth. 
  * In regard to staking or governance transaction, noncelock will be released when a transaction is signed, instead of when the commit is completed. 
- Enhance usability:
  * Support configuring the number of Backup Validators.
  * Add more Lity related test cases. 
  
### FIXES
- Fix an error in punishing a Validator committed Byzantine failures.
- Fix an error caused by travis tx in synchronising the new Validators.

## v0.1.0-rc.4

_August 3rd, 2018_

### FEATURES
- Upgrade on Lity and CVM: Support registration or upgrade of libENI in Governance.
- Enhancement in DPoS Protocol: Replace Ranking Power with Voting Power. On top of stakes, ranking and compensation of Validator will be determined by participation, diversity, loyalty and growth of community. 
For detailed algorithm, please refer to our DPoS Protocol: https://www.cybermiles.io/validator/

### IMPROVEMENTS
- Modify Governance mechanism: Support setting an expiration date on a Governance proposal, either with timestamp or block height. 
- Improve system stability: Trigger db transaction with every block created. Skating and governance in SQLite database is operated in the same db transaction. 

## v0.1.0-rc.3

_July 20th, 2018_

### FEATURES
We hit a big milestone this week.

Lity is a new programming language for developing smart contract on Cybermiles blockchain. It consists of a dynamically extensible language, a compiler, and a virtual machine runtime. 

Lity is a superset of Solidity with ourstanding flexibility, performance and security - The dApp developers would love these upgrades. 

- The libENI dynamic VM extension allows native functions to be added to the virtual machine on the fly, without stopping, forking or upgrading the blockchain. 
- The ERC checker not only checks but also automagically fixes common security bugs in smart contracts. 
- The upcoming Lity Rules Engine allows formal business rules to be embedded in smart contracts. 

For more information, visit https://www.litylang.org/

### IMPROVEMENTS
1. Compatible with Ethereum: Upgrade go-ethereum to version 1.8.12 
2. Improve security: Staking in CMT cube requires signing by address. 
3. Complete the Governance and Staking mechanism Documentation: http://travis.readthedocs.io/

### FIXES
Fix some small bugs.

## v0.1.0-rc.2

_July 13th, 2018_

### IMPROVEMENTS
- Modify the governance mechanism: A validator can vote multiple times before the proposal is executed. Only the last vote counts.
- Update tendermit to v0.22.0.  
- Improve network security by adding：
  * Backup Validator test-cases
  * System parameters test-cases 
  * Block Award calculation test-cases
- Fix compatibility issues of 0x0 address.

### FIXES

- Correct Validator and Backup Validator block award calculation errors.

## v0.1.0-rc.1

_July 5th, 2018_

### FEATURES

- Gas fee: Charge Validator for declaring candidacy, updating candidate information and proposing transactions. 
- Governance Transactions: Change system parameters through governance transactions.

### IMPROVEMENTS

- Update tendermit to v0.20.0.
- Add Candidate information fields: name, email, profile.
- Change parameters of ChainId:  18: mainnet, 19: testnet, 20: staging.
- Add cmt.syncing to get node syncing status.

### FIXES

- If the maximum staking amount decreases, Validator self-staked CMTs won't be charged.
- Correct non-running Validators won’t get block awards.
- Fix Block Award calculation error.
- Correct delegator address when a Validator withdraws candidacy.
