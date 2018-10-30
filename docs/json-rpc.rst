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

cmt_getBlockByNumber
--------------------

Returns a block matching the block number.

**Parameters**

	* ``blockNumber`` Number - The block number.

**Returns**

	The block object.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_getBlockByNumber","params":[78],"id":1}'

	// Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"block_meta": {
				"block_id": {
					"hash": "E0F9C6439B41E1B80E4D2C4C96EDFD100B4BAEC7",
					"parts": {
						"total": 1,
						"hash": "C78D31D2B57749A3C67EC8F04A6A9DF396365588"
					}
				},
				"header": {
					"chain_id": "CyberMiles",
					"height": 78,
					"time": "2018-10-15T13:41:41.109630547Z",
					"num_txs": 0,
					"last_block_id": {
						"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
						"parts": {
							"total": 1,
							"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
						}
					},
					"total_txs": 0,
					"last_commit_hash": "E05893F0935E1BC259514BC44B61DD8E8962BE8A",
					"data_hash": "",
					"validators_hash": "3760C3CD67AC9A819AF01747476E1B04DABCD05B",
					"consensus_hash": "D6B74BB35BDFFD8392340F2A379173548AE188FE",
					"app_hash": "2144AC53826041B1406CB6B8ABEDC37064211CA5",
					"last_results_hash": "",
					"evidence_hash": ""
				}
			},
			"block": {
				"header": {
					"chain_id": "CyberMiles",
					"height": 78,
					"time": "2018-10-15T13:41:41.109630547Z",
					"num_txs": 0,
					"last_block_id": {
						"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
						"parts": {
							"total": 1,
							"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
						}
					},
					"total_txs": 0,
					"last_commit_hash": "E05893F0935E1BC259514BC44B61DD8E8962BE8A",
					"data_hash": "",
					"validators_hash": "3760C3CD67AC9A819AF01747476E1B04DABCD05B",
					"consensus_hash": "D6B74BB35BDFFD8392340F2A379173548AE188FE",
					"app_hash": "2144AC53826041B1406CB6B8ABEDC37064211CA5",
					"last_results_hash": "",
					"evidence_hash": ""
				},
				"data": {
					"txs": null
				},
				"evidence": {
					"evidence": null
				},
				"last_commit": {
					"block_id": {
						"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
						"parts": {
							"total": 1,
							"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
						}
					},
					"precommits": [{
						"validator_address": "04A515F3B6B9E7FC7E2B5AAC4304D82BE9D6573C",
						"validator_index": 0,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.824095471Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [246, 31, 74, 206, 79, 252, 63, 8, 62, 221, 28, 28, 174, 45, 191, 121, 163, 69, 96, 83, 245, 141, 165, 145, 28, 240, 248, 236, 42, 14, 180, 184, 194, 78, 146, 10, 24, 193, 243, 43, 50, 166, 7, 159, 99, 23, 155, 56, 35, 167, 152, 4, 86, 107, 14, 51, 9, 203, 38, 149, 248, 147, 226, 7]
					}]
				}
			}
		}
	}

cmt_getTransactionByHash
------------------------

Returns a transaction matching the given transaction hash.

**Parameters**

	* ``transactionHash`` String - The transaction hash.

