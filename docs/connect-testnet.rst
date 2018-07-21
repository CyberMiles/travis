===============
A Network Node
===============

In this document, we will discuss how to connect to the CyberMiles
testnet and the production mainnet.

We assume that you have already built and ran your CyberMiles Travis node
as described in :doc:`Getting Started <getting-started>`.

Connect to testnet
----------------------------

First, stop the local node. You can simply kill the process. Next, checkout the Travis TestNet config from our Github repo.

.. code:: bash

  git clone https://github.com/CyberMiles/testnet.git

Next, you should replace ``config`` and ``vm`` in the local node's home directory with the ones you just checked out from Github.

Docker
``````

If you are running a Docker node, you can do the following to replace the directory content since the local ``~/volumes/local`` directory is mapped to ``/travis`` on Docker, which is the Docker node's home directory.

.. code:: bash

  rm -rf ~/volumes/local/config
  rm -rf ~/volumes/local/vm
  cp -r testnet/travis/init/config ~/volumes/local/
  cp -r testnet/travis/init/vm ~/volumes/local/
  
Start the node again, and wait for it to sync to the testnet.

Compiled from source
````````````````````

If you are running a node compiled from source, you can do the following to replace the directory content assuming the node's home directory is the default ``~/.travis``.

.. code:: bash

  rm -rf ~/.travis/config
  rm -rf ~/.travis/vm
  cp -r testnet/travis/init/config ~/.travis/
  cp -r testnet/travis/init/vm ~/.travis/
  
Start the node again, and wait for it to sync to the testnet.


Connect to mainnet
----------------------------

This section will be completed when the mainnet launches in Q3 2018.


