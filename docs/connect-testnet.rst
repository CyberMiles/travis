=====
A Network Node (for Docker and Travis built from source)
=====

In this document, we will discuss how to connect to the CyberMiles TestNet and the production MainNet. 

Prerequisite
=====
For Docker: It's assumed that you have `setup docker <https://docs.docker.com/engine/installation/>`_.

For Travis built from source: It's assumed that you have `installed Travis via source builds <http://travis.readthedocs.io/en/latest/getting-started.html#use-docker>`_. (Stop before you connect to a local node)

Docker Image (for Docker)
=====
Docker image for Travis is stored on `Docker Hub <https://hub.docker.com/r/ywonline/travis/tags/>`_. TestNet environment is using the `'lastest' <https://github.com/cybermiles/travis/tree/staging>`_ branch which can be pulled automatically from Travis:

::

  $ docker pull ywonline/travis

Note: Configuration and data will be stored at /travis directory in the container. The directory will also be exposed as a volume. The ports 8545, 26656 and 26657 will be exposed for connection.

Getting Travis TestNet Config (for both)
=====

Checkout the Travis TestNet config from our `Github repo <https://github.com/CyberMiles/testnet>`_. Place the config files in the $HOME/.travis directory:

::

  $ cd
  $ sudo rm -rf .travis
  $ git clone https://github.com/CyberMiles/testnet.git
  $ cd testnet/travis
  $ git pull
  $ cp -r init $HOME/.travis

Start the Node and Join Travis TestNet (for Docker)
=====

Run the docker Travis application:

::

  $ docker run --name travis -v $HOME/.travis:/travis -p 26657:26657 -p 8545:8545 -t ywonline/travis node start --home /travis

Now your node is syncing with TestNet, the output will look like:

::

  INFO [07-20|03:13:26.229] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=1.002ms   mgasps=0.000    number=3363 hash=4884c0…212e75 cache=2.22mB
  I[07-20|03:13:26.241] Committed state                              module=state height=3363 txs=0 appHash=3E0C01B22217A46676897FCF2B91DB7398B34262
  I[07-20|03:13:26.443] Executed block                               module=state height=3364 validTxs=0 invalidTxs=0
  I[07-20|03:13:26.443] Updates to validators                        module=state updates="[{\"address\":\"\",\"pub_key\":\"VPsUJ1Eb73tYPFhNjo/8YIWY9oxbnXyW+BDQsTSci2s=\",\"power\":27065},{\"address\":\"\",\"pub_key\":\"8k17vhQf+IcrmxBiftyccq6AAHAwcVmEr8GCHdTUnv4=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"PoDmSVZ/qUOEuiM38CtZvm2XuNmExR0JkXMM9P9UhLU=\",\"power\":27048},{\"address\":\"\",\"pub_key\":\"2Tl5oI35/+tljgDKzypt44rD1vjVHaWJFTBdVLsmcL4=\",\"power\":27048}]"

To access the TestNet type the following in a seperte terminal console:

first get your IP address then use your IP address to connect to the TestNet

::

  $ docker inspect -f '{{ .NetworkSettings.IPAddress }}' travis
  172.17.0.2:8545
  $ docker run --rm -it ywonline/travis attach http://172.17.0.2:8545

Start the Node and Join Travis TestNet (for Travis built from source)
=====

Run the Travis application:

::

  $ travis node start --home ~/.travis

To access the TestNet type the following in a seperte terminal console:

::

  $ travis attach http://localhost:8545

=====
Connect to MainNet
=====

This section will be completed when the mainnet launches in Q3 2018.
