======================
Deploy a TestNet Node
======================

In this document, we will discuss how to connect to the CyberMiles Travis TestNet. We will cover binary, Docker and "build from source" scenarios. If you are new to CyberMiles, deploying a Docker node is probably easier.

While we highly recommend you to run your own Travis node, you can also ask for direct access to one of the nodes maintained by the CyberMiles Foundation. Send an email to travis@cybermiles.io to apply for access credentials. You still need the ``travis`` client either from Docker or source to access the node.

Binary
======

Make sure your os is Ubuntu 16.04 or CentOS 7

Download pre-built binaries from `release page <https://github.com/CyberMiles/travis/releases/tag/vTestnet>`_
-----------------------------------------------------------------------------------------------------------

::

  mkdir -p $HOME/release
  cd $HOME/release
  
  # if your os is Ubuntu
  wget https://github.com/CyberMiles/travis/releases/download/vTestnet/travis_vTestnet_ubuntu-16.04.zip
  unzip travis_vTestnet_ubuntu-16.04.zip

  # or if your os is CentOS
  wget https://github.com/CyberMiles/travis/releases/download/vTestnet/travis_vTestnet_centos-7.zip
  unzip travis_vTestnet_centos-7.zip

Getting Travis TestNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  cd $HOME/release
  ./travis node init --env testnet
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/genesis.json > $HOME/.travis/config/genesis.json


Change your name from default name ``local``

::

  cd $HOME/.travis
  vim $HOME/.travis/config/config.toml

  # here you can change your name
  moniker = "<your_custom_name>"

Copy libeni into the default Travis data directory
--------------------------------------------------

::

  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib
  
  # set env variables for eni lib
  # for convenience, you should also put these two lines in your .bashrc or .zshrc
  export ENI_LIBRARY_PATH=$HOME/.travis/eni/lib
  export LD_LIBRARY_PATH=$HOME/.travis/eni/lib

Start the Node and Join Travis TestNet
--------------------------------------

::

  cd $HOME/release
  ./travis node start


Docker
======

Prerequisite
------------
Please `setup docker <https://docs.docker.com/engine/installation/>`_.

Docker Image
------------
Docker image for Travis is stored on `Docker Hub <https://hub.docker.com/r/cybermiles/travis/tags/>`_. TestNet environment is using the `'vTestnet' <https://github.com/CyberMiles/travis/releases/tag/vTestnet>`_ release which can be pulled automatically from Travis:

::

  docker pull cybermiles/travis:vTestnet

Note: Configuration and data will be stored at /travis directory in the container. The directory will also be exposed as a volume. The ports 8545, 26656 and 26657 will be exposed for connection.

Getting Travis TestNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  docker run --rm -v $HOME/.travis:/travis -t cybermiles/travis:vTestnet node init --env testnet --home /travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/genesis.json > $HOME/.travis/config/genesis.json

Start the Node and Join Travis TestNet
--------------------------------------
First change your name from default name ``local``

::

  vim ~/.travis/config/config.toml

  # here you can change your name
  moniker = "<your_custom_name>"

Run the docker Travis application:

::

  docker run --name travis -v $HOME/.travis:/travis -p 26657:26657 -p 8545:8545 -t cybermiles/travis:vTestnet node start --home /travis

Now your node is syncing with TestNet, the output will look like the following. Wait until it completely syncs.

::

  INFO [07-20|03:13:26.229] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=1.002ms   mgasps=0.000    number=3363 hash=4884c0…212e75 cache=2.22mB
  I[07-20|03:13:26.241] Committed state                              module=state height=3363 txs=0 appHash=3E0C01B22217A46676897FCF2B91DB7398B34262
  I[07-20|03:13:26.443] Executed block                               module=state height=3364 validTxs=0 invalidTxs=0
  I[07-20|03:13:26.443] Updates to validators                        module=state updates="[{\"address\":\"\",\"pub_key\":\"VPsUJ1Eb73tYPFhNjo/8YIWY9oxbnXyW+BDQsTSci2s=\",\"power\":27065},{\"address\":\"\",\"pub_key\":\"8k17vhQf+IcrmxBiftyccq6AAHAwcVmEr8GCHdTUnv4=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"PoDmSVZ/qUOEuiM38CtZvm2XuNmExR0JkXMM9P9UhLU=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"2Tl5oI35/+tljgDKzypt44rD1vjVHaWJFTBdVLsmcL4=\",\"power\":27048}]"

To access the TestNet type the following in a seperte terminal console to get your IP address then use your IP address to connect to the TestNet.

::

  docker inspect -f '{{ .NetworkSettings.IPAddress }}' travis
  172.17.0.2
  docker run --rm -it cybermiles/travis:vTestnet attach http://172.17.0.2:8545

Now, you should see the web3-cmt JavaScript console, you can now jump to the "Test transactions" section to send test transactions.


Snapshot
========

Make sure your os is Ubuntu 16.04 or CentOS 7

Download snapshot file from AWS S3 `travis-ss-testnet <https://s3-us-west-2.amazonaws.com/travis-ss-testnet>`_
------------------------------------------------------------------------------------------------------------

You can splice the file name from the bucket list. The downloading url will be like ``https://s3-us-west-2.amazonaws.com/travis-ss-testnet/testnet/travis_ss_testnet_1542277779_226170.tar``. You must have found that the file name contains timestamp and block number at which the snapshot is made.

