================================
Frequently Asked Questions (FAQ)
================================

In this document, we will list the frequently asked technical and general questions and answers. If you can't find you question in this FAQ, create an issue on `CyberMiles <https://github.com/CyberMiles/travis/issues/new>`_.

General
----------------------------









Technical
-----------------------------

1. How to check my address, transactions, and validator status ?
`````````````````````````````````````````````````````````````````
Mainnet explorer:
https://www.cmttracking.io/

Testnet explorer:
https://testnet.cmttracking.io/


2. How could I get some CMTs in testnet/mainnet ?
```````````````````````````````````````````````````
Mainnet CMT:

get CMTs in crypto exchanges(binance, huobi, okex etc.)

Testnet CMT:

use testnet faucet to get test CMTs.
http://travis-faucet.cybermiles.io/index.html


3. What's the chainId for testnet and mainnet ?
````````````````````````````````````````````````
Mainnet chainId: 18

Testnet chainId: 19

4. How to access testnet/mainnet ?
```````````````````````````````````
Travis cli

.. code:: bash
   
   travis attach http://{your_host}:8545

RPC hosts

  * mainnet: rpc.cybermiles.io 8545
  * testnet: testnet.cmtwallet.io 8545

RPC developer doc: 
https://travis.readthedocs.io/en/latest/json-rpc.html


5. How to open 8545 port?
``````````````````````````
Attention: You may lose your asset when 8545 port is turned on and is accessible by hackers.

set rpc = true.

.. code:: bash

   vim $HOME/.travis/config/config.toml
   
   # this is an example in mainnet
   # at bottom of the config file
   [vm]
   chainid = 18
   rpc = true
   rpcapi = "cmt,eth,net,web3"
   rpcaddr = "0.0.0.0"
   rpcport = 8545
   rpccorsdomain = "*"
   rpcvhosts = "localhost"
   ws = false
   verbosity = "3"

6. How to use RPC ?
````````````````````````````````

https://travis.readthedocs.io/en/latest/json-rpc.html

7. How to calculate validator's voting power or delegator's rewards ?
``````````````````````````````````````````````````````````````````````
English version(page 10 ~ 21): https://www.cybermiles.io/wp-content/uploads/2018/10/CN_CyberMiles_DPoS.pdf

中文版(第8 ~ 17页): https://www.cybermiles.io/wp-content/uploads/2018/10/EN_CyberMiles_DPoS_1.4.2.pdf

8. How to recover/re-sync my node when it crashed or missed too many blocks?
`````````````````````````````````````````````````````````````````````````````
use snapshot to catch up the block data quickly. https://travis.readthedocs.io/en/latest/connect-mainnet.html#snapshot

9. When did CyberMiles transfer CMT from Ethereum and what’s the height?
`````````````````````````````````````````````````````````````````````````
Ethereum Transaction Hash: https://etherscan.io/tx/0x5ff71c1be6ee3512bb574eb900e0954041031d3e0b7e544535ca7c95c8f4ccf0

Height: 6486489 (https://etherscan.io/block/6486489)

Official CMT Contract address in Ethereum: https://etherscan.io/address/0xf85feea2fdd81d51177f6b8f35f0e6734ce45f5f

10. How to update a new address of existing validator? Does this action affect all the delegators of this validator?
``````````````````````````````````````````````````````````````````````````````````````````````````````````````````````
For validator, use web3-cmt.js or rpc to call updateAccount method to update your validator address. follow the command and example in https://cybermiles.github.io/web3-cmt.js/api/validator.html#updateaccount. It will affect your delegators. 

But if your validator misses more than 120 blocks when you try to update your address, your validator node will be slashed and all the delegators on this node will be affeced. If so, manual activate your node is needed(https://cybermiles.github.io/web3-cmt.js/api/validator.html#activate).
