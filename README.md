# Travis
The first production version of the CyberMiles blockchain.

## Installation

```shell
$ go get github.com/CyberMiles/travis
$ cd $GOPATH/src/github.com/CyberMiles/travis
$ make all
```

## Initialize

```
$ travis node init --home ~/.travis
```

## Start Travis

```
$ travis node start --home ~/.travis
```

## Send transactions

```
$ geth attach http://localhost:8545
```

```
personal.unlockAccount("from_address")
eth.sendTransaction({"from": "from_address", "to": "to_address", "value": web3.toWei(0.001, "ether")})
```
