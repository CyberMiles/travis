================================
Frequently Asked Questions (FAQ)
================================

In this document, we will list the frequently asked technical and general questions and answers. If you can't find you question in this FAQ, create an issue on `CyberMiles <https://github.com/CyberMiles/travis/issues/new>`_.

Technical
-----------------------------
1. Hardware Requirement for Validator
``````````````````````````````````````
Operating System:

Ubuntu 16.04 LTS/Centos 7, docker is not recommended for production environment due to performance monitoring and debug complexity.

Disk Space:

Minimal disk space is 500GB SSD for the first year, more disk space will be needed as applications on travis are widely applied and transactions increase. For cloud server, it is easy to increase the disk space to at most 64T without shutting down servers or restarting service.

Physical Server:

CPU: 8 cores, 16 threads
Memory: 14 RAM
Port: 26656, 26657
Networking: 10 Gbps

Cloud Server:


Cost:

It is hard to estimate the total expense of physical servers depended on various types of CPU and SSD and different cable services. Yearly cost estimation for 1 validator server + 2 sentry nodes + 500G x 3 in GCP is $12443.89 plus around $1000 network fee(depend on the traffic load):

2. Sentry Node
```````````````
The Sentry Node Architecture(SNA) is an infrastructure example for DDoS mitigation on travis validator nodes.
On the travis network, a supernode can be attacked using the Distributed Denial of Service method. The validator node has a fixed IP address and it opens a RESTful API port facing the Internet. To mitigate the issue, multiple distributed nodes (sentry nodes) are deployed in cloud environments. With the possibility of easy scaling, it is harder to make an impact on the validator node. New sentry nodes can be brought up during a DDoS attack and using the gossip network they can be integrated into the transaction flow.


3. How to check my address, transactions, and validator status ?
`````````````````````````````````````````````````````````````````
Mainnet explorer:
https://www.cmttracking.io/

Testnet explorer:
https://testnet.cmttracking.io/


4. How could I get some CMTs in testnet/mainnet ?
```````````````````````````````````````````````````
Mainnet CMT:

get CMTs in crypto exchanges(binance, huobi, okex etc.)

Testnet CMT:

use testnet faucet to get test CMTs.
http://travis-faucet.cybermiles.io/index.html


5. What's the chainId for testnet and mainnet ?
````````````````````````````````````````````````
Mainnet chainId: 18

Testnet chainId: 19

6. How to access testnet/mainnet ?
```````````````````````````````````
Travis cli

.. code:: bash
   
   travis attach http://{your_host}:8545

RPC hosts

  * mainnet: rpc.cybermiles.io 8545
  * testnet: testnet.cmtwallet.io 8545

RPC developer doc: 
https://travis.readthedocs.io/en/latest/json-rpc.html


7. How to open 8545 port?
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

8. How to use RPC ?
````````````````````````````````

https://travis.readthedocs.io/en/latest/json-rpc.html

9. How to calculate validator's voting power or delegator's rewards ?
``````````````````````````````````````````````````````````````````````
English version(page 10 ~ 21): https://www.cybermiles.io/wp-content/uploads/2018/10/EN_CyberMiles_DPoS_1.4.2.pdf

中文版(第8 ~ 17页): https://www.cybermiles.io/wp-content/uploads/2018/10/CN_CyberMiles_DPoS.pdf

10. How to recover/re-sync my node when it crashed or missed too many blocks?
`````````````````````````````````````````````````````````````````````````````
use snapshot to catch up the block data quickly. https://travis.readthedocs.io/en/latest/connect-mainnet.html#snapshot

11. When did CyberMiles transfer CMT from Ethereum and what’s the height?
`````````````````````````````````````````````````````````````````````````
Ethereum Transaction Hash: https://etherscan.io/tx/0x5ff71c1be6ee3512bb574eb900e0954041031d3e0b7e544535ca7c95c8f4ccf0

Height: 6486489 (https://etherscan.io/block/6486489)

Official CMT Contract address in Ethereum: https://etherscan.io/address/0xf85feea2fdd81d51177f6b8f35f0e6734ce45f5f

12. How to update a new address of existing validator? Does this action affect all the delegators of this validator?
``````````````````````````````````````````````````````````````````````````````````````````````````````````````````````
For validator, use web3-cmt.js or rpc to call updateAccount method to update your validator address. follow the command and example in https://cybermiles.github.io/web3-cmt.js/api/validator.html#updateaccount. It will affect your delegators. 

But if your validator misses more than 120 blocks when you try to update your address, your validator node will be slashed and all the delegators on this node will be affeced. If so, manual activate your node is needed(https://cybermiles.github.io/web3-cmt.js/api/validator.html#activate).
