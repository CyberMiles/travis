===============
Validator
===============

In this document, we will explain staking / unstaking transactions related to validator election, 
as well as the recommended technical specifcations for validator operations.

cmt_declareCandidacy
-------- 
Allow a potential validator to declare its candidacy. Signed by validator address.

**Parameters**

- pubKey: String - Validator node public key
- from: String - An account address associated with this validator (for self-staking and getting block award)
- nonce: Number - (optional) The number of transactions made by the sender prior to this one.
- maxAmount: String - Max amount of CMTs in Wei to be staked.
- compRate: String - Validator compensation rate - the percentage of block awards distributed to the validator
- description: Object - Description of the candidacy
    - name: String - Name of candidate
    - website: String - Web page link
    - location: String - Location (network and geo)
    - email: String - Email address
    - profile: String - Team introduction
"params": [{
    "from": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "transferFrom": "0x77beb894fc9b0ed41231e51f128a347043960a9d",
    "transferTo": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "amount": "0x1",
    "reason": "……"
}]

**Returns**

- Object - Result object
  - `height`: `Number` - The block number where transaction occurs. =0 if failed.
  - `hash`: `String` - Hash of the transaction.
  - `check_tx`: `Object` - CheckTx result. Contain error code and log if failed.
  - `deliver_tx`: `Object` - DeliverTx result. Contain error code and log if failed.

**Example**

  // Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_declareCandidacy","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_withdrawCandidacy
-------- 
Allow a validator to withdraw. All staked tokens will be returned to delegator addresses. Self-staked CMTs will be returned to the validator address. Signed by validator address.

**Parameters**

- from: String - Validator address
- nonce: Number - (optional) The number of transactions made by the sender prior to this one.
  params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 2
  }

**Returns**

* Object - Result object
    * height: Number - The block number where transaction occurs. =0 if failed.
    * hash: String - Hash of the transaction.
    * check_tx: Object - CheckTx result. Contain error code and log if failed.
    * deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_withdrawCandidacy","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_updateCandidacy
-------- 
Allow a validator candidate to update its candidacy. Signed by validator address.

**Parameters**

* from: String - An account address associated with this validator (for self-staking and getting block awards)
* nonce: Number - (optional) The number of transactions made by the sender prior to this one.
* maxAmount: String - Max amount of CMTs in Wei to be staked.
* description: Object - Description of the candidacy.
    * name: String - Name
    * website: String - Web page link
    * location: String - Location (network and geo)
    * email: String - Email address
    * profile: String - Team introduction

params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 3,
  "maxAmount": "1000000000000000000000", // 1000CMT
  "description": {"name": "xxx", "website": "https://yourdomain.com", "location": "CA, US", "email": "admin@email.com", "profile": "..."}
}

**Returns**

* height: Number - The block number where transaction occurs. =0 if failed.
* hash: String - Hash of the transaction.
* check_tx: Object - CheckTx result. Contain error code and log if failed.
* deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_updateCandidacy","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_verifyCandidacy
-------- 

**Parameters**

* from: String - A special address Foundation owns.
* nonce: Number - (optional) The number of transactions made by the sender prior to this one.
* candidateAddress: String - Validator address.
* verified: Boolean - True of false, default to false. 

params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 4,
  "candidateAddress": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
  "verified": true
}

**Returns**

* Object - Result object
    * height: Number - The block number where transaction occurs. =0 if failed.
    * hash: String - Hash of the transaction.
    * check_tx: Object - CheckTx result. Contain error code and log if failed.
    * deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_verifyCandidacy","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_activateCandidacy
-------- 
Allow Foundation to verify a validator's information. It is signed by a special address Foundation owns. This tx can be called multiple times to update the verified status.

**Parameters**

* from: String - Validator address
* nonce: Number - (optional) The number of transactions made by the sender prior to this one.

params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 5
}

**Returns**

* Object - Result object
    * height: Number - The block number where transaction occurs. =0 if failed.
    * hash: String - Hash of the transaction.
    * check_tx: Object - CheckTx result. Contain error code and log if failed.
    * deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_activateCandidacy","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_delegate
-------- 
This tx is used when a delegator stakes CMTs to a validator. Signed by delegator address.

**Parameters**

* from: String - Delegator address.
* nonce: Number - (optional) The number of transactions made by the sender prior to this one.
* validatorAddress: String - Validator address
* amount: String - Amount of CMTs in Wei to stake.

params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 5
  "validatorAddress": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
  "amount": "200000000000000000000" // 200CMT
}

**Returns**

