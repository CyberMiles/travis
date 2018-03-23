# Travis Benchmarking

## Requirement
* node `^8.0.0`
* yarn `^1.0.0` or npm `^5.0.0`

## Installation
```bash
yarn install    # Install project dependencies (or `npm install`)
```

## Usage

* send raw transactions from one account
```bash
node oneAccount
```

* send raw transactions from multiple account
```bash
node multipleAccount
```

## Configuration
Configuration file: config/default.json.

* `providers` A provider list to connect.
* `address` The address that all transactions are directed to.
* `wallet` The wallet to generate more sending accounts.
* `password` The password of the from account, to sign the transaction with.
* `n` Number of transactions to send for each account.
* `accounts` Number of accounts to send transactions.
* `blockTimeout` Max blocks to wait before stop testing.
