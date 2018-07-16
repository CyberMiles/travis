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
    slash_amount: web3.toBigNumber(0),
    shares: web3.toBigNumber(0)
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
    delegation.shares = delegation.delegate_amount
      .plus(delegation.award_amount)
      .minus(delegation.withdraw_amount)
      .minus(delegation.slash_amount)
  }
  logger.debug(
    "delegation: --> ",
    `delegate_amount: ${delegation.delegate_amount.toString(10)}`,
    `award_amount: ${delegation.award_amount.toString(10)}`,
    `withdraw_amount: ${delegation.withdraw_amount.toString(10)}`,
    `slash_amount: ${delegation.slash_amount.toString(10)}`,
    `shares: ${delegation.shares.toString(10)}`
  )
  return delegation
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

  let r = web3.cmt.governance.listProposals()
  logger.debug("listProposals:", r)

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
  let gasPrice = web3.toBigNumber(Globals.Params.gas_price)
  let gasLimit = 0
  switch (txType) {
    case "declareCandidacy":
      gasLimit = web3.toBigNumber(Globals.Params.declare_candidacy)
      break
    case "updateCandidacy":
      gasLimit = web3.toBigNumber(Globals.Params.update_candidacy)
      break
    case "proposeTransferFund":
      gasLimit = web3.toBigNumber(Globals.Params.transfer_fund_proposal)
      break
    case "proposeChangeParam":
      gasLimit = web3.toBigNumber(Globals.Params.change_params_proposal)
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
        let initAmount = 10000,
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

const getBlockAward = () => {
  const inflation_rate = Globals.Params.inflation_rate
  let cmts = web3.toWei(web3.toBigNumber(1000000000), "cmt")
  let blocksYr = (365 * 24 * 3600) / 10
  let blockAward = cmts.times(inflation_rate / 100).dividedToIntegerBy(blocksYr)
  return blockAward
}

const calcValAward = (award, vals) => {
  logger.debug("calcValAward ->")
  // percentage
  let shares = vals.map(v => web3.toBigNumber(v.shares))
  let total = shares.reduce((s, v) => {
    return s.plus(v)
  })

  const toFixed = f => Number(parseFloat(f).toFixed(12))
  let perc0 = shares.map(s => toFixed(s.div(total)))
  logger.debug(perc0)

  // threshold
  const validator_size_threshold = Number(
    Globals.Params.validator_size_threshold
  )
  let perc1 = perc0.map(
    s => (s > validator_size_threshold ? validator_size_threshold : s)
  )
  logger.debug(perc1)

  // 1st round with threshold
  let round1 = vals.map((v, idx) =>
    award.times(perc1[idx]).dividedToIntegerBy(1)
  )
  logger.debug("round1: ", round1.map(r => r.toString(10)))

  // 2nd round for the rest
  total = round1.reduce((s, v) => {
    return s.plus(v)
  })
  let left = award.minus(total)
  let round2 = vals.map((v, idx) =>
    left.times(perc0[idx]).dividedToIntegerBy(1)
  )
  logger.debug("round2: ", round2.map(r => r.toString(10)))

  // final
  vals.forEach((v, idx) => {
    v.shares = shares[idx]
      .plus(round1[idx])
      .plus(round2[idx])
      .toString(10)
  })
  logger.debug("<- calcValAward")
  return vals
}

const calcAwards = (nodes, blocks) => {
  // block award -> validators and backups
  let blockAward = getBlockAward()
  let valAward = blockAward,
    bakAward = 0

  const max_vals = Number(Globals.Params.max_vals)
  const val_ratio = Number(Globals.Params.validators_block_award_ratio / 100)
  if (nodes.length > max_vals) {
    valAward = blockAward.times(val_ratio).dividedToIntegerBy(1)
    bakAward = blockAward.minus(valAward)
  }

  // distribute awards
  let vals = nodes.filter(n => n.state == "Validator")
  let baks = nodes.filter(n => n.state == "Backup Validator")

  nodes.forEach(v => logger.debug(v.owner_address, v.shares, v.state))
  for (i = 0; i < blocks; ++i) {
    logger.debug("validators")
    vals = calcValAward(valAward, vals)
    logger.debug("backup validators")
    if (baks.length > 0) baks = calcValAward(bakAward, baks)
  }
  vals.forEach(v => logger.debug(v.owner_address, v.shares))
  baks.forEach(v => logger.debug(v.owner_address, v.shares))
  return vals.concat(baks)
}

const calcDeleAward = (award, comp_rate, shares) => {
  logger.debug("calcDeleAward ->")
  let comp = award.times(comp_rate).dividedToIntegerBy(1)
  let dele = award.minus(comp)
  logger.debug(award.toString(10), comp.toString(10), dele.toString(10))

  let total = shares.reduce((s, v) => {
    return s.plus(v)
  })
  const toFixed = f => Number(parseFloat(f).toFixed(12))
  percs = shares.map(s => toFixed(s.div(total)))
  logger.debug(percs)

  let result = percs.map((p, idx) =>
    shares[idx].plus(dele.times(p)).dividedToIntegerBy(1)
  )
  result[2] = result[2].plus(comp)
  logger.debug(result.map(r => r.toString(10)))
  logger.debug("<- calcDeleAward")
  return result
}

const calcDeleAwards = (award, comp_rate, shares, blocks) => {
  for (i = 0; i < blocks; ++i) {
    shares = calcDeleAward(award, comp_rate, shares)
  }
  return shares
}

const delegatorAccept = (delegator, validator, deleAmount) => {
  let nonce = web3.cmt.getTransactionCount(delegator)

  let payload = {
    from: delegator,
    validatorAddress: validator,
    amount: deleAmount,
    cubeBatch: Globals.CubeBatch,
    sig: web3.cubeSign(delegator, nonce)
  }
  let r = web3.cmt.stake.delegator.accept(payload)
  expectTxSuccess(r)
}

module.exports = {
  transfer,
  getBalance,
  newContract,
  tokenTransfer,
  tokenKill,
  getTokenBalance,
  getDelegation,
  vote,
  getProposal,
  waitInterval,
  waitMultiple,
  waitBlocks,
  expectTxFail,
  expectTxSuccess,
  gasFee,
  addFakeValidators,
  removeFakeValidators,
  calcAwards,
  calcDeleAwards,
  delegatorAccept
}