**Returns**

	The transaction object.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_getTransactionByHash","params":["1F64261396674A1A7328B250EC3043E5512010D8"],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"blockNumber": "0x1c6f6",
			"from": "0x245323885234fd5adc48ffb546a54c5df99e9ace",
			"gas": "0x0",
			"gasPrice": "0x0",
			"hash": "0xa73b917243b5d3fb810dfb5f1880daab71564aafbb183c5f1e1f40665832aad5",
			"cmtHash": "1F64261396674A1A7328B250EC3043E5512010D8",
			"input": "0x7b2274797065223a227374616b655c2f64656c6567617465222c2264617461223a7b2276616c696461746f725f61646472657373223a22307846394664333937343836414335353136656561323330346641373031634239373637633436354432222c22616d6f756e74223a223334333731303030303030303030303030303030303030222c22637562655f6261746368223a223032222c22736967223a2232356338393665316235303563643238626463633236656539306439333465356361313135383532663230393737356635636434336230636166393665613134643939623633653034343830383764353236383438313739626165626433616430353366643832663661386530626536326537326161366438633462316435303238623166383663656432353539363832376566623237393461346431343835306533383238653138336635623466326636383336303034666336303863323264353262326464323336336632343339633531623930373235613430613962653562623264323830376164356335636435383237623264643738366431623236227d7d",
			"cmtInput": {
				"type": "stake/delegate",
				"data": {
					"validator_address": "0xf9fd397486ac5516eea2304fa701cb9767c465d2",
					"amount": "34371000000000000000000",
					"cube_batch": "02",
					"sig": "25c896e1b505cd28bdcc26ee90d934e5ca115852f209775f5cd43b0caf96ea14d99b63e0448087d526848179baebd3ad053fd82f6a8e0be62e72aa6d8c4b1d5028b1f86ced25596827efb2794a4d14850e3828e183f5b4f2f6836004fc608c22d52b2dd2363f2439c51b90725a40a9be5bb2d2807ad5c5cd5827b2dd786d1b26"
				}
			},
			"nonce": "0x0",
			"to": null,
			"transactionIndex": "0x0",
			"value": "0x0",
			"v": "0x48",
			"r": "0x224015941f4373e5aee27a1173b9ae112317dfdc3b2a1a86cf557c2446c255e4",
			"s": "0x2798b6ab9f403b938fea0b640476de20a6f09d1e12f86f0cd5e18369164e56ef",
			"txResult": {
				"fee": {}
			}
		}
	}

cmt_getTransactionFromBlock
------------------------

Returns a transaction based on a block hash or number and the transactions index position

**Parameters**

	* ``blockNumber`` Number - The block number.
	* ``indexNumber`` Number - The transactions index position.

**Returns**

	The transaction object.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_getTransactionFromBlock","params":[116470, 0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"blockNumber": "0x1c6f6",
			"from": "0x245323885234fd5adc48ffb546a54c5df99e9ace",
			"gas": "0x0",
			"gasPrice": "0x0",
			"hash": "0xa73b917243b5d3fb810dfb5f1880daab71564aafbb183c5f1e1f40665832aad5",
			"cmtHash": "1F64261396674A1A7328B250EC3043E5512010D8",
			"input": "0x7b2274797065223a227374616b655c2f64656c6567617465222c2264617461223a7b2276616c696461746f725f61646472657373223a22307846394664333937343836414335353136656561323330346641373031634239373637633436354432222c22616d6f756e74223a223334333731303030303030303030303030303030303030222c22637562655f6261746368223a223032222c22736967223a2232356338393665316235303563643238626463633236656539306439333465356361313135383532663230393737356635636434336230636166393665613134643939623633653034343830383764353236383438313739626165626433616430353366643832663661386530626536326537326161366438633462316435303238623166383663656432353539363832376566623237393461346431343835306533383238653138336635623466326636383336303034666336303863323264353262326464323336336632343339633531623930373235613430613962653562623264323830376164356335636435383237623264643738366431623236227d7d",
			"cmtInput": {
				"type": "stake/delegate",
				"data": {
					"validator_address": "0xf9fd397486ac5516eea2304fa701cb9767c465d2",
					"amount": "34371000000000000000000",
					"cube_batch": "02",
					"sig": "25c896e1b505cd28bdcc26ee90d934e5ca115852f209775f5cd43b0caf96ea14d99b63e0448087d526848179baebd3ad053fd82f6a8e0be62e72aa6d8c4b1d5028b1f86ced25596827efb2794a4d14850e3828e183f5b4f2f6836004fc608c22d52b2dd2363f2439c51b90725a40a9be5bb2d2807ad5c5cd5827b2dd786d1b26"
				}
			},
			"nonce": "0x0",
			"to": null,
			"transactionIndex": "0x0",
			"value": "0x0",
			"v": "0x48",
			"r": "0x224015941f4373e5aee27a1173b9ae112317dfdc3b2a1a86cf557c2446c255e4",
			"s": "0x2798b6ab9f403b938fea0b640476de20a6f09d1e12f86f0cd5e18369164e56ef",
			"txResult": {
				"fee": {}
			}
		}
	}


