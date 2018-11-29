======================
Deploy a TestNet Node
======================

In this document, we will discuss how to start your own node and connect to the CyberMiles Travis TestNet. While we highly recommend you to run your own Travis node, you could still directly access `RPC services <https://travis.readthedocs.io/en/latest/json-rpc.html>`_ from a node provided by the CyberMiles Foundation at ``testnet.cmtwallet.io:8545``.


********
Snapshot
********

The easiest and fastest way to start a node is to use a snapshot. It is also recommended for most people. You can run the node inside a Docker container or on Ubuntu 16.04 / CentOS 7 servers.

Option 1: Docker from a snapshot
=================================

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

  rm -rf $HOME/.travis && mkdir -p $HOME/.travis/config
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/genesis.json > $HOME/.travis/config/genesis.json

Download snapshot
------------------

Get a list of recent snapshots of the testnet from AWS S3 `travis-ss-testnet <https://s3-us-west-2.amazonaws.com/travis-ss-testnet>`_

You can splice the file name from the bucket list. The downloading url will be like ``https://s3-us-west-2.amazonaws.com/travis-ss-testnet/testnet/travis_ss_testnet_1542623121_254975.tar``. You must have found that the file name contains timestamp and block number at which the snapshot is made.

Extract the file and copy the ``data`` and ``vm`` subdirectories from the uncompressed directory to ``$HOME/.travis``

Start the Node and Join Travis TestNet
--------------------------------------

Change your name from default name ``local``.

::

  vim ~/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

Run the docker Travis application:

::

  docker run --name travis -v $HOME/.travis:/travis -t -p 26657:26657 cybermiles/travis:vTestnet node start --home /travis


Attach to the Node and run web3-cmt.js 
---------------------------------------

In another terminal window, log into the Docker container and then run the ``travis`` client and attach to the node. It will open a console to run ``web3-cmt.js`` commands.

::

  docker exec -it travis bash
  > ./travis attach http://localhost:8545

----

Option 2: Binary from a snapshot
=================================

**Make sure your os is Ubuntu 16.04 or CentOS 7**

Download snapshot
------------------

Get a list of recent snapshots of the testnet from AWS S3 `travis-ss-testnet <https://s3-us-west-2.amazonaws.com/travis-ss-testnet>`_

You can splice the file name from the bucket list. The downloading url will be like ``https://s3-us-west-2.amazonaws.com/travis-ss-testnet/testnet/travis_ss_testnet_1542623121_254975.tar``. You must have found that the file name contains timestamp and block number at which the snapshot is made.

