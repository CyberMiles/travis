# Travis
[![Build Status develop branch](https://travis-ci.org/CyberMiles/travis.svg?branch=develop)](https://travis-ci.org/CyberMiles/travis)

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
$ travis attach http://localhost:8545
```

```
personal.unlockAccount("from_address")
cmt.sendTransaction({"from": "from_address", "to": "to_address", "value": web3.toWei(0.001, "cmt")})
```