Stake Validator methods
=======================

cmt_declareCandidacy
--------------------

Allows a potential validator declares its candidacy.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. It will be associated with this validator (for self-staking and in order to get paid).
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``pubKey`` String - Validator node public key.
	* ``maxAmount`` String - Max amount of CMTs in Wei to be staked.
	* ``compRate`` String - Validator compensation. That is the percentage of block awards to be distributed back to the validators.
	* ``description`` Object - (optional) Description object as follows:
		* ``name`` String - Validator name.
		* ``website`` String - Web page link.
		* ``location`` String - Location(network and geo).
		* ``email`` String - Email.
		* ``profile`` String - Detailed description.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_declareCandidacy","params":[{"from":"0xc4abd0339eb8d57087278718986382264244252f", "pubKey":"051FUvSNJmVL4UiFL7ucBr3TnGqG6a5JgUIgKf4UOIA=", "maxAmount":"0xF4240", "compRate":"0.2"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				"gasUsed": "1000000",
				"fee": {
					"key": "R2FzRmVl",
					"value": "2000000000000000"
				}
			},
			hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
			height: 271
		}
	}

cmt_updateCandidacy
-------------------

Allows a validator candidate to change its candidacy.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``pubKey`` String - (optional) Validator node public key.
	* ``maxAmount`` String - (optional) New max amount of CMTs in Wei to be staked.
	* ``compRate`` String - (optional) Validator compensation. That is the percentage of block awards to be distributed back to the validators.
	* ``description`` Object - (optional) When updated, the verified status will set to false:
		* ``name`` String - Validator name.
		* ``website`` String - Web page link.
		* ``location`` String - Location(network and geo).
		* ``email`` String - Email.
		* ``profile`` String - Detailed description.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_updateCandidacy","params":[{"from":"0xc4abd0339eb8d57087278718986382264244252f", "maxAmount":"0xF4240", "description": {"website": "https://www.example.com"}}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				"gasUsed": "1000000",
				"fee": {
					"key": "R2FzRmVl",
					"value": "2000000000000000"
				}
			},
			hash: '1B11C4D5EA9664DB1DD3A9CDD86741D6C8E226E9',
			height: 297
		}
	}

cmt_withdrawCandidacy
---------------------

Allows a validator to withdraw.

**Parameters**

	* ``from`` String - The address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_withdrawCandidacy","params":[{"from":"0xc4abd0339eb8d57087278718986382264244252f"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: '4A723894821166EFC7DDD4FD92BE8D855B3FDBAC',
			height: 311
		}
	}

cmt_verifyCandidacy
-------------------

Allows the foundation to "verify" a validator's information.

**Parameters**

	* ``from`` String - A special address the foundation owns. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``candidateAddress`` String - The address of validator to verfify.
	* ``verified`` Boolean - (optional) Verified true or false, default to false.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_verifyCandidacy","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "candidateAddress":"0xc4abd0339eb8d57087278718986382264244252f", "verified":true}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: 'EADC546C764AFF6C176B843321B5AB090FBEC0DA',
			height: 334
		}
	}

cmt_activateCandidacy
---------------------

Allows a "removed" validator to re-activate itself.

**Parameters**

	* ``from`` String - The address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_activateCandidacy","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: 'FB70A78AD62A0E0B24194CA951725770B2EFBC0A',
			height: 393
		}
	}

cmt_deactivateCandidacy
-----------------------

Allows a validator to deactivate itself. 

**Parameters**

	* ``from`` String - The address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_deactivateCandidacy","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: 'FB70A78AD62A0E0B24194CA951725770B2EFBC0A',
			height: 393
		}
	}


cmt_setCompRate
---------------

Allows a validator to update the compensation rate for its delegators.

