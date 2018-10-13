===============
Governance
===============

In this document, we will explain governance transactions
to change the system state on the fly.

Send governance transaction using JSON-RPC.

cmt_propose
````````````````````````
Propose to transfer fund. Compse a proposal to transfer "amount" of CMTs from "transferFrom" to "transferTo".

**Parameters**

Object - The proposal object

* from: DATA, 20 Bytes - The address the transaction is sent from.
* transferFrom: DATA, 20 Bytes - The address from where the fund should be transfered.
* transferTo: DATA, 20 Bytes - The address where the fund will be transfered to.
* amount: QUANTITY, Integer of the fund value.
* reason: DATA - Why do you compose this proposal.

.. code:: bash

"params": [{
    "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "transferFrom": "0x77beb894fc9b0ed41231e51f128a347043960a9d",
    "transferTo": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "amount": "0x1",
    "reason": "……"
}]

**Returns**

JSON - In the deliver_tx section of the result, there is a data field that stands for the proposal id. It should be used for voting on that proposal.

**Example**

.. code:: bash

// Request
curl -X POST --data '{ "jsonrpc":"2.0", "method":"cmt_propose", "params": [{ "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "transferFrom": "0x77beb894fc9b0ed41231e51f128a347043960a9d", "transferTo": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "amount": "0x1", "reason": "***" }], "id": 1 }'

// Response
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "check_tx": {
            "fee": {}
        },
        "deliver_tx": {
            "data": "7I19v9y7lzFKL3FXqCVU24DELIYOqBSjDcg80+r3yEI=",
            "gasUsed": "2000000",
            "fee": {
                "key": "R2FzRmVl",
                "value": "4000000000000000"
            }
        },
        "hash": "F287844CC7534C9EF2CF67230D51D86952910F22",
        "height": 189
    }
}

cmt_proposeChangeParam
````````````````````````

Compose a proposal to change value of system parameter.

**Parameter**

Object - The proposal object

* from: DATA, 20 Bytes - The address the transaction is sent from.
* name: DATA - Name of the system parameter.
* value: DATA - The new value to be set.
* reason: DATA - Why do you compose this proposal.

.. code:: bash

"params": [{
    "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "name": "inflation_rate",
    "value": "8",
    "reason": "noreason"
}]

**Returns**

JSON - In the deliver_tx section of the result there is a data field that stands for the proposal id. It should be used for voting on that proposal.

**Example**

.. code:: bash

// Request
curl -X POST --data '{ "jsonrpc":"2.0", "method":"cmt_proposeChangeParam", "params": [{ "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "name": "inflation_rate", "value": "8", "reason": "***" }], "id": 1 }'

// Response
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "check_tx": {
            "fee": {}
        },
        "deliver_tx": {
            "data": "CwKl9oFI5STw9wviReF3CQ24joZ+tpWF/vdIqr2rH/c=",
            "gasUsed": "2000000",
            "fee": {
                "key": "R2FzRmVl",
                "value": "4000000000000000"
            }
        },
        "hash": "D83B8DE370DD4D6EBB0A76DB04AE5D66FACE44A4",
        "height": 210
    }
}

cmt_vote
````````````````````````
Vote on the proposal.

**Parameters**


Object - The vote object

* proposalId: Data - The proposal's Id.
* from: DATA, 20 Bytes - The address the transaction is sent from.
* answer: DATA, Y/N - With the "answer" of "Y" to approve the proposal and "N" to refuse it.

.. code:: bash

"params": [{
    "proposalId": "AZcoh+DNcRQu5AgcT8+gKvBW5Bha9tepemWCoa4pw+I=",
    "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "answer": "Y"
}]

**Returns**

JSON

**Example**

.. code:: bash

// Request
curl -X POST --data '{ "jsonrpc":"2.0", "method":"cmt_vote", "params": [{ "proposalId": "CwKl9oFI5STw9wviReF3CQ24joZ+tpWF/vdIqr2rH/c=", "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "answer": "Y" }], "id": 1 }'

// Response
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "check_tx": {
            "fee": {}
        },
        "deliver_tx": {
            "fee": {}
        },
        "hash": "E1D315D2D7207B172BEE838CDACCD771B533090E",
        "height": 212
    }
}