* Object - Result object
    * height: Number - The block number where transaction occurs. =0 if failed.
    * hash: String - Hash of the transaction.
    * check_tx: Object - CheckTx result. Contain error code and log if failed.
    * deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_delegate","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_withdraw
-------- 
This tx is used when a delegator unstakes CMTs from a validator. It will free up some slots from the validator's allocation. Signed by the delegator.

**Parameters**

* from: String - Delegator address
* nonce: Number - (optional) The number of transactions made by the sender prior to this one.
* validatorAddress: String - Validator address
* amount: String - Amount of CMTs to unstake.

params: {
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "nonce": 5
  "validatorAddress": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
  "amount": "200000000000000000000" // 200CMT
}

**Returns**

* Object - Result object

    * height: Number - The block number where transaction occurs. =0 if failed.
    * hash: String - Hash of the transaction.
    * check_tx: Object - CheckTx result. Contain error code and log if failed.
    * deliver_tx: Object - DeliverTx result. Contain error code and log if failed.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_withdraw","params":[{see above}]}'

// Result
{
    check_tx: { fee: {} },
    deliver_tx: { fee: {} },
    hash: '1573F39376D8C10C6B890861CD25FD0BA917556F',
    height: 271
}

cmt_queryValidators
-------- 
Get a list of all running validators, backup validators, and validator candidates, with the amount of CMT staked to each one. Not signed, and no parameter.

**Parameters**

* No parameters.

**Returns**

* Object - Result object
    * height: Number -  Current block number or the block number if specified.  
    * data: Array - An array of all running validators, backup validators and validator candidates.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_queryValidators"}'

// Result
{ 
  "height": 38,
  "data": [
    {
      "pub_key": {
        "type": "AC26791624DE60",
        "value": "6DwZIWYS2BJb0XKtQT7PSJ4f8Qe+hbdn6CHVasl5NYc="
      },
      "owner_address": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
      "shares": "3080166354134956874646",
      "voting_power": 3080,
      "max_shares": "10000000000000000000000",
      "comp_rate": "0.5",
      "created_at": "2018-07-03T10:44:40Z",
      "updated_at": "2018-07-03T14:37:57Z",
      "description": {
        "name": "",
        "website": "",
        "location": "",
        "email": "",
        "profile": ""
      },
      "verified": "N",
      "active": "Y",
      "block_height": 1,
      "rank": 0,
      "state": ""
    }
  ]
}

cmt_queryValidator
-------- 
Query the current stake status of the validator. Not signed.

**Parameters**

* validatorAddress: String - Validator address

params: {
  "validatorAddress": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
}

**Returns**

* Object - Result object
    * height: Number -  Current block number or the block number if specified.  
    * data: Object - The validator, backup validator or candidate object.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_queryValidator","params":{see above}}'

// Result
{ 
  "height": 38,
  "data": {
    "pub_key": {
      "type": "AC26791624DE60",
      "value": "6DwZIWYS2BJb0XKtQT7PSJ4f8Qe+hbdn6CHVasl5NYc="
    },
    "owner_address": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
    "shares": "3359212523592085235879",
    "voting_power": 3359,
    "max_shares": "10000000000000000000000",
    "comp_rate": "0.5",
    "created_at": "2018-07-03T10:44:40Z",
    "updated_at": "2018-07-03T14:39:48Z",
    "description": {
      "name": "",
      "website": "",
      "location": "",
      "email": "",
      "profile": ""
    },
    "verified": "N",
    "active": "Y",
    "block_height": 1,
    "rank": 0,
    "state": ""
  }
}

cmt_queryDelegator
-------- 
Query the current stake status of a delegator. Not signed.

**Parameters**

* delegatorAddress: String - Delegator address

params: {
  "delegatorAddress": "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
}

**Returns**

* Object - Result object
    * height: Number -  Current block number or the block number if specified.  
    * data: Object - The delegator object.

**Example**

// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"cmt_queryDelegator","params":[{see above}]}'

// Result
{ 
  "height": 38,
  "data": [
    {
      "delegator_address": "0x84f444c0405c762afa4ee3e8d8a5b3653ea52549",
      "pub_key": {
        "type": "AC26791624DE60",
        "value": "6DwZIWYS2BJb0XKtQT7PSJ4f8Qe+hbdn6CHVasl5NYc="
      },
      "delegate_amount": "1000000000000000000000",
      "award_amount": "2536787358701166920300",
      "withdraw_amount": "0",
      "slash_amount": "0",
      "created_at": "2018-07-03T10:44:40Z",
      "updated_at": "2018-07-03T14:40:58Z"
    }
  ]
}