**Parameters**

	* ``from`` String - The address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``delegatorAddress`` String - The adddress of delegator.
	* ``compRate`` String - New compensation rate to set for the delegator. Compensation rate is the percentage of block awards to be distributed back to the validators.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_setCompRate","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "delegatorAddress":"0x38d7b32e7b5056b297baf1a1e950abbaa19ce949", "compRate":"0.3"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				"gasUsed": "1000000",
				"fee": {
					"key": "R2FzRmVl",
					"value": "2000000000000000"
				}
			},
			hash: 'C61BAEEEF637CB554157261DF27F7D1CFE50F251',
			height: 393
		}
	}

cmt_updateCandidacyAccount
--------------------------

A validator requests to update its binding address.

**Parameters**

	* ``from`` String - The address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``newCandidateAccount`` String - The new adddress of the validator.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed. If successful, the requestId will be set in the data property(base64 encoded), for the new address to accept later.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_updateCandidacyAccount","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "newCandidateAccount":"0x283ED77f880D87dBdE8721259F80517A38ae5b4f"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				data: "MQ==",
				gasUsed: "1000000",
				fee: {
					key: "R2FzRmVl",
					value: "2000000000000000"
				}
			},
			hash: "34B157D42AFF2D8327FC8CEA8DFFC1E61E9C0D93",
			height: 105
		}
	}

cmt_acceptCandidacyAccountUpdate
--------------------------------

A validator uses its new address to accept an account updating request.

**Parameters**

	* ``from`` String - The new address for the validator. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``accountUpdateRequestId`` int64 - The account updating request id.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_acceptCandidacyAccountUpdate","params":[{"from":"0x283ed77f880d87dbde8721259f80517a38ae5b4f", "accountUpdateRequestId":1}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				gasUsed: "1000000",
				fee: {
					key: "R2FzRmVl",
					value: "2000000000000000"
				}
			},
			hash: "D343D115C152D1A78B7DB9CAA2160E3BA31A3F63",
			height: 67
		}
	}

cmt_queryValidator
------------------

Query the current stake status of a specific validator.

**Parameters**

	* ``validatorAddress`` String - The validator address.
	* ``height`` Number - The block number. Default to 0, means current head of the blockchain. NOT IMPLEMENTED YET.

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Object - The validator object.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryValidator","params":["0x858578e81a0259338b4d897553afa7b9c363e769", 0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"jsonrpc": "2.0",
			"id": 1,
			"result": {
				"height": 116992,
				"data": {
					"id": 29,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "NEUjQM4EvOkTruH0aufgQM4tLEKrCJSAvEDKwZ771ng="
					},
					"owner_address": "0x858578e81a0259338b4d897553aFA7b9c363e769",
					"shares": "2098954378147353283849105",
					"voting_power": 161882,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539619422,
					"description": {
						"name": "Rfinex",
						"website": "https://www.rfinex.com",
						"location": "Geneva, Switzerland",
						"email": "",
						"profile": "Make Crypto Greater"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 881,
					"rank": 15,
					"state": "Validator",
					"num_of_delegators": 2
				}
			}
		}
	}

cmt_queryValidators
-------------------

Returns a list of all current validators and validator candidates.

**Parameters**

	* ``height`` Number - The block number. Default to 0, means current head of the blockchain. NOT IMPLEMENTED YET.

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Array - An array of all current validators and validator candidates.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryValidators","params":[0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"jsonrpc": "2.0",
			"id": 1,
			"result": {
				"height": 117008,
				"data": [{
					"id": 9,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "aIVtdAdQlQ4uuTMmsU+8z9d//+URrPKX2vcobWDO6HA="
					},
					"owner_address": "0x5c158B32dE3037d5BC6D2Ebff1b9cF099daF1F7D",
					"shares": "525829971878780385668796",
					"voting_power": 48854,
					"pending_voting_power": 0,
					"max_shares": "5000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 0,
					"description": {
						"name": "Seed Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 0,
					"rank": 18,
					"state": "Validator",
					"num_of_delegators": 10
				}]
			}
		}
	}

cmt_queryAwardInfos
-------------------

Returns award information of all current validators and backup validators.

