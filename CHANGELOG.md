# Changelog

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
Fixed some small bugs.

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
