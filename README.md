# Travis

## Initialize tendermint & ethermint

```
$ mkdir -p ~/.ethermint/tendermint
$ cp -r $GOPATH/src/github.com/tendermint/ethermint/setup/* ~/.ethermint/
$ tendermint init --home ~/.ethermint/tendermint/
$ ethermint --datadir ~/.ethermint/ init ~/.ethermint/genesis.json
```


## Start Travis
```
$ make all
$ travis node start
```

## Start Tendermint
```
tendermint --home ~/.ethermint/tendermint/ node 
```

## Start Ethermint
```
ethermint --datadir ~/.ethermint --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin --abci_laddr tcp://0.0.0.0:8848
```

## Send transactions

```
$ geth attach http://localhost:8545
```

```
personal.unlockAccount("from_address")
eth.sendTransaction({"from": "from_address", "to": "to_address", "value": web3.toWei(0.001, "ether")})
```

