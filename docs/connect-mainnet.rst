======================
Deploy a MainNet Node
======================

In this document, we will discuss how to start your own node and connect to the CyberMiles MainNet. While we highly recommend you to run your own Travis node, you could still directly access `RPC services <https://travis.readthedocs.io/en/latest/json-rpc.html>`_ from a node provided by the CyberMiles Foundation at ``https://rpc.cybermiles.io:8545``.


********
Snapshot
********

The easiest and fastest way to start a node is to use a snapshot. You can run the node inside a Docker container or on Ubuntu 16.04 / CentOS 7 servers.

Option 1: Docker from snapshot
===============================

Prerequisite
------------

Please `setup docker <https://docs.docker.com/engine/installation/>`_.

Docker Image
------------

Docker image for Travis is stored on `Docker Hub <https://hub.docker.com/r/cybermiles/travis/tags/>`_. MainNet environment is currently at the `'v0.1.8-beta-hotfix' <https://github.com/CyberMiles/travis/releases/tag/v0.1.8-beta-hotfix>`_ release which can be pulled as follows.

::

  docker pull cybermiles/travis:v0.1.8-beta-hotfix

Note: Configuration and data will be stored at ``/travis`` directory in the container. The directory will also be exposed as a volume. The ports 8545, 26656 and 26657 will be exposed for connection.

Getting Travis MainNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  docker run --rm -v $HOME/.travis:/travis -t cybermiles/travis:v0.1.8-beta-hotfix node init --env mainnet --home /travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json

Download snapshot
------------------

Get a list of recent snapshots of the mainnet from AWS S3 `travis-ss-bucket <https://s3-us-west-2.amazonaws.com/travis-ss-bucket/latest.html>`_

You can splice the file name from the bucket list. The downloading url will be like ``https://s3-us-west-2.amazonaws.com/travis-ss-bucket/mainnet/travis_ss_mainnet_1558862782_1724747.tar``. You must have found that the file name contains timestamp and block number at which the snapshot is made.

