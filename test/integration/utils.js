const expect = require("chai").expect
const fs = require("fs")
const async = require("async")
const logger = require("./logger")
const { Settings } = require("./constants")
const Globals = require("./global_vars")

const transfer = (f, t, v, gasPrice, nonce) => {
  let payload = {
    from: f,
    to: t,
    value: v,
    gasPrice: web3.toWei(gasPrice || 0, "gwei")
  }
  if (nonce) payload.nonce = nonce
  let hash = null
  try {
    hash = web3.cmt.sendTransaction(payload)
    logger.debug(`transfer ${v} wei from ${f} to ${t}, hash: ${hash}`)
    // check hash
    expect(hash).to.not.empty
  } catch (err) {
    logger.error(err.message)
  }
  return hash
}

const getBalance = (index = null) => {
  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    if (index === null || i == index) {
      balance[i] = web3.cmt.getBalance(Globals.Accounts[i], "latest")
    }
  }
  logger.debug(`balance in wei: --> ${balance}`)
  return balance
}

const tokenFile = "./ETHToken.json"
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
  let hash = null
  try {
    hash = tokenInstance.transfer.sendTransaction(t, v, option)
    logger.debug("token transfer hash: ", hash)
    // check hash
    expect(hash).to.not.empty
  } catch (err) {
    logger.error(err.message)
  }
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
    balance[i] = tokenInstance.balanceOf(Globals.Accounts[i])
  }
  logger.debug(`token balance: --> ${balance}`)
  return balance
}

const calcAward = powers => {
  let total = powers.reduce((s, v) => {
    return s + v
  })
  let origin = powers.map(p => p / total)
  let round1 = origin.map(p => (p > 0.1 ? 0.1 : p))

  let left =
    1 -
    round1.reduce((s, v) => {
      return s + v
    })
  let round2 = origin.map(p => left * p)

  const strip = (x, precision = 12) => parseFloat(x.toPrecision(precision))
  let final = round1.map((p, idx) => {
    return strip(p + round2[idx])
  })
  // console.log(final)

  let result = powers.map((p, idx) => p + final[idx] * Globals.BlockAwards)
  // console.log(result)
  return result
}

const calcAwards = (powers, blocks) => {
  for (i = 0; i < blocks; ++i) {
    powers = calcAward(powers)
  }
  return powers
}

const vote = (proposalId, from, answer) => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.vote({
    from: from,
    proposalId: proposalId,
    answer: answer
  })
  expectTxSuccess(r)
}

const getProposal = proposalId => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.queryProposals()
  expect(r.data.length).to.be.above(0)
  if (r.data.length > 0) {
    proposal = r.data.filter(d => d.Id == proposalId)
    expect(proposal.length).to.equal(1)
    return proposal[0]
  }
  return {}
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
  let waitAll = arrTxhash
    .filter(e => {
      return e
    })
    .map(txhash => {
      return waitInterval.bind(null, txhash)
    })

  async.parallel(waitAll, (err, res) => {
    if (err) {
      return cb(err, res)
    }
    cb(null, res)
  })
}

const waitBlocks = (done, blocks = 1) => {
  let startingBlock = web3.cmt.blockNumber
  logger.debug("waiting start: ", startingBlock)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    if (blocksGone == blocks) {
      logger.debug("waiting end. ")
      clearInterval(interval)
      done()
    }
  }, Settings.IntervalMs || 100)
}

const expectTxFail = (r, check_err, deliver_err) => {
  logger.debug(r)
  expect(r)
    .to.have.property("height")
    .and.to.eq(0)

  if (check_err) {
    expect(r.check_tx.code).to.eq(check_err)
  } else if (deliver_err) {
    expect(r.deliver_tx.code).to.eq(deliver_err)
  }
}

const expectTxSuccess = r => {
  logger.debug(r)
  expect(r)
    .to.have.property("height")
    .and.to.gt(0)
}

module.exports = {
  transfer,
  getBalance,
  newContract,
  tokenTransfer,
  tokenKill,
  getTokenBalance,
  calcAwards,
  vote,
  getProposal,
  waitInterval,
  waitMultiple,
  waitBlocks,
  expectTxFail,
  expectTxSuccess
}
