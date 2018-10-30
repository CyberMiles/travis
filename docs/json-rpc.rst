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
					}, {
						"validator_address": "1F6C181B603013A946246C1392E955200F4E925D",
						"validator_index": 1,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.825570137Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [88, 99, 90, 121, 144, 91, 41, 194, 57, 53, 219, 143, 128, 89, 200, 204, 220, 164, 94, 100, 238, 99, 79, 65, 224, 142, 93, 198, 181, 28, 19, 110, 224, 10, 200, 74, 216, 195, 127, 74, 33, 14, 60, 198, 107, 183, 29, 34, 31, 7, 118, 198, 4, 0, 185, 56, 141, 39, 84, 128, 228, 64, 195, 5]
					}, {
						"validator_address": "3C4398098DA1918E79AF14ABEDE89CE271CE513D",
						"validator_index": 2,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.798335809Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [254, 198, 244, 223, 73, 199, 56, 11, 88, 112, 64, 200, 120, 213, 168, 240, 21, 213, 11, 33, 30, 178, 29, 13, 88, 123, 160, 138, 188, 23, 9, 226, 74, 88, 102, 64, 225, 245, 12, 28, 226, 188, 200, 4, 10, 12, 123, 66, 128, 60, 192, 126, 74, 149, 51, 173, 40, 9, 203, 241, 66, 213, 115, 12]
					}, {
						"validator_address": "3F5E4EB12B99508F41A500D057ADFE17F58B4A9F",
						"validator_index": 3,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.825605198Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [195, 37, 239, 205, 193, 79, 137, 165, 189, 50, 147, 176, 96, 102, 141, 170, 38, 138, 88, 124, 132, 243, 238, 33, 72, 193, 111, 220, 46, 121, 0, 224, 95, 109, 98, 155, 42, 226, 103, 3, 174, 141, 148, 87, 181, 43, 198, 221, 200, 89, 250, 81, 63, 94, 135, 124, 184, 109, 58, 164, 128, 38, 219, 11]
					}, {
						"validator_address": "44BC4129772DAA3ECF8A4027AF1CF251B9D05DB4",
						"validator_index": 4,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.808380112Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [31, 40, 7, 154, 15, 35, 65, 176, 177, 100, 35, 52, 110, 103, 9, 70, 132, 90, 12, 18, 208, 230, 138, 0, 6, 65, 41, 130, 207, 194, 57, 101, 16, 158, 200, 2, 61, 248, 99, 79, 244, 116, 86, 181, 184, 56, 174, 201, 33, 185, 210, 27, 76, 72, 73, 110, 180, 91, 185, 95, 105, 148, 193, 15]
					}, {
						"validator_address": "4A2BAD606492F71D6EBA2D9BE933AD1F54DA538F",
						"validator_index": 5,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.787835613Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [45, 24, 158, 241, 2, 81, 116, 175, 209, 37, 185, 98, 27, 48, 178, 196, 4, 147, 181, 4, 160, 231, 143, 107, 210, 26, 172, 51, 149, 79, 123, 44, 168, 125, 63, 60, 123, 203, 128, 165, 10, 27, 108, 205, 111, 206, 19, 22, 248, 136, 78, 54, 228, 110, 11, 29, 245, 212, 154, 166, 99, 193, 249, 2]
					}, {
						"validator_address": "619CC5EAD1A4B6BD6BB55A25F94F23117B92A8DC",
						"validator_index": 6,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.789609821Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [99, 94, 213, 143, 78, 216, 220, 100, 31, 71, 21, 213, 210, 60, 96, 128, 43, 61, 171, 94, 40, 29, 145, 35, 130, 184, 66, 54, 119, 106, 61, 54, 175, 123, 9, 18, 19, 137, 214, 194, 229, 244, 231, 62, 173, 127, 194, 58, 252, 95, 230, 50, 190, 122, 155, 29, 38, 250, 17, 213, 14, 123, 202, 10]
					}, {
						"validator_address": "61E6A6027D5516A9335C7DA65096B916E8412F41",
						"validator_index": 7,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.806934139Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [52, 195, 94, 4, 174, 112, 179, 183, 90, 173, 159, 234, 11, 9, 174, 50, 68, 28, 216, 212, 20, 219, 98, 92, 6, 229, 143, 190, 246, 228, 54, 41, 64, 47, 170, 198, 208, 212, 58, 40, 7, 19, 202, 133, 161, 94, 126, 70, 186, 165, 194, 249, 223, 113, 43, 97, 108, 95, 228, 120, 9, 251, 243, 11]
					}, {
						"validator_address": "6D62D310DB55AFB3FD3191268C4C874AFE5C28AA",
						"validator_index": 8,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.827155705Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [94, 25, 147, 153, 67, 175, 85, 49, 252, 100, 255, 147, 154, 142, 99, 108, 45, 82, 34, 194, 75, 92, 102, 231, 84, 228, 172, 252, 189, 2, 206, 189, 183, 247, 170, 194, 198, 211, 68, 189, 253, 97, 108, 35, 173, 87, 191, 23, 193, 122, 31, 124, 98, 25, 14, 4, 210, 70, 89, 255, 96, 48, 224, 10]
					}, {
						"validator_address": "6E5AB041827C61DD6F48CA19533C5647C0920DA5",
						"validator_index": 9,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.825080604Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [237, 29, 31, 126, 29, 152, 113, 245, 167, 143, 137, 217, 191, 118, 21, 26, 245, 187, 2, 24, 49, 250, 175, 208, 105, 119, 198, 40, 21, 208, 94, 121, 61, 134, 30, 225, 48, 203, 175, 150, 7, 37, 208, 202, 255, 252, 66, 25, 1, 144, 80, 237, 140, 87, 126, 152, 233, 206, 99, 119, 162, 28, 82, 3]
					}, {
						"validator_address": "757AF694C907F4BF6734EC2177DD7451A3235116",
						"validator_index": 10,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.82643479Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [52, 227, 202, 190, 182, 211, 134, 240, 47, 118, 248, 165, 217, 215, 6, 232, 102, 122, 91, 113, 105, 183, 190, 74, 245, 53, 206, 95, 144, 13, 167, 163, 135, 112, 221, 142, 252, 252, 62, 196, 220, 11, 37, 55, 198, 222, 209, 37, 116, 134, 55, 58, 59, 227, 85, 4, 81, 212, 98, 236, 133, 12, 250, 11]
					}, {
						"validator_address": "96CEF2CFFEE517525B2AA421EF8729793001B335",
						"validator_index": 11,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.824460325Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [106, 252, 104, 14, 139, 180, 113, 119, 158, 47, 214, 79, 188, 62, 11, 4, 1, 193, 13, 188, 52, 130, 216, 111, 182, 93, 33, 55, 224, 232, 36, 22, 212, 216, 196, 38, 133, 223, 143, 109, 202, 216, 91, 107, 23, 244, 10, 9, 255, 91, 153, 218, 73, 17, 30, 59, 65, 247, 11, 154, 213, 119, 145, 5]
					}, {
						"validator_address": "B734553C0375DAF7C93C1316B5D80B6A276673D7",
						"validator_index": 12,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.798340289Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [206, 143, 217, 19, 3, 240, 83, 136, 177, 131, 35, 97, 120, 35, 80, 246, 232, 251, 25, 17, 5, 207, 62, 186, 129, 200, 115, 97, 170, 236, 107, 94, 70, 222, 250, 108, 89, 104, 155, 45, 92, 114, 203, 221, 136, 116, 35, 69, 200, 175, 93, 86, 135, 12, 189, 155, 0, 72, 201, 106, 23, 192, 16, 1]
					}, {
						"validator_address": "B8EA3701DDD82F32193CA24B1C205B97CC3BBAE5",
						"validator_index": 13,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.826557761Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [72, 214, 19, 116, 83, 80, 47, 135, 59, 243, 135, 209, 58, 152, 202, 159, 203, 162, 150, 116, 178, 42, 225, 108, 37, 50, 149, 93, 150, 165, 171, 143, 194, 41, 94, 40, 125, 68, 116, 120, 182, 102, 221, 249, 138, 49, 173, 131, 201, 39, 52, 134, 224, 125, 21, 47, 18, 246, 123, 73, 41, 223, 39, 1]
					}, {
						"validator_address": "B979C7CA5D2D5503F4A418D24447C282C84481D4",
						"validator_index": 14,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.826480251Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [23, 175, 156, 39, 157, 7, 90, 87, 248, 88, 6, 235, 147, 144, 143, 74, 128, 112, 74, 242, 195, 79, 189, 238, 48, 192, 127, 171, 96, 171, 125, 71, 159, 208, 38, 57, 33, 95, 48, 181, 62, 162, 38, 13, 64, 200, 204, 122, 67, 8, 181, 82, 34, 20, 96, 64, 142, 90, 98, 125, 44, 43, 7, 0]
					}, {
						"validator_address": "DC69C6804337288B064A5673422126D95CCEFF24",
						"validator_index": 15,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.806626531Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [227, 76, 174, 6, 180, 162, 94, 59, 1, 222, 84, 106, 130, 153, 167, 231, 144, 43, 52, 134, 92, 210, 70, 116, 253, 193, 226, 245, 244, 187, 88, 205, 49, 25, 93, 227, 6, 140, 204, 163, 133, 132, 57, 95, 130, 192, 106, 219, 91, 74, 169, 126, 186, 80, 232, 119, 158, 224, 159, 185, 112, 197, 129, 6]
					}, {
						"validator_address": "DE429339330886947974E3BC4C60A494BDDDD04E",
						"validator_index": 16,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.796493127Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [104, 52, 116, 179, 198, 247, 181, 190, 172, 151, 101, 40, 73, 86, 86, 41, 109, 88, 13, 225, 13, 221, 12, 232, 86, 87, 126, 44, 22, 140, 223, 68, 127, 232, 213, 121, 123, 171, 53, 234, 238, 194, 86, 79, 232, 62, 53, 49, 193, 75, 87, 169, 249, 194, 204, 37, 170, 216, 196, 105, 137, 116, 35, 3]
					}, {
						"validator_address": "E61E0BBF16E737F5724497C5BC3026900F57FE79",
						"validator_index": 17,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.779029782Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [3, 238, 115, 154, 252, 44, 255, 254, 208, 83, 222, 243, 83, 163, 77, 86, 223, 29, 210, 48, 153, 168, 66, 248, 175, 219, 254, 2, 139, 66, 2, 126, 79, 210, 30, 251, 20, 190, 108, 231, 85, 78, 122, 195, 207, 102, 211, 156, 105, 202, 102, 201, 108, 168, 196, 148, 38, 36, 161, 229, 165, 132, 189, 14]
					}, {
						"validator_address": "E93EB466A5A9A7E0B09385A60A92BC78A2059D6A",
						"validator_index": 18,
						"height": 77,
						"round": 0,
						"timestamp": "2018-10-15T13:41:30.796864767Z",
						"type": 2,
						"block_id": {
							"hash": "50E1722AE5E7FA9D2C4E939356FF0F9487C26E03",
							"parts": {
								"total": 1,
								"hash": "94F88571468784DD05B8BB963358D5E47B68EDB6"
							}
						},
						"signature": [47, 54, 177, 209, 45, 71, 54, 78, 178, 83, 244, 236, 180, 95, 66, 53, 28, 234, 102, 10, 163, 171, 233, 0, 96, 179, 202, 234, 203, 176, 50, 64, 140, 116, 172, 239, 110, 19, 129, 203, 112, 54, 134, 226, 31, 146, 129, 69, 87, 131, 255, 255, 206, 131, 70, 37, 132, 199, 157, 22, 85, 23, 111, 15]
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
				}, {
					"id": 20,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "G2LmdNfRTo7xgKRKvrljuUKf81kt7fWsnv0LuOiostc="
					},
					"owner_address": "0x482A7CBb8f66A9Db6B25808861B182c670c79259",
					"shares": "355306538538502705594152",
					"voting_power": 22197,
					"pending_voting_power": 0,
					"max_shares": "3000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 1539616790,
					"description": {
						"name": "Seed Backup Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external backup validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 629,
					"rank": 22,
					"state": "Backup Validator",
					"num_of_delegators": 1
				}, {
					"id": 21,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "3zJgyq7S36dSAZbBxSlWSL+wDHG4PceQ41GfL6LdIjM="
					},
					"owner_address": "0x04BA6Cf9a4035294958678dd0f540A195b260D0E",
					"shares": "360842852078185954523980",
					"voting_power": 27830,
					"pending_voting_power": 0,
					"max_shares": "3000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 1539616854,
					"description": {
						"name": "Seed Backup Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external backup validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 635,
					"rank": 20,
					"state": "Backup Validator",
					"num_of_delegators": 2
				}, {
					"id": 22,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "W1Cy2m4+BeMuMKF/LtnZDXdfZyDsojLha7RZRXeCxT0="
					},
					"owner_address": "0xe218509490578f75dfc6eD6C8a80158675071A8C",
					"shares": "360836820076859684767365",
					"voting_power": 27830,
					"pending_voting_power": 0,
					"max_shares": "3000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 1539616918,
					"description": {
						"name": "Seed Backup Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external backup validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 641,
					"rank": 19,
					"state": "Backup Validator",
					"num_of_delegators": 2
				}, {
					"id": 23,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "+uiRWgjNNOw9KsLuEVw+ebsMyr5NMJoeJjrGJyf5Img="
					},
					"owner_address": "0xAe3BeFdc5D0F5397B9e448fE136F10360DddDE28",
					"shares": "356020195035550651143757",
					"voting_power": 22241,
					"pending_voting_power": 0,
					"max_shares": "3000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 1539616993,
					"description": {
						"name": "Seed Backup Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external backup validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 648,
					"rank": 21,
					"state": "Backup Validator",
					"num_of_delegators": 1
				}, {
					"id": 24,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "la9UNAzRQUc2vpGGRnt8AeVkN3+MvR0PghWwxsgZmgI="
					},
					"owner_address": "0x72cf924c62BAFf2ED74A5ceb885082B814216E55",
					"shares": "354471018130724534293675",
					"voting_power": 22145,
					"pending_voting_power": 0,
					"max_shares": "3000000000000000000000000",
					"comp_rate": "99/100",
					"created_at": 1539617047,
					"description": {
						"name": "Seed Backup Validator",
						"website": "https://www.cybermiles.io/seed-validator/",
						"location": "HK",
						"email": "developer@cybermiles.io",
						"profile": "To be replaced by an external backup validator."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 653,
					"rank": 23,
					"state": "Backup Validator",
					"num_of_delegators": 1
				}, {
					"id": 25,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "tt7IMrBNAIn37vp6HMLzSAEY62fw6EYFCxjUno1Jc4Y="
					},
					"owner_address": "0xFD0e8E4C4DeA053f10e72E8800B08ac875e5Ac49",
					"shares": "2129465497865759280748953",
					"voting_power": 183898,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539617981,
					"description": {
						"name": "Hash Tower",
						"website": "",
						"location": "Seoul, South Korea",
						"email": "",
						"profile": "The Freedom of Finance"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 740,
					"rank": 13,
					"state": "Validator",
					"num_of_delegators": 4
				}, {
					"id": 26,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "nPId4U9gPCJRoDd61O6G3QIUz+iqB2n4qJPvngu114w="
					},
					"owner_address": "0x1724D4a82F29D93A1eB96c19B4BB6B219dc18F23",
					"shares": "4998171464708511154584293",
					"voting_power": 359595,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539618966,
					"description": {
						"name": "Hayek Capital",
						"website": "http://www.hayek.capital",
						"location": "Australia",
						"email": "",
						"profile": "Contributing the Greatness of CyberMiles"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 838,
					"rank": 5,
					"state": "Validator",
					"num_of_delegators": 37
				}, {
					"id": 27,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "KqG4LPBhHqvB8NqKKOA+7ZhcKGRmVmwYgQEsfo7aQfY="
					},
					"owner_address": "0x1ac7d4F1D4Fa3eaEF67d8208a2B1B84670211e75",
					"shares": "2340509468453632060954289",
					"voting_power": 204790,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539619028,
					"description": {
						"name": "TGL Capital",
						"website": "",
						"location": "Beijing China",
						"email": "",
						"profile": "For a better CyberMiles"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 844,
					"rank": 10,
					"state": "Validator",
					"num_of_delegators": 5
				}, {
					"id": 28,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "KRuHqKmSypC9tAP93OWNrQ+9kFuQuFUKpq3yxqTnUJ8="
					},
					"owner_address": "0xEB65290b802DF113300120C52B313F1896e80d38",
					"shares": "2176145424557633182219051",
					"voting_power": 198207,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "11/20",
					"created_at": 1539619092,
					"description": {
						"name": "Moon Fund",
						"website": "",
						"location": "San Francisco, USA",
						"email": "",
						"profile": "A serious builder of CMT ecosystem"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 850,
					"rank": 11,
					"state": "Validator",
					"num_of_delegators": 10
				}, {
					"id": 29,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "NEUjQM4EvOkTruH0aufgQM4tLEKrCJSAvEDKwZ771ng="
					},
					"owner_address": "0x858578e81a0259338b4d897553aFA7b9c363e769",
					"shares": "2098963173886421224173073",
					"voting_power": 161883,
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
				}, {
					"id": 30,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "sbmLYMzeezCgqJKQBXNVAiZtsdSAx75JUzAtwzWv9pw="
					},
					"owner_address": "0x70A52fF393256f016939Ae2926CBd999508A555B",
					"shares": "5199889981065108184838090",
					"voting_power": 332626,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539619476,
					"description": {
						"name": "Noomi",
						"website": "https://noomiwallet.com/",
						"location": "Malta",
						"email": "supernode@noomi-email.com",
						"profile": "Crypto made simple"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 886,
					"rank": 6,
					"state": "Validator",
					"num_of_delegators": 40
				}, {
					"id": 31,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "PvJM9eB+eG8Q+wSGQVV9Laaar2hiI1Gh1GCOpO/2/0A="
					},
					"owner_address": "0x4cdaf011CadbA6c3997252738E4D6Dd30C8865b9",
					"shares": "2105928256333725254471463",
					"voting_power": 162420,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539625859,
					"description": {
						"name": "Portland Master Limited",
						"website": "",
						"location": "Hong Kong",
						"email": "",
						"profile": "The Tiger Sniffs the Rose"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 1482,
					"rank": 14,
					"state": "Validator",
					"num_of_delegators": 2
				}, {
					"id": 32,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "toVWDAEdxoLTgfnwYsv7Ol/jByQ8+oGu87oaD539iJs="
					},
					"owner_address": "0x0Da518EcF4761A86965c1F77Ac4c1bD6e19904E3",
					"shares": "2310172266298909256227801",
					"voting_power": 192651,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539626062,
					"description": {
						"name": "Snow Eagle Group Limited",
						"website": "",
						"location": "Israel",
						"email": "",
						"profile": "Bring CMT to Israel"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 1501,
					"rank": 12,
					"state": "Validator",
					"num_of_delegators": 4
				}, {
					"id": 33,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "b3Vt3qMWJjwn1yDeNATOr7625CtrrAbCz2gj68RmoFk="
					},
					"owner_address": "0x34C5f1c0E10701dbaF0dF1Ad2a7826bE41a3A380",
					"shares": "23633622169458333224011367",
					"voting_power": 938879,
					"pending_voting_power": 0,
					"max_shares": "31000000000000000000000000",
					"comp_rate": "3/10",
					"created_at": 1539626266,
					"description": {
						"name": "SSSnodes",
						"website": "http://www.sssnodes.com",
						"location": "Shenzhen, China",
						"email": "cmt100@sssnodes.com",
						"profile": "CMT, you're not alone"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 1520,
					"rank": 1,
					"state": "Validator",
					"num_of_delegators": 360
				}, {
					"id": 34,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "cDMNNNmRgSBVEGO2dvLX4sCOM95clhJC34480sli4dM="
					},
					"owner_address": "0x8958618332dF62AF93053cb9c535e26462c959B0",
					"shares": "3474243976409926295161881",
					"voting_power": 307196,
					"pending_voting_power": 0,
					"max_shares": "3e+25",
					"comp_rate": "1/2",
					"created_at": 1539664044,
					"description": {
						"name": "COBINHOOD",
						"website": "https://cobinhood.com",
						"location": "Taipei, Taiwan",
						"email": "",
						"profile": ""
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 5018,
					"rank": 7,
					"state": "Validator",
					"num_of_delegators": 10
				}, {
					"id": 35,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "IbIGjMB55C8HxZAKhY5Z1e6K1KrqcOZLNTVRzJZzsic="
					},
					"owner_address": "0x1B92C5BB82972aF385d8cD8c1230502083898BA6",
					"shares": "19676321159726395118467646",
					"voting_power": 1111200,
					"pending_voting_power": 0,
					"max_shares": "139999000000000000000000000",
					"comp_rate": "11/20",
					"created_at": 1539670799,
					"description": {
						"name": "CyberMiles Vietnam",
						"website": "https://andrewkg.com/ico_listing/genesis-validator-daocities/",
						"location": "Sai Gon Vietnam",
						"email": "",
						"profile": "Shall the first be the last"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 5640,
					"rank": 0,
					"state": "Validator",
					"num_of_delegators": 169
				}, {
					"id": 36,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "CktEyd0r55nJRjxDlM0ydZGcJiqCiMK8g7Sc4TVxkPE="
					},
					"owner_address": "0xcD3090E881170f6D036fdb3aE5a3d36EAD5bCF83",
					"shares": "4063833662354279701592619",
					"voting_power": 368427,
					"pending_voting_power": 0,
					"max_shares": "39000000000000000000000000",
					"comp_rate": "11/20",
					"created_at": 1539680931,
					"description": {
						"name": "Lvl99",
						"website": "",
						"location": "Jakarta, Indonesia",
						"email": "edward@lvl99.org",
						"profile": ""
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 6584,
					"rank": 4,
					"state": "Validator",
					"num_of_delegators": 7
				}, {
					"id": 37,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "ykA4RmtjWMi1fkVnHbrQcJf2FiRlPCtIyAPXtOGj8gk="
					},
					"owner_address": "0x3Af427d092F9BF934d2127408935C1455170ea8a",
					"shares": "2970534241579262663385816",
					"voting_power": 246606,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "2/5",
					"created_at": 1539684685,
					"description": {
						"name": "Wancloud",
						"website": "https://www.wancloud.cloud/",
						"location": "Shanghai, China",
						"email": "",
						"profile": "Wancloud Enterprise Contributor of Blockchain Innovation"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 6933,
					"rank": 8,
					"state": "Validator",
					"num_of_delegators": 21
				}, {
					"id": 39,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "T0Y58ijBb85kVOw90dQt+RUF5IwRaUVSc2aL+uK75y0="
					},
					"owner_address": "0xDD781d32effa5689A22c0A5Ab2B5f2Cd95B91205",
					"shares": "10000000000000000000",
					"voting_power": 0,
					"pending_voting_power": 0,
					"max_shares": "100000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539758299,
					"description": {
						"name": "",
						"website": "",
						"location": "",
						"email": "",
						"profile": ""
					},
					"verified": "N",
					"active": "Y",
					"block_height": 13740,
					"rank": 40,
					"state": "Candidate",
					"num_of_delegators": 1
				}, {
					"id": 40,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "LMLLZg21urvxd5YARKe925eMYc/SonlARbPSdnY1zXE="
					},
					"owner_address": "0xF9Fd397486AC5516eea2304fA701cB9767c465D2",
					"shares": "2394329724131416054270401",
					"voting_power": 205204,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539792560,
					"description": {
						"name": "Huobipool",
						"website": "https://www.huobipool.com/",
						"location": "Beijing, China",
						"email": "",
						"profile": "Make mining easier, Make wealth more free"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 16908,
					"rank": 9,
					"state": "Validator",
					"num_of_delegators": 12
				}, {
					"id": 41,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "OWvoqtElEUFldZsGaOtFmStya8Z+VR4QHNSiT5hdcv8="
					},
					"owner_address": "0x654E1DfE66519B9a09305aD58392d9A1c61296b3",
					"shares": "2048735069918186933093236",
					"voting_power": 127987,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539792799,
					"description": {
						"name": "Hui Zhi Zai Xian",
						"website": "http://www.huizhizaixian.info/",
						"location": "Beijing, China",
						"email": "",
						"profile": "We can change the world"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 16930,
					"rank": 17,
					"state": "Validator",
					"num_of_delegators": 1
				}, {
					"id": 42,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "4DPHPK5LVTOiiqQDoEQXMcPpYrQndMx3EW2riLqB7oU="
					},
					"owner_address": "0xf9A431660DC8e425018564cE707d44A457301Eb9",
					"shares": "2099615614193938432712533",
					"voting_power": 158627,
					"pending_voting_power": 0,
					"max_shares": "20000000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539794471,
					"description": {
						"name": "LiMaGo",
						"website": "https://limago123.com/",
						"location": "Taipei Taiwan",
						"email": "",
						"profile": "Think outside the box change the way you travel"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 17083,
					"rank": 16,
					"state": "Validator",
					"num_of_delegators": 2
				}, {
					"id": 43,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "lZq/Ql4NZZLIerPtoznXFw+CZAkdlpt4m6pnJaLx9og="
					},
					"owner_address": "0x221507F21AAc826263a664538580e57DED401978",
					"shares": "15193103060821758936179637",
					"voting_power": 889931,
					"pending_voting_power": 0,
					"max_shares": "81886217160000000000000000",
					"comp_rate": "2/5",
					"created_at": 1539831861,
					"description": {
						"name": "Krypital Group",
						"website": "https://krypital.com/",
						"location": "Cayman Islands",
						"email": "Contact@krypital.com",
						"profile": "Positioned as the alpha of blockchain consulting firms, Krypital Group leads the industry by strategically mapping out the ecosystem while consistently setting record performances. We create real value for our clients through a complete package of services that include project incubation, advisory, branding, marketing, community operation and management, technical support, and tokenization support. In order to better assist our clients with listing on exchanges. We are also super nodes of Huobi and OKex as well as in-depth collaboration with several other exchanges."
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 20508,
					"rank": 2,
					"state": "Validator",
					"num_of_delegators": 149
				}, {
					"id": 44,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "d2zntt70VR1FUOJwMW4PqxPnRGmJPwNFJY1H8kUqyFI="
					},
					"owner_address": "0x9A3482Fd81D706d5aA941f38946Af69A448e08C3",
					"shares": "6707116354598875179420130",
					"voting_power": 524176,
					"pending_voting_power": 0,
					"max_shares": "64848000000000000000000000",
					"comp_rate": "1/2",
					"created_at": 1539843587,
					"description": {
						"name": "ArcBlock",
						"website": "https://www.arcblock.io",
						"location": "USA",
						"email": "contact@arcblock.io",
						"profile": "ArcBlock Cybermiles Super Node"
					},
					"verified": "Y",
					"active": "Y",
					"block_height": 21583,
					"rank": 3,
					"state": "Validator",
					"num_of_delegators": 9
				}, {
					"id": 45,
					"pub_key": {
						"type": "tendermint/PubKeyEd25519",
						"value": "B/vBc87QhilumZDFFvsato5ZdKBHEGKcW15RBT5cfBA="
					},
					"owner_address": "0xE27fA42083bc018F78f2192ea89C9Ea55e2D5c21",
					"shares": "1022937071680964088372",
					"voting_power": 0,
					"pending_voting_power": 55,
					"max_shares": "10000000000000000000000",
					"comp_rate": "11/20",
					"created_at": 1539849597,
					"description": {
						"name": "AlphaCoin Fund",
						"website": "http://www.Alphacoinfund.com",
						"location": "",
						"email": "info@alphacoinfund.com",
						"profile": ""
					},
					"verified": "N",
					"active": "Y",
					"block_height": 22136,
					"rank": 24,
					"state": "Candidate",
					"num_of_delegators": 1
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