::

  mkdir -p $HOME/release
  cd $HOME/release
  wget https://s3-us-west-2.amazonaws.com/travis-ss-testnet/testnet/travis_ss_testnet_1542277779_226170.tar
  tar xf travis_ss_testnet_1542277779_226170.tar

  # if your os is Ubuntu
  mv .travis/app/travis .
  mkdir .travis/eni
  mv .travis/app/lib .travis/eni
  mv .travis $HOME

  # or if your os is CentOS
  mv .travis $HOME
  wget https://github.com/CyberMiles/travis/releases/download/vTestnet/travis_vTestnet_centos-7.zip
  unzip travis_vTestnet_centos-7.zip
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib

Set env variables for eni lib
--------------------------------------------------

::

  # for convenience, you should also put these two lines in your .bashrc or .zshrc
  export ENI_LIBRARY_PATH=$HOME/.travis/eni/lib
  export LD_LIBRARY_PATH=$HOME/.travis/eni/lib

Start the Node and Join Travis TestNet
--------------------------------------
First download the config and change your name from default name ``local``, set persistent peers

::

  mkdir $HOME/.travis/config
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  vim ~/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

Start the application

::

  cd $HOME/release
  ./travis node start --home $HOME/.travis


Build from source
=================

Prerequisite
------------
Please `install Travis via source builds <http://travis.readthedocs.io/en/latest/getting-started.html#build-from-source>`_. (STOP before you connect to a local node)

Getting Travis TestNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  travis node init --env testnet
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/genesis.json > $HOME/.travis/config/genesis.json

Start the Node and Join Travis TestNet
--------------------------------------
Run the Travis application:

::

  travis node start --home ~/.travis

Now your node is syncing with TestNet, the output will look like the following. Wait until it completely syncs.

::

  INFO [07-20|03:13:26.229] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=1.002ms   mgasps=0.000    number=3363 hash=4884c0…212e75 cache=2.22mB
  I[07-20|03:13:26.241] Committed state                              module=state height=3363 txs=0 appHash=3E0C01B22217A46676897FCF2B91DB7398B34262
  I[07-20|03:13:26.443] Executed block                               module=state height=3364 validTxs=0 invalidTxs=0
  I[07-20|03:13:26.443] Updates to validators                        module=state updates="[{\"address\":\"\",\"pub_key\":\"VPsUJ1Eb73tYPFhNjo/8YIWY9oxbnXyW+BDQsTSci2s=\",\"power\":27065},{\"address\":\"\",\"pub_key\":\"8k17vhQf+IcrmxBiftyccq6AAHAwcVmEr8GCHdTUnv4=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"PoDmSVZ/qUOEuiM38CtZvm2XuNmExR0JkXMM9P9UhLU=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"2Tl5oI35/+tljgDKzypt44rD1vjVHaWJFTBdVLsmcL4=\",\"power\":27048}]"

To access the TestNet, type the following in a seperte terminal console (make sure that the seperate console also has travis environment):

::

  travis attach http://localhost:8545

You should now the see the web3-cmt JavaScript console and can now test some transactions.

Test transactions
=================

In this section, we will use the ``travis`` client's web3-cmt JavaScript console to send some transactions and verify that the system is set up properly. You can't test transactions untill you are completely in sync with the TestNet. It might take hours to sync.

Create and fund a test account
-------------------------------

Once you attach the ``travis`` to the node as above, create two accounts on the TestNet.

::

  Welcome to the Geth JavaScript console!
  > personal.newAccount()
  ...

Now you have created TWO accounts ``0x1234FROM`` and ``0x1234DEST`` on the Travis TestNet. It is time to get some test CMTs. Please go visit the website below, and ask for 1000 TestNet CMTs for account ``0x1234FROM``. We will also send 1000 TEST tokens, issued by the TEST smart contract, to the account.

http://travis-faucet.cybermiles.io
 

Test transactions
-----------------

You can test transactions between your two accounts. Remember to unlock both of your accounts.

::

  > personal.unlockAccount("0x1234FROM","password")
  true
  ...
  > cmt.sendTransaction({from:"0x1234FROM", to:"0x1234DEST",value:1000})
  ...
  > cmt.getBalance("0x1234DEST")
  ...
  
You can also test smart contract transactions for the TEST token as below.

::

  > abi = [{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"INITIAL_SUPPLY","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_subtractedValue","type":"uint256"}],"name":"decreaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_addedValue","type":"uint256"}],"name":"increaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"previousOwner","type":"address"},{"indexed":true,"name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]
  > tokenContract = web3.cmt.contract(abi)
  > tokenInstance = tokenContract.at("0xb6b29ef90120bec597939e0eda6b8a9164f75deb")
  > tokenInstance.transfer.sendTransaction("0x1234DEST", 1000, {from: "0x1234FROM"})

After 10 seconds, you can check the balance of the receiving account as follows.

::

  > tokenInstance.balanceOf.call("0x1234DEST")

Fee free transactions
---------------------

On CyberMiles blockchain, we have made most transactions (except for heavy users or spammers) fee-free. You can try it like this in ``travis`` client console.

::

  > cmt.sendTransaction({from:"0x1234FROM", to:"0x1234DEST",value:1000,gasPrice:0})
  ...

To try a fee-free smart contract-based token transaction, use the following in the ``travis`` client console.

::

  > tokenInstance.transfer.sendTransaction("0x1234DEST", 1000, {from: "0x1234FROM", gasPrice: 0})


