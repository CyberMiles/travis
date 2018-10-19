=====
Deploy a MainNet Node
=====

In this document, we will discuss how to connect to the CyberMiles Travis MainNet. We will cover binary, Docker and "build from source" scenarios. If you are new to CyberMiles, deploying a Docker node is probably easier.

While we highly recommend you to run your own Travis node, you can also ask for direct access to one of the nodes maintained by the CyberMiles Foundation. Send an email to travis@cybermiles.io to apply for access credentials. You still need the ``travis`` client either from Docker or source to access the node.

Binary
======

Make sure your os is Ubuntu 16.04 or CentOS 7

Download pre-built binaries from `release page <https://github.com/CyberMiles/travis/releases>`_
-----------------------------------------------------------------------------------------------------------

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
  mkdir -p $HOME/.travis
  cd $HOME/release

  ./travis node init --env mainnet --home $HOME/.travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json

Change your name from default name ``local``, set persistent peers

::

  cd $HOME/.travis
  vim $HOME/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

  # find the seeds option and change its value
  seeds = "7e7a321470a6a437457293b32eeeddf0cb498f93@35.185.101.6:26656,607dd907c7e60f4884cd26001cf0d97fed227123@35.229.126.251:26656"

Copy libeni into the default Travis data directory
--------------------------------------------------

::

  mkdir -p $HOME/.travis/eni
  cp -r $HOME/release/lib/. $HOME/.travis/eni/lib
  
  # set env variables for eni lib
  export ENI_LIBRARY_PATH=$HOME/.travis/eni/lib
  export LD_LIBRARY_PATH=$HOME/.travis/eni/lib

Start the Node and Join Travis MainNet
--------------------------------------

::

  cd $HOME/release
  ./travis node start --home $HOME/.travis


Docker
======

Prerequisite
------------
Please `setup docker <https://docs.docker.com/engine/installation/>`_.

Docker Image
------------
Docker image for Travis is stored on `Docker Hub <https://hub.docker.com/r/cybermiles/travis/tags/>`_. MainNet environment is using the `'v0.1.2-beta' <https://github.com/CyberMiles/travis/releases/tag/v0.1.2-beta>`_ branch which can be pulled automatically from Travis:

::

  docker pull cybermiles/travis:v0.1.2-beta

Note: Configuration and data will be stored at /travis directory in the container. The directory will also be exposed as a volume. The ports 8545, 26656 and 26657 will be exposed for connection.

Getting Travis MainNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  docker run --rm -v $HOME/.travis:/travis cybermiles/travis:v0.1.2-beta node init --env mainnet --home /travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json

Start the Node and Join Travis MainNet
--------------------------------------
First change your name from default name ``local``, set persistent peers

::

  vim ~/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

  # find the seeds option and change its value
  seeds = "7e7a321470a6a437457293b32eeeddf0cb498f93@35.185.101.6:26656,607dd907c7e60f4884cd26001cf0d97fed227123@35.229.126.251:26656"

Run the docker Travis application:

::

  docker run --name travis -v $HOME/.travis:/travis -t -p 26657:26657 cybermiles/travis:v0.1.2-beta node start --home /travis


Build from source
=================

Prerequisite
------------
Please `install Travis via source builds <http://travis.readthedocs.io/en/latest/getting-started.html#build-from-source>`_. (STOP before you connect to a local node)

Getting Travis MainNet Config
-----------------------------

::

  rm -rf $HOME/.travis
  mkdir -p $HOME/.travis
  cd $HOME/release

  ./travis node init --env mainnet --home $HOME/.travis
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/config.toml > $HOME/.travis/config/config.toml
  curl https://raw.githubusercontent.com/CyberMiles/testnet/master/travis/init-mainnet/genesis.json > $HOME/.travis/config/genesis.json

Change your name from default name ``local``, set persistent peers

::

  cd $HOME/.travis
  vim $HOME/.travis/config/config.toml
  # here you can change your name
  moniker = "<your_custom_name>"

  # find the seeds option and change its value
  seeds = "7e7a321470a6a437457293b32eeeddf0cb498f93@35.185.101.6:26656,607dd907c7e60f4884cd26001cf0d97fed227123@35.229.126.251:26656"

Start the Node and Join Travis MainNet
--------------------------------------
Run the Travis application:

::

  travis node start --home ~/.travis


Access the MainNet
==================

For the security concern, the rpc service is disabled by default, you can enable it by changing the config.toml:

::

  vim $HOME/.travis/config/config.toml
  rpc = true

Then restart travis service and type the following in a seperte terminal console (make sure that the seperate console also has travis environment):

::

  travis attach http://localhost:8545


You should now the see the web3-cmt JavaScript console and have fun with MainNet..
