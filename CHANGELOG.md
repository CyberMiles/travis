# Changelog

## v0.1.0-rc.1

_July 5th, 2018_

### FEATURES

- Charge gas fee for declareCandidacy, updateCandidacy and propose transactions.
- Make changes to the system parameters through governance transactions.

### IMPROVEMENTS

- Update tendermit to v0.20.0.
- Add more fields to Candidacy(name, email, profile).
- ChainId: 18: mainnet, 19: testnet, 20: staging.
- Add cmt.syncing to get node syncing status.

### FIXES

- If the max amount of CMTs is decreased, no additional self-staked CMTs should be charged.
- Fake validators shouldn't get block awards.
- Block award calculation error.
- Incorrect delegator address provided while withdrawing candidacy.