::

  wget $(curl -s http://s3-us-west-2.amazonaws.com/travis-ss-bucket/latest.html)

Extract the file and copy the ``data`` and ``vm`` subdirectories from the uncompressed directory to ``$HOME/.travis``

Start the Node and Join Travis MainNet
--------------------------------------

Change your name from default name ``local``.

::

  vim $HOME/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

For the security concern, the rpc service is disabled by default, you can enable it by changing the ``config.toml``:

::

  vim $HOME/.travis/config/config.toml
  rpc = true

Run the docker Travis application:

::

  docker run --privileged --name travis -v $HOME/.travis:/travis -t -p 26657:26657 cybermiles/travis:v0.1.8-beta-hotfix node start --home /travis


Attach to the Node and run web3-cmt.js 
---------------------------------------

In another terminal window, log into the Docker container and then run the ``travis`` client and attach to the node. It will open a console to run ``web3-cmt.js`` commands.

::

  docker exec -it travis bash
  > ./travis attach http://localhost:8545

----

Option 2: Binary from snapshot
===============================

**Make sure your os is Ubuntu 16.04 or CentOS 7**

Download snapshot
------------------

Get a list of recent snapshots of the mainnet from AWS S3 `travis-ss-bucket <https://s3-us-west-2.amazonaws.com/travis-ss-bucket/latest.html>`_

You can splice the file name from the bucket list. The downloading url will be like ``https://s3-us-west-2.amazonaws.com/travis-ss-bucket/mainnet/travis_ss_mainnet_1558862782_1724747.tar``. You must have found that the file name contains timestamp and block number at which the snapshot is made.

::

  rm -rf $HOME/.travis
  
  mkdir -p $HOME/release
  cd $HOME/release
  SNAPSHOT_URL=$(curl -s http://s3-us-west-2.amazonaws.com/travis-ss-bucket/latest.html)
  wget $SNAPSHOT_URL
  TAR_FILE="${SNAPSHOT_URL##*/}"
  tar xf $TAR_FILE

  # if your os is Ubuntu 16.04
  mv .travis $HOME
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.8-beta-hotfix/travis_v0.1.8-beta-hotfix_ubuntu-16.04.zip
  unzip travis_v0.1.8-beta-hotfix_ubuntu-16.04.zip
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib
  
  # or if your os is CentOS 7
  mv .travis $HOME
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.8-beta-hotfix/travis_v0.1.8-beta-hotfix_centos-7.zip
  unzip travis_v0.1.8-beta-hotfix_centos-7.zip
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib


Set env variables for eni lib
------------------------------

::

  # for convenience, you should also put these two lines in your .bashrc or .zshrc
  export ENI_LIBRARY_PATH=$HOME/.travis/eni/lib
  export LD_LIBRARY_PATH=$HOME/.travis/eni/lib

Start the Node and Join MainNet
--------------------------------------

Download the mainnet config and change your name from default name ``local``.

::

  mkdir -p $HOME/.travis/config
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  vim $HOME/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"
  
For the security concern, the rpc service is disabled by default, you can enable it by changing the ``config.toml``:

::

  vim $HOME/.travis/config/config.toml
  rpc = true


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
Sync from Genesis
******************

**Expert Only:** You can always start a new CyberMiles blockchain node from genesis, and sync it all the way to the current block height. The process is fairly involved since it requires you to upgrade and restart the node at certain block heights.

One of the key characteristics of the CyberMiles blockchain is the finality of each block. The blockchain will never fork. It will only produce a new block when 2/3 of the validator voting power reach consensus. Software upgrade on the CyberMiles blockchain is done via consensus. That is, at an agreed upon block height, all nodes must upgrade to a new version of the software to continue. Any node that does not upgrade will not reach consensus with the rest of the blockchain and stop.

The table below shows the software version and their corresponding block heights on the mainnet.

====================== ==================== =====================================================================
Blocks                 Software version     Note
====================== ==================== =====================================================================
0 - 230767             v0.1.2-beta          The chain stops itself at 230767
230768 - 386223        v0.1.3-beta-hotfix1  
386224 - 386245        v0.1.3-beta-hotfix2  Mannualy stop the chain within this height range and deploy hotfix2
286246 - 1321175       v0.1.3-beta-hotfix2  The chain stops itself at 1321175
1321176 - 1700580      v0.1.7-beta          The chain stops itself at 1700580
1700581 -              v0.1.8-beta-hotfix   v0.1.8-beta crashes at 1724748
====================== ==================== =====================================================================

The general process for syncing a node from genesis is as follows:

* The 0.1.2-beta software starts from genesis
* It automatically stops at block 230767
* You will download 0.1.3-beta-hotfix1 software, and restart the node
* The process repeats until the block height is current

In the instructions below, we will explain how to switch from one version of the software to the next at specific block heights.

Option 3 (the hard way): Binary from Genesis
=============================================

**Make sure your os is Ubuntu 16.04 or CentOS 7**

Download pre-built binaries
----------------------------

Get software version ``0.1.2-beta`` from from `release page <https://github.com/CyberMiles/travis/releases/>`_

::

  mkdir -p $HOME/release
  cd $HOME/release
  
  # if your os is Ubuntu
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.2-beta/travis_v0.1.2-beta_ubuntu-16.04.zip
  unzip travis_v0.1.2-beta_ubuntu-16.04.zip

  # or if your os is CentOS
  wget https://github.com/CyberMiles/travis/releases/download/v0.1.2-beta/travis_v0.1.2-beta_centos-7.zip
  unzip travis_v0.1.2-beta_centos-7.zip

Getting Travis MainNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  cd $HOME/release
  ./travis node init --env mainnet
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json


Change your name from default name ``local``.

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

Start the Node and Join Travis MainNet
--------------------------------------

::

  cd $HOME/release
  ./travis node start

Upgrade and Continue
---------------------

At certain block heights, the node will stop. Download the next version of the software (e.g., ``0.1.3-beta-hotfix1`` at block height 230767), and restart. A notable exception is the switch between ``0.1.3-beta-hotfix1`` and ``0.1.3-beta-hotfix2`` -- that has to happen manually within a specific range of block heights.

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

Option 4 (the hard way): Docker from Genesis
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

Getting Travis MainNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  docker run --rm -v $HOME/.travis:/travis -t cybermiles/travis:v0.1.2-beta node init --env mainnet --home /travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json

Start the Node and Join MainNet
--------------------------------------
First change your name from default name ``local``.

::

  vim ~/.travis/config/config.toml

  # here you can change your name
  moniker = "<your_custom_name>"
  
Run the docker Travis application:

::

  docker run --privileged --name travis -v $HOME/.travis:/travis -p 26657:26657 -t cybermiles/travis:v0.1.2-beta node start --home /travis

Upgrade and Continue
---------------------

At certain block heights, the node will stop. Download the next version of the software (e.g., ``0.1.3-beta-hotfix1`` at block height 230767), and restart. A notable exception is the switch between ``0.1.3-beta-hotfix1`` and ``0.1.3-beta-hotfix2`` -- that has to happen manually within a specific range of block heights.

::

  docker stop travis
  docker rm travis
  
  docker pull cybermiles/travis:v0.1.3-beta-hotfix1
  docker run --privileged --name travis -v $HOME/.travis:/travis -p 26657:26657 -t cybermiles/travis:v0.1.3-beta-hotfix1 node start --home /travis
  
