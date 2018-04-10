# Travis
[![Build Status develop branch](https://travis-ci.org/CyberMiles/travis.svg?branch=develop)](https://travis-ci.org/CyberMiles/travis)

The first production version of the CyberMiles blockchain.

## Installation

```shell
$ go get github.com/CyberMiles/travis
$ cd $GOPATH/src/github.com/CyberMiles/travis
$ git checkout master
$ make all
```

If the system cannot find `glide` at the last step, make sure that you have `$GOPATH/bin` under the `$PATH` variable.

## Initialize a Travis node

```
$ travis node init --home ~/.travis
```

## Start a Travis node

```
$ travis node start --home ~/.travis
```

## Start a Travis client and send transactions

```
$ travis attach http://localhost:8545
```

```
personal.unlockAccount("from_address")
cmt.sendTransaction({"from": "from_address", "to": "to_address", "value": web3.toWei(0.001, "cmt")})
```
