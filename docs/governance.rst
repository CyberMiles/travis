===============
Governance
===============

In this document, we will explain governance transactions
to change the system state on the fly.

Send governance transaction using JSON-RPC.

Propose to transfer fund
````````````````````````

.. code:: bash

  curl -s -H "Content-Type: application/json" -d '{ "jsonrpc":"2.0", "method":"cmt_propose", "params": [{ "sequence": 0, "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "transferFrom": "0x77beb894fc9b0ed41231e51f128a347043960a9d", "transferTo": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "amount": "0x3", "reason": "***" }], "id": 1 }'

Compse a proposal to transfer "amount" of CMTs from "transferFrom" to "transferTo".

In the deliver section of the returned data there is a data field stands for the proposal id which should be used for voting on the proposal.

Vote on the proposal
````````````````````

.. code:: bash

  curl -s -H "Content-Type: application/json" -d '{ "jsonrpc":"2.0", "method":"cmt_vote", "params": [{ "proposalId": "***", "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "answer": "Y" }], "id": 1 }'

With the "answer" of "Y" to approve the proposal and "N" to refuse it.

Proposal to change sys param
````````````````````````````

.. code:: bash

  curl -s -H "Content-Type: application/json" -d '{ "jsonrpc":"2.0", "method":"cmt_proposeChangeParam", "params": [{ "sequence": 0, "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "name": "inflation_rate", "value": "8", "reason": "***" }], "id": 1 }'

Compose a proposal to change "inflation_rate" to 8.
