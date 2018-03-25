# Travis Benchmarking

## Requirement
* node `^8.0.0`
* yarn `^1.0.0` or npm `^5.0.0`

## Installation
```bash
yarn install    # Install project dependencies (or `npm install`)
```

## Usage

* send raw transactions
```bash
node sendRawTx
```

* send transactions
```bash
node sendTx
```

## Configuration
Configuration file: config/default.json.

* `providers` A provider list to connect.
* `address` The address that all transactions are directed to.
* `wallet` The wallet of the sending account.
* `password` The password of the from account, to sign the transaction with.
* `txs` Number of transactions to send for each account.
* `blockTimeout` Max blocks to wait before stop testing.
* `concurrency` The maximum number of parallel requests at a time.
* `waitInterval` The intervals (in milliseconds) to check if all transactions are finished processing.
