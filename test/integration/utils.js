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
    gasPrice: gasPrice || 0
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
  return index == null ? balance : balance[index]
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
    gasPrice: gasPrice || 0
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

const getDelegation = (acc_index, pk_index) => {
  let delegation = {
    delegate_amount: web3.toBigNumber(0),
    award_amount: web3.toBigNumber(0),
    withdraw_amount: web3.toBigNumber(0),
    slash_amount: web3.toBigNumber(0)
  }
  result = web3.cmt.stake.delegator.query(Globals.Accounts[acc_index], 0)
  if (result && result.data) {
    let data = result.data.find(
      d => d.pub_key.value == Globals.PubKeys[pk_index]
    )
    if (data)
      delegation = {
        delegate_amount: web3.toBigNumber(data.delegate_amount),
        award_amount: web3.toBigNumber(data.award_amount),
        withdraw_amount: web3.toBigNumber(data.withdraw_amount),
        slash_amount: web3.toBigNumber(data.slash_amount)
      }
  }
  logger.debug(
    "delegation: --> ",
    `delegate_amount: ${delegation.delegate_amount.toString(10)}`,
    `award_amount: ${delegation.award_amount.toString(10)}`,
    `withdraw_amount: ${delegation.withdraw_amount.toString(10)}`,
    `slash_amount: ${delegation.slash_amount.toString(10)}`
  )
  return delegation
}

const calcAward = powers => {
  let total = powers.reduce((s, v) => {
    return s + v
  })
  let origin = powers.map(p => p / total)
  let round1 = origin.map(
    p => (p > Globals.ValSizeLimit ? Globals.ValSizeLimit : p)
  )

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
  logger.debug("getProposal:", r)

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
  let startingTime = Math.round(new Date().getTime() / 1000)

  logger.debug("Starting block:", startingBlock)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    let timeGone = Math.round(new Date().getTime() / 1000) - startingTime

    if (blocksGone > Settings.BlockTimeout) {
      clearInterval(interval)
      cb(new Error(`Pending full after ${Settings.BlockTimeout} blocks`))
      return
    }
    if (timeGone > Settings.WaitTimeout) {
      clearInterval(interval)
      logger.error(`Pending full after ${Settings.WaitTimeout} seconds`)
      process.exit(1)
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
    logger.debug(`Blocks Passed ${blocksGone}`)
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

const gasFee = txType => {
  let gasPrice = web3.toBigNumber(Globals.GasPrice)
  let gasLimit = 0
  switch (txType) {
    case "declareCandidacy":
      gasLimit = web3.toBigNumber(Globals.GasLimit.DeclareCandidacy)
      break
    case "updateCandidacy":
      gasLimit = web3.toBigNumber(Globals.GasLimit.UpdateCandidacy)
      break
    case "governancePropose":
      gasLimit = web3.toBigNumber(Globals.GasLimit.TransferFundProposal)
      break
  }
  return gasPrice.times(gasLimit)
}

const addFakeValidators = () => {
  if (Globals.TestMode == "single") {
    let result = web3.cmt.stake.validator.list()
    let valsToAdd = 4 - result.data.length

    if (valsToAdd > 0) {
      Globals.Accounts.forEach((acc, idx) => {
        if (idx >= valsToAdd) return
        let initAmount = 1000,
          compRate = "0.8"
        let payload = {
          from: acc,
          pubKey: Globals.PubKeys[idx],
          maxAmount: web3.toWei(initAmount, "cmt"),
          compRate: compRate
        }
        let r = web3.cmt.stake.validator.declare(payload)
        logger.debug(r)
        logger.debug(`validator ${acc} added, max_amount: ${initAmount} cmt`)
      })
    }
  }
}

const removeFakeValidators = () => {
  if (Globals.TestMode == "single") {
    let result = web3.cmt.stake.validator.list()
    result.data.forEach((val, idx) => {
      // skip the first one
      if (idx == 0) return
      // remove all others
      let acc = val.owner_address
      let r = web3.cmt.stake.validator.withdraw({ from: acc })
      logger.debug(r)
      logger.debug(`validator ${acc} removed`)
    })
  }
}

module.exports = {
  transfer,
  getBalance,
  newContract,
  tokenTransfer,
  tokenKill,
  getTokenBalance,
  getDelegation,
  calcAwards,
  vote,
  getProposal,
  waitInterval,
  waitMultiple,
  waitBlocks,
  expectTxFail,
  expectTxSuccess,
  gasFee,
  addFakeValidators,
  removeFakeValidators
}
