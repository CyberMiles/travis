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

* `provider` The provider to connect.
* `wallet` The wallet of the sending account.
* `password` The password of the from account, to sign the transaction with.
* `to` The address that all transactions are directed to.
* `contractAddress` The contract address for testing token transfer.
* `value` The value transferred for the transaction in Wei, or token number if it's a token transfer testing.
* `txs` Number of transactions to send for each account.
* `blockTimeout` Max blocks to wait before stop testing.
* `concurrency` The maximum number of parallel requests at a time.
* `waitInterval` The intervals (in milliseconds) to check if all transactions are finished processing.
