const expect = require("chai").expect
const fs = require("fs")
const async = require("async")
const logger = require("./logger")
const { Settings } = require("./constants")

const transfer = (f, t, v, gasPrice, nonce) => {
  let payload = {
    from: f,
    to: t,
    value: web3.toWei(v, "gwei"),
    gasPrice: web3.toWei(gasPrice || 0, "gwei")
  }
  if (nonce) payload.nonce = nonce
  let hash = web3.cmt.sendTransaction(payload)
  logger.debug(`transfer ${v} gwei from ${f} to ${t}, hash: ${hash}`)
  // check hash
  expect(hash).to.not.empty
  return hash
}

const getBalance = () => {
  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    balance[i] = web3.fromWei(
      web3.cmt.getBalance(accounts[i], "latest"),
      "gwei"
    )
  }
  logger.debug(`balance in gwei: --> ${balance}`)
  return balance
}

const tokenFile = "./TestToken.json"
const tokenJSON = JSON.parse(fs.readFileSync(tokenFile).toString())
const abi = tokenJSON["abi"]
const bytecode = tokenJSON["bytecode"]

const newContract = function(deployAddress, cb) {
  let tokenContract = web3.cmt.contract(abi)
  tokenContract.new(
    {
      from: deployAddress,
      data: bytecode,
      gas: "4700000"
    },
    function(e, contract) {
      if (e) throw e
      if (typeof contract.address !== "undefined") {
        logger.debug(
          "Contract mined! address: " +
            contract.address +
            " transactionHash: " +
            contract.transactionHash
        )
        expect(contract.address).to.not.empty
        cb(contract.address)
      }
    }
  )
}

const tokenTransfer = function(f, t, v, gasPrice, nonce) {
  let tokenContract = web3.cmt.contract(abi)
  let tokenInstance = tokenContract.at(contractAddress)
  let option = {
    from: f,
    gasPrice: web3.toWei(gasPrice || 0, "gwei")
  }
  if (nonce) option.nonce = nonce
  let hash = tokenInstance.transfer.sendTransaction(t, v, option)
  logger.debug("token transfer hash: ", hash)
  // check hash
  expect(hash).to.not.empty
  return hash
}

const tokenKill = deployAdrress => {
  let tokenContract = web3.cmt.contract(abi)
  let tokenInstance = tokenContract.at(contractAddress)
  let hash = tokenInstance.kill({ from: deployAdrress })
  logger.debug("token kill hash: ", hash)
  return hash
}

const getTokenBalance = () => {
  let tokenContract = web3.cmt.contract(abi)
  let tokenInstance = tokenContract.at(contractAddress)

  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    balance[i] = tokenInstance.balanceOf(accounts[i])
  }
  logger.debug(`token balance: --> ${balance}`)
  return balance
}

const waitInterval = function(txhash, cb) {
  let startingBlock = web3.cmt.blockNumber

  logger.debug("Starting block:", startingBlock)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    if (blocksGone > Settings.BlockTimeout) {
      clearInterval(interval)
      cb(new Error(`Pending full after ${Settings.BlockTimeout} blocks`))
      return
    }

    let receipt = web3.cmt.getTransactionReceipt(txhash)
    logger.debug(`Blocks Passed ${blocksGone}, ${txhash} receipt: ${receipt}`)

    if (receipt != null && receipt.blockNumber > 0) {
      clearInterval(interval)
      cb(null, receipt)
    }
  }, Settings.IntervalMs || 100)
}

const waitMultiple = function(arrTxhash, cb) {
  let waitAll = arrTxhash.map(txhash => {
    return waitInterval.bind(null, txhash)
  })

  async.parallel(waitAll, (err, res) => {
    if (err) {
      return cb(err, res)
    }
    cb(null, res)
  })
}

module.exports = {
  transfer,
  getBalance,
  newContract,
  tokenTransfer,
  tokenKill,
  getTokenBalance,
  waitInterval,
  waitMultiple
}
