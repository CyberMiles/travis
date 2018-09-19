===============
Getting Started
===============

In this document, we will discuss how to create and run a single node CyberMiles blockchain on your computer. 
It allows you to connect and test basic features such as coin transactions, staking and unstaking for validators, 
governance, and smart contracts.


Use Docker
----------------------------

The easiest way to get started is to use our pre-build Docker images. Please make sure that you have 
`Docker installed <https://docs.docker.com/install/>`_ and that your `Docker can work without sudo <https://docs.docker.com/install/linux/linux-postinstall/>`_.

Initialize
``````````

Let’s initialize a docker image for the Travis build first.

.. code:: bash

  docker run --rm -v ~/volumes/local:/travis ywonline/travis node init --home /travis

The node’s data directory is ``~/volumes/local`` on the local computer. 

Run
```

Now you can start the CyberMiles Travis node in docker.

.. code:: bash

  docker run --name travis -v ~/volumes/local:/travis -t -p 26657:26657 -p 8545:8545 ywonline/travis node start --home /travis

At this point, you can Ctrl-C to exit to the terminal and travis will remain running in the background. 
You can check the CyberMiles Travis node’s logs at anytime via the following docker command.

.. code:: bash

  docker logs -f travis

You should see blocks like the following in the log.

.. code:: bash

  INFO [07-14|07:23:05] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=431.085µs mgasps=0.000 number=163 hash=05e16c…a06228
  INFO [07-14|07:23:15] Imported new chain segment               blocks=1 txs=0 mgas=0.000 elapsed=461.465µs mgasps=0.000 number=164 hash=933b97…0c340c

Connect
```````

You can connect to the local CyberMiles node by attaching an instance of the Travis client.

.. code:: bash

  # Get the IP address of the travis node
  docker inspect -f '{{ .NetworkSettings.IPAddress }}' travis
  172.17.0.2

  # Use the IP address from above to connect
  docker run --rm -it ywonline/travis attach http://172.17.0.2:8545

It opens the web3-cmt JavaScript console to interact with the virtual machine. The example below shows how to unlock the
coinbase account so that you have coins to spend.

.. code:: bash

  Welcome to the Travis JavaScript console!

  instance: vm/v1.6.7-stable/linux-amd64/go1.9.3
  coinbase: 0x7eff122b94897ea5b0e2a9abf47b86337fafebdc
  at block: 231 (Sat, 14 Jul 2018 07:34:25 UTC)
   datadir: /travis
   modules: admin:1.0 cmt:1.0 eth:1.0 net:1.0 personal:1.0 rpc:1.0 web3:1.0
  
  > personal.unlockAccount('0x7eff122b94897ea5b0e2a9abf47b86337fafebdc', '1234')
  true
  > 

Build from source
----------------------------

Currently, we only support source builds for CentOS 7 and Ubuntu 16.04 linux distributions.

Prerequisite
````````````

You must have GO language version 1.10+ installed in order to build and run a Travis node. 
The easiest way to get GO 1.10 is through the GVM. Below are the commands on a Linux server.

.. code:: bash

  $ bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
  $ vim ~/.bash_profile
  inset into the bash profile: source "$HOME/.bashrc"
  log out and log in
  $ sudo apt-get install bison
  $ gvm version
  output should look like: Go Version Manager v1.0.22 installed at /home/myuser/.gvm
  $ gvm install go1.10.3
  $ gvm use go1.10.3 --default


Build
`````

First we need to checkout the correct branch of Travis from Github:

.. code:: bash

  go get github.com/CyberMiles/travis (ignore if an error occur)
  cd $GOPATH/src/github.com/CyberMiles/travis
  git checkout master

Next, we need to build libENI and put it into the default Travis data directory ``~/.travis/``.

.. code:: bash

  sudo rm -rf ~/.travis
  wget -O $HOME/libeni.tgz https://github.com/CyberMiles/libeni/releases/download/v1.3.2/libeni-1.3.2_ubuntu-16.04.tgz
  tar zxvf $HOME/libeni.tgz -C $HOME
  mkdir -p $HOME/.travis/eni
  cp -r $HOME/libeni-1.3.2/lib $HOME/.travis/eni/lib

Currently libENI can only run on Ubuntu 16.04 and CentOS 7. If your operating system is CentOS, please change the downloading url. You can find it here: `https://github.com/CyberMiles/libeni/releases <https://github.com/CyberMiles/libeni/releases>`_

Now, we can build and install Travis binary. It will populate additional configuration files into ``~/.travis/``

.. code:: bash

  cd $GOPATH/src/github.com/CyberMiles/travis
  make all

If the system cannot find glide at the last step, make sure that you have ``$GOPATH/bin`` under the ``$PATH`` variable.

Run
```

Let's start a  Travis node locally using the ``~/.travis/`` data directory.

.. code:: bash

  travis node init
  travis node start

Connect
```````

You can connect to the local CyberMiles node by attaching an instance of the Travis client.

.. code:: bash

  travis attach http://localhost:8545

It opens the web3-cmt JavaScript console to interact with the virtual machine. The example below shows how to unlock the
coinbase account so that you have coins to spend.

.. code:: bash

  Welcome to the Travis JavaScript console!

  instance: vm/v1.6.7-stable/linux-amd64/go1.9.3
  coinbase: 0x7eff122b94897ea5b0e2a9abf47b86337fafebdc
  at block: 231 (Sat, 14 Jul 2018 07:34:25 UTC)
   datadir: /travis
   modules: admin:1.0 cmt:1.0 eth:1.0 net:1.0 personal:1.0 rpc:1.0 web3:1.0
  
  > personal.unlockAccount('0x7eff122b94897ea5b0e2a9abf47b86337fafebdc', '1234')
  true
  > 

Test transactions
----------------------------

You can now send a transaction between accounts like the following.

.. code:: bash

  personal.unlockAccount("from_address")
  cmt.sendTransaction({"from": "from_address", "to": "to_address", "value": web3.toWei(0.001, "cmt")})

Next, you can paste the following script into the Travis client console, at the > prompt.

.. code:: bash

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
  
You can now run the script in the console, and see the results.

.. code:: bash

  > checkAllBalances();
  cmt.accounts[0]: 	0x6....................................230 	balance: 466.798526 CMT
  cmt.accounts[1]: 	0x6....................................244 	balance: 1531 CMT
  Total balance: 1997.798526CMT
  
 
 