**Parameters**

	* ``height`` Number - The block number. Default to 0, means current head of the blockchain.

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Array - An array of award information of all current validators and backup validators.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryAwardInfos","params":[0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"jsonrpc": "2.0",
			"id": 1,
			"result": {
				"height": 117024,
				"data": [{
					"address": "0x1ac7d4f1d4fa3eaef67d8208a2b1b84670211e75",
					"state": "Validator",
					"amount": "695443350547290695"
				}, {
					"address": "0x1724d4a82f29d93a1eb96c19b4bb6b219dc18f23",
					"state": "Validator",
					"amount": "1221147599113753229"
				}, {
					"address": "0x4cdaf011cadba6c3997252738e4d6dd30c8865b9",
					"state": "Validator",
					"amount": "551560392982316129"
				}, {
					"address": "0xfd0e8e4c4dea053f10e72e8800b08ac875e5ac49",
					"state": "Validator",
					"amount": "624496861299062028"
				}, {
					"address": "0x8958618332df62af93053cb9c535e26462c959b0",
					"state": "Validator",
					"amount": "1043204078311188515"
				}, {
					"address": "0x5c158b32de3037d5bc6d2ebff1b9cf099daf1f7d",
					"state": "Validator",
					"amount": "165901770329933149"
				}, {
					"address": "0x1b92c5bb82972af385d8cd8c1230502083898ba6",
					"state": "Validator",
					"amount": "3773502902478876527"
				}, {
					"address": "0xcd3090e881170f6d036fdb3ae5a3d36ead5bcf83",
					"state": "Validator",
					"amount": "1251133119890215962"
				}, {
					"address": "0x3af427d092f9bf934d2127408935c1455170ea8a",
					"state": "Validator",
					"amount": "837444996842711426"
				}, {
					"address": "0x70a52ff393256f016939ae2926cbd999508a555b",
					"state": "Validator",
					"amount": "1129557623931196482"
				}, {
					"address": "0x34c5f1c0e10701dbaf0df1ad2a7826be41a3a380",
					"state": "Validator",
					"amount": "3188316617454823732"
				}, {
					"address": "0xeb65290b802df113300120c52b313f1896e80d38",
					"state": "Validator",
					"amount": "673088346779289101"
				}, {
					"address": "0xf9fd397486ac5516eea2304fa701cb9767c465d2",
					"state": "Validator",
					"amount": "696852636065097266"
				}, {
					"address": "0xf9a431660dc8e425018564ce707d44a457301eb9",
					"state": "Validator",
					"amount": "538679862936435824"
				}, {
					"address": "0x9a3482fd81d706d5aa941f38946af69a448e08c3",
					"state": "Validator",
					"amount": "1780039672657381152"
				}, {
					"address": "0x858578e81a0259338b4d897553afa7b9c363e769",
					"state": "Validator",
					"amount": "549733415612243995"
				}, {
					"address": "0x0da518ecf4761a86965c1f77ac4c1bd6e19904e3",
					"state": "Validator",
					"amount": "654220900184269072"
				}, {
					"address": "0x221507f21aac826263a664538580e57ded401978",
					"state": "Validator",
					"amount": "3022099031334153293"
				}, {
					"address": "0x654e1dfe66519b9a09305ad58392d9a1c61296b3",
					"state": "Validator",
					"amount": "434627049560264340"
				}, {
					"address": "0x482a7cbb8f66a9db6b25808861b182c670c79259",
					"state": "Backup Validator",
					"amount": "460623994544478727"
				}, {
					"address": "0x04ba6cf9a4035294958678dd0f540a195b260d0e",
					"state": "Backup Validator",
					"amount": "577538694065296547"
				}, {
					"address": "0xe218509490578f75dfc6ed6c8a80158675071a8c",
					"state": "Backup Validator",
					"amount": "577517942432438752"
				}, {
					"address": "0xae3befdc5d0f5397b9e448fe136f10360dddde28",
					"state": "Backup Validator",
					"amount": "461557818023079509"
				}, {
					"address": "0x72cf924c62baff2ed74a5ceb885082b814216e55",
					"state": "Backup Validator",
					"amount": "459544909635873380"
				}]
			}
		}
	}


Stake Delegator methods
=======================

cmt_delegate
------------

