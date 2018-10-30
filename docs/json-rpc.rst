======
Travis JSON-RPC
======

Travis is compatible with Ethereum so most methods especially the ones for normal transaction are the same with Ethereum. Please refer to Ethereum `JSON-RPC <https://github.com/ethereum/wiki/wiki/JSON-RPC>`_ for more information.

CMT methods
===========

cmt_syncing
-----------

Returns the sync object.

**Parameters**

	none

**Returns**

	* ``latest_block_hash`` Number - The hash of the latest block.
	* ``latest_app_hash`` Number - The hash of latest application state.
	* ``latest_block_height`` Number - The latest block number.
	* ``latest_block_time`` Number - The latest block time.
	* ``catching_up`` Boolean - Whether the node is syncing or not.

**Example**

::

  // Request
  curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_syncing","params":[],"id":1}'

  // Result
  {
  	"jsonrpc": "2.0",
  	"id": 1,
  	"result": {
  		"latest_block_hash": "94C0363F68AD5184A861FAE0010BE0D44FDD3254",
  		"latest_app_hash": "BB510006FDB4A907A3C7BEAA4A8A2F493252DDCD",
  		"latest_block_height": 115851,
  		"latest_block_time": "2018-10-30T04:58:17.895717492Z",
  		"catching_up": false
  	}
  }


