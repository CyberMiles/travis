===============
Transactions
===============

The CyberMiles blockchain is fully backward compatible with the Ethereum protocol. That means 
all Ethereum transactions are supported on the CyberMiles blockchain. For developers, we recommend you use the
`web3-cmt.js <https://github.com/CyberMiles/web3-cmt.js/>`_ library interact with the blockchain. The ``web3-cmt.js`` library is a customized version of the 
Ethereum `web3.js <https://github.com/ethereum/web3.js/>`_ library, with the ``eth`` module renamed to the ``cmt`` module. 
The ``web3-cmt.js`` library is integrated into the ``travis`` client console by default.

..
  // send a transfer transaction
  web3.cmt.sendTransaction(
    {
      from: "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe",
      to: "0x11f4d0A3c12e86B4b5F39B213F7E19D048276DAe",
      value: web3.toWei(100, "cmt")
    },
    (err, res) => {
      // ...
    }
  )
  
  // get the balance of an address
  var balance = web3.cmt.getBalance("0x11f4d0A3c12e86B4b5F39B213F7E19D048276DAe")


Beyond Ethereum, however, the most important transactions that are specific for the CyberMiles blockchain are for
DPoS staking operations and for blockchain governance operations.

Staking transactions
-------- 

A key characteristic of the CyberMiles blockchain is the DPoS consensus mechanism. You can read more about the 
`CyberMiles DPoS protocol <https://www.cybermiles.io/validator>`_. With the staking transactions, CMT holders
can declare candidacy for validators, stake and vote for candidates, and unstake as needed. Learn more about the
`staking transactions for validators <https://cybermiles.github.io/web3-cmt.js/api/validator.html>`_ and the 
`staking transactions for delegators <https://cybermiles.github.io/web3-cmt.js/api/delegator.html>`_.


Governance transactions
-------- 

With the DPoS consensus mechanism, the CyberMiles validators can make changes to the blockchain network's
key parameters, deploy new `libENI libraries <https://www.litylang.org/performance/>`_, 
create `trusted contracts <https://www.litylang.org/trusted/>`_, and make other policy changes. Anyone on the blockchain
can propose governance changes, but only the current validators can vote. Learn more about the
`governance transactions <https://cybermiles.github.io/web3-cmt.js/api/governance.html>`_.