Used by a delegator to stake CMTs to a validator.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``validatorAddress`` String - The address of validator to delegate.
	* ``amount`` String - Amount of CMTs in Wei to delegate.
	* ``cubeBatch`` String - The batch number of the CMT cube. Use "01" for testing.
	* ``sig`` String - delegator_address|nonce signed by the CMT cube. Check this for how to generate a signature for testing.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_delegate","params":[{"from":"0x38d7b32e7b5056b297baf1a1e950abbaa19ce949", "validatorAddress":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "amount":"0x186A0", "cubeBatch":"01", "sig":"036b6dddefdb1d798a9847121dde8c38713721869a24c77abe2249534f6d98622727720994f663ee9cc446c6e246781caa3a88b7bff78a4ffc9de7c7eded00caef61c2ea36be6a0763ed2bf5af4cf38e38bd6b257857f314c4bbb902d83c8b4413ba2f880d24bf0d6874e392807dfbc2bd03910c58989bc69a9090eddefe8e55"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: '8A40C44D31316BFB2D417A1985E03DA36145EF5A',
			height: 319
		}
	}

cmt_withdraw
------------

Used by a delegator to unbind staked CMTs from a validator.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``validatorAddress`` String - The address of validator to withdraw.
	* ``amount`` String - Amount of CMTs in Wei to withdraw.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_withdraw","params":[{"from":"0x38d7b32e7b5056b297baf1a1e950abbaa19ce949", "validatorAddress":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "amount":"0x186A0"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: '8A40C44D31316BFB2D417A1985E03DA36145EF5A',
			height: 319
		}
	}

cmt_queryDelegator
------------------

Query the current stake status of a specific delegator.

**Parameters**

	* ``delegatorAddress`` String - The delegator address.
	* ``height`` Number - The block number. Default to 0, means current head of the blockchain. NOT IMPLEMENTED YET.

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Object - The delegator object.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryDelegator","params":["0x3a436deae68b7d4c8ff9f1cb0498913a397472d7", 0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"height": 117466,
			"data": [{
				"id": 780,
				"delegator_address": "0xcc64debb948ff9a2cb9ac5cbd292cef1d380221f",
				"pub_key": {
					"type": "tendermint/PubKeyEd25519",
					"value": "sbmLYMzeezCgqJKQBXNVAiZtsdSAx75JUzAtwzWv9pw="
				},
				"validator_address": "0x70A52fF393256f016939Ae2926CBd999508A555B",
				"delegate_amount": "34310000000000000000000",
				"award_amount": "39040378244652451965",
				"withdraw_amount": "0",
				"pending_withdraw_amount": "0",
				"slash_amount": "0",
				"comp_rate": "1/4",
				"voting_power": 970,
				"created_at": 1540551045,
				"state": "Y",
				"block_height": 86269,
				"average_staking_date": 4,
				"candidate_id": 30
			}]
		}
	}

Governance methods
==================

cmt_proposeTransferFund
-----------------------

Propose a fund recovery proposal.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. Must be a validator.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``transferFrom`` String - From account address.
	* ``transferTo`` String - To account address.
	* ``amount`` String - Amount of CMTs in Wei.
	* ``reason`` String - (optional) Reason.
	* ``expireBlockHeight`` Number - (optional) Expiration block height.
	* ``expireTimestamp`` Number - (optional) Timestamp when the proposal will expire.

	Note: You can specify expiration block height or timestamp, but not both. If none is specified, a default of 7 days, as measured in block height(7 * 24 * 60 * 60 / 10), will be used.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed. If successful, the ProposalID will be set in the data property, for validators to vote later.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_proposeTransferFund","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "transferFrom":"0xc4abd0339eb8d57087278718986382264244252f", "transferTo":"0x11f4d0A3c12e86B4b5F39B213F7E19D048276DAe", "amount":"0x186A0"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				data: 'JTUx+ODH0/OSdgfC0Sn66qjn2tX8LfvbiwnArzNpIus=',
				gasUsed ": '2000000',
				fee: {
					key: 'R2FzRmVl',
					value: "4000000000000000'
				}
			},
			hash: '95A004438F89E809657EB119ACBDB42A33725B39',
			height: 561
		}
	}