::

  rm -rf $HOME/.travis
  
  mkdir -p $HOME/release
  cd $HOME/release
  SNAPSHOT_URL=$(curl -s http://s3-us-west-2.amazonaws.com/travis-ss-testnet/latest.txt)
  wget $SNAPSHOT_URL
  TAR_FILE="${SNAPSHOT_URL##*/}"
  tar xf $TAR_FILE

  # if your os is Ubuntu 16.04
  mv .travis $HOME
  wget https://github.com/CyberMiles/travis/releases/download/vTestnet/travis_vTestnet_ubuntu-16.04.zip
  unzip travis_vTestnet_ubuntu-16.04.zip
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib
  
  # or if your os is CentOS 7
  mv .travis $HOME
  wget https://github.com/CyberMiles/travis/releases/download/vTestnet/travis_vTestnet_centos-7.zip
  unzip travis_vTestnet_centos-7.zip
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib


Set env variables for eni lib
------------------------------

::

  # for convenience, you should also put these two lines in your .bashrc or .zshrc
  export ENI_LIBRARY_PATH=$HOME/.travis/eni/lib
  export LD_LIBRARY_PATH=$HOME/.travis/eni/lib

Start the Node and Join Travis TestNet
--------------------------------------

Download the testnet config and change your name from default name ``local``.

::

  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init/config/config.toml > $HOME/.travis/config/config.toml
  vim ~/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

Start the application

::

  cd $HOME/release
  ./travis node start --home $HOME/.travis


Attach to the Node and Run web3-cmt.js 
---------------------------------------

In another terminal window, run the ``travis`` client and attach to the node. It will open a console to run ``web3-cmt.js`` commands.

::

  cd $HOME/release
  ./travis attach http://localhost:8545


******************
Test transactions
******************

In this section, we will use the ``travis`` client's web3-cmt JavaScript console to send some transactions and verify that the system is set up properly. You can't test transactions untill you are completely in sync with the TestNet. It might take hours to sync.

Create and fund a test account
===============================

Once you attach the ``travis`` to the node as above, create two accounts on the TestNet.

::

  Welcome to the Geth JavaScript console!
  > personal.newAccount()
  ...

Now you have created TWO accounts ``0x1234FROM`` and ``0x1234DEST`` on the Travis TestNet. It is time to get some test CMTs. Please go visit the website below, and ask for 1000 TestNet CMTs for account ``0x1234FROM``. We will also send 1000 TEST tokens, issued by the TEST smart contract, to the account.

http://travis-faucet.cybermiles.io
 

Test transactions
===============================

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
===============================

On CyberMiles blockchain, we have made most transactions (except for heavy users or spammers) fee-free. You can try it like this in ``travis`` client console.

::

  > cmt.sendTransaction({from:"0x1234FROM", to:"0x1234DEST",value:1000,gasPrice:0})
  ...

To try a fee-free smart contract-based token transaction, use the following in the ``travis`` client console.

::

  > tokenInstance.transfer.sendTransaction("0x1234DEST", 1000, {from: "0x1234FROM", gasPrice: 0})



******************
Sync from Genesis
******************

**Experts Only**: This section is not recommend not necessary for most people. But it is important that we can always start the CyberMiles blockchain from genesis to prove its correctness.

You can always start a new CyberMiles blockchain node from genesis, and sync it all the way to the current block height. The process is fairly involved since it requires you to upgrade and restart the node at certain block heights.

One of the key characteristics of the CyberMiles blockchain is the finality of each block. The blockchain will never fork. It will only produce a new block when 2/3 of the validator voting power reach consensus. Software upgrade on the CyberMiles blockchain is done via consensus. That is, at an agreed upon block height, all nodes must upgrade to a new version of the software to continue. Any node that does not upgrade will not reach consensus with the rest of the blockchain and stop.

The table below shows the software version and their corresponding block heights on the testnet.

============ ====================
Blocks       Software version
============ ====================
0 - 224550   0.1.2-beta
224551 -     0.1.3-beta-hotfix1
============ ====================

The general process for syncing a node from genesis is as follows:

* The 0.1.2-beta software starts from genesis
* It automatically stops at block 224550
* You will download 0.1.3-beta-hotfix1 software, and restart the node
* The process repeats until the block height is current

In the instructions below, we will explain how to sync a Linux binary node and a Docker node from genesis.

Option 3 (the hard way): Binary from genesis
=============================================

**Make sure your os is Ubuntu 16.04 or CentOS 7**

Download pre-built binaries
----------------------------

Get software version 0.1.2-beta from from `release page <https://github.com/CyberMiles/travis/releases/>`_

::

  mkdir -p $HOME/release
  cd $HOME/release
  
  # if your os is Ubuntu
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.2-beta/travis_v0.1.2-beta_ubuntu-16.04.zip
  unzip travis_v0.1.2-beta_ubuntu-16.04.zip

  # or if your os is CentOS
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.2-beta/travis_v0.1.2-beta_centos-7.zip
  unzip travis_v0.1.2-beta_centos-7.zip

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

Upgrade and Continue
---------------------

At certain block heights, the node will stop. Download the next version of the software (e.g., ``0.1.3-beta-hotfix1`` at block height 224550), and restart.

::

  rm -rf $HOME/release
  mkdir -p $HOME/release
  cd $HOME/release
  
  # if your os is Ubuntu
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.3-beta-hotfix1/travis_v0.1.3-beta-hotfix1_ubuntu-16.04.zip
  unzip travis_v0.1.3-beta-hotfix1_ubuntu-16.04.zip

  # or if your os is CentOS
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.3-beta-hotfix1/travis_v0.1.3-beta-hotfix1_centos-7.zip
  unzip travis_v0.1.3-beta-hotfix1_centos-7.zip
  
  ./travis node start

----

Option 4 (the hard way): Docker from genesis
=============================================

Prerequisite
------------

Please `setup docker <https://docs.docker.com/engine/installation/>`_.

Docker Image
------------

Docker image for Travis is stored on `Docker Hub <https://hub.docker.com/r/cybermiles/travis/tags/>`_. Genesis starts from software version ``0.1.2-beta``

::

  docker pull cybermiles/travis:v0.1.2-beta

Note: Configuration and data will be stored at ``/travis`` directory in the container. The directory will also be exposed as a volume. The ports 8545, 26656 and 26657 will be exposed for connection.

Getting Travis TestNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  docker run --rm -v $HOME/.travis:/travis -t cybermiles/travis:v0.1.2-beta node init --env testnet --home /travis
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

  docker run --name travis -v $HOME/.travis:/travis -p 26657:26657 -t cybermiles/travis:v0.1.2-beta node start --home /travis

Upgrade and Continue
---------------------

At certain block heights, the node will stop. Download the next version of the software (e.g., ``0.1.3-beta-hotfix1`` at block height 224550), and restart.

::

  docker stop travis
  docker rm travis
  
  docker pull cybermiles/travis:v0.1.3-beta-hotfix1
  docker run --name travis -v $HOME/.travis:/travis -p 26657:26657 -t cybermiles/travis:v0.1.3-beta-hotfix1 node start --home /travis
  







