# Travis
[![Build Status develop branch](https://travis-ci.org/CyberMiles/travis.svg?branch=develop)](https://travis-ci.org/CyberMiles/travis)

The first production version of the CyberMiles blockchain.

You MUST have GO language version 1.9+ installed in order to build and run a Travis node. The easiest way to get GO 1.9 is through the GVM.

```shell
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source $HOME/.gvm/scripts/gvm
gvm install go1.9.2 -B
gvm use go1.9.2 --default
```

## Installation

```shell
$ go get github.com/CyberMiles/travis
$ cd $GOPATH/src/github.com/CyberMiles/travis
$ git checkout master
$ make all
```

If the system cannot find `glide` at the last step, make sure that you have `$GOPATH/bin` under the `$PATH` variable.

For example:

```
export GOPATH=~/.gvm/pkgsets/go1.9.2/global
export GOBIN=$GOPATH/go/bin
export PATH=$GOBIN:$PATH
```

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

## Write a script to check all Travis client account balances
The following script can be pasted into the Travis client, at the > prompt.

```
function checkAllBalances() {
    var totalBal = 0;
    for (var acctNum in cmt.accounts) {
        var acct = cmt.accounts[acctNum];
        var acctBal = web3.fromWei(cmt.getBalance(acct), "cmt");
        totalBal += parseFloat(acctBal);
        console.log("  cmt.accounts[" + acctNum + "]: \t" + acct + " \tbalance: " + acctBal + " CMT");
    }
    console.log("  Total balance: " + totalBal + "CMT");
};
```
### Run the checkAllBalances() script

```
checkAllBalances();
```

### Example output

```
> checkAllBalances();
  cmt.accounts[0]: 	0x6....................................230 	balance: 466.798526 CMT
  cmt.accounts[1]: 	0x6....................................244 	balance: 1531 CMT
  Total balance: 1997.798526CMT
```