cmt_proposeChangeParam
----------------------

Propose a system parameter change.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. Must be a validator.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``name`` String - The name of the parameter.
	* ``value`` String - New value of the parameter.
	* ``reason`` String - (optional) Reason.
	* ``expireBlockHeight`` Number - (optional) Expiration block height.
	* ``expireTimestamp`` Number - (optional) Timestamp when the proposal will expire.

	Note: You can specify expiration block height or timestamp, but not both. If none is specified, a default of 7 days, as measured in block height(7 * 24 * 60 * 60 / 10), will be used.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed. If successful, the ProposalID will be set in the data property, for validators to vote later.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_proposeChangeParam","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "name":"gas_price", "value":"3000000000"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				data: 'JTUx+ODH0/OSdgfC0Sn66qjn2tX8LfvbiwnArzNpIus=',
				gasUsed ": '2000000',
				fee: {
					key: 'R2FzRmVl',
					value: "4000000000000000'
				}
			},
			hash: '95A004438F89E809657EB119ACBDB42A33725B39',
			height: 561
		}
	}

cmt_proposeDeployLibEni
-----------------------

Propose a new library for ENI.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. Must be a validator.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``name`` String - The name of the library.
	* ``version`` String - Version of the library, data format: vX.Y.Z, where X, Y, and Z are non-negative integers.
	* ``fileUrl`` String - JSON string of key/value pairs. Key is the name of the OS(so far, only ubuntu and centos are supported), value is the URL array to retrieve the library file.
	* ``md5`` String - JSON string of key/value pairs. Key is the name of the OS(so far, only ubuntu and centos are supported), value is the MD5 of the library file.
	* ``reason`` String - (optional) Reason.
	* ``deployBlockHeight`` Number - (optional) The block number where the new ENI library will deploy.
	* ``deployTimestamp`` Number - (optional) Timestamp when the new ENI library will deploy.

	Note: You can specify deploy block height or timestamp, but not both. If none is specified, a default of 7 days, as measured in block height(7 * 24 * 60 * 60 / 10), will be used.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed. If successful, the ProposalID will be set in the data property, for validators to vote later.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_proposeDeployLibEni","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "name":"reverse", "version":"v1.0.0", "fileUrl":"{\"ubuntu\": [\"<url1>\", \"<url2>\"], \"centos\": [\"<url1>\", \"<url2>\"]}", "md5":"{\"ubuntu\": \"<md5 text>\", \"centos\": \"<md5 text>\"}"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				data: 'JTUx+ODH0/OSdgfC0Sn66qjn2tX8LfvbiwnArzNpIus=',
				gasUsed ": '2000000',
				fee: {
					key: 'R2FzRmVl',
					value: "4000000000000000'
				}
			},
			hash: '95A004438F89E809657EB119ACBDB42A33725B39',
			height: 561
		}
	}

cmt_proposeRetireProgram
------------------------

Propose to retire the program.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. Must be a validator.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``preservedValidators`` String - A comma seperated validator public key list. Valiators in this list will be preserved, other validators will be deactivated.
	* ``reason`` String - (optional) Reason.
	* ``retiredBlockHeight`` Number - (optional) The block number where the program will retire. If not specified, a default of 7 days, as measured in block height(7 * 24 * 60 * 60 / 10), will be used.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed. If successful, the ProposalID will be set in the data property, for validators to vote later.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_proposeRetireProgram","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "preservedValidators":"Esdo0ZN+nHduoi/kNqjdQSNFmNyv2M3Tie/eZeC25gM=,X6qJkoWxW8YkEHquJQM7mZcfpt5r+l8V6C8rbg8dEHQ=", "reason":"System Upgrade"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				data: 'JTUx+ODH0/OSdgfC0Sn66qjn2tX8LfvbiwnArzNpIus=',
				gasUsed ": '2000000',
				fee: {
					key: 'R2FzRmVl',
					value: "4000000000000000'
				}
			},
			hash: '95A004438F89E809657EB119ACBDB42A33725B39',
			height: 561
		}
	}

cmt_vote
--------

Vote on proposals of making changes to the system state.

Here are some use cases:

	* Vote to change system wide parameters such as the system inflation rate.
	* Vote to accept new native libraries for ENI.
	* Vote to recover funds for users.

**Parameters**

	* ``from`` String - The address for the sending account. Uses the web3.cmt.defaultAccount property, if not specified. Must be a validator.
	* ``nonce`` Number - (optional) The number of transactions made by the sender prior to this one.
	* ``proposalId`` String - The Proposal ID to vote.
	* ``answer`` String - Y or N.

**Returns**

	* ``height`` Number - The block number where the transaction is in. =0 if failed.
	* ``hash`` String - Hash of the transaction.
	* ``check_tx`` Object - CheckTx result. Contains error code and log if failed.
	* ``deliver_tx`` Object - DeliverTx result. Contains error code and log if failed.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_vote","params":[{"from":"0x7eff122b94897ea5b0e2a9abf47b86337fafebdc", "proposalId":"JTUx+ODH0/OSdgfC0Sn66qjn2tX8LfvbiwnArzNpIus=", "answer":"Y"}],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			check_tx: {
				fee: {}
			},
			deliver_tx: {
				fee: {}
			},
			hash: '95A004438F89E809657EB119ACBDB42A33725B39',
			height: 561
		}
	}

cmt_queryProposals
------------------

Returns a list of all proposals.

**Parameters**

	none

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Array - An array of all proposals

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryProposals","params":[],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"height": 58,
			"data": [{
					"Id": "/YRNInf2DpWJ6KBcS+Xqa+EUiBH3DMgeM2T57tsMd2E=",
					"Type": "transfer_fund",
					"Proposer": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
					"BlockHeight": 15,
					"ExpireBlockHeight": 20,
					"CreatedAt": "2018-07-03T14:27:11Z",
					"Result": "Expired",
					"ResultMsg": "",
					"ResultBlockHeight": 20,
					"ResultAt": "2018-07-03T14:28:01Z",
					"Detail": {
						"amount": "16",
						"from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
						"reason": "",
						"to": "0xd5bb0351974eca5d116eff840a03a9b96d8ba9e7"
					}
				},
				{
					"Id": "DN6utTAmgX9Iy7naroaKgO2dEbIkwmwRPmmfk35cdEE=",
					"Type": "change_param",
					"Proposer": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
					"BlockHeight": 16,
					"ExpireBlockHeight": 60496,
					"CreatedAt": "2018-07-03T14:27:21Z",
					"Result": "",
					"ResultMsg": "",
					"ResultBlockHeight": 0,
					"ResultAt": "",
					"Detail": {
						"name": "gas_price",
						"reason": "test",
						"value": "3000000000"
					}
				}
			]
		}
	}

cmt_queryParams
---------------

Returns current settings of system parameters.

**Parameters**

	* ``height`` Number - The block number. Default to 0, means current head of the blockchain.

**Returns**

	* ``height`` Number - Current block number or the block number if specified.
	* ``data`` Array - An array of all proposals.

**Example**

::

	// Request
	curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"cmt_queryParams","params":[0],"id":1}'

    // Result
	{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"height": 1000,
			"data": {
				"max_vals": 19,
				"backup_vals": 5,
				"self_staking_ratio": "1/10",
				"inflation_rate": "2/25",
				"validator_size_threshold": "3/25",
				"unstake_waiting_period": 60480,
				"proposal_expire_period": 60480,
				"declare_candidacy_gas": 1000000,
				"update_candidacy_gas": 1000000,
				"set_comp_rate_gas": 21000,
				"update_candidate_account_gas": 1000000,
				"accept_candidate_account_update_request_gas": 1000000,
				"transfer_fund_proposal_gas": 2000000,
				"change_params_proposal_gas": 2000000,
				"deploy_libeni_proposal_gas": 2000000,
				"retire_program_proposal_gas": 2000000,
				"upgrade_program_proposal_gas": 2000000,
				"gas_price": 2000000000
			}
		}
	}

