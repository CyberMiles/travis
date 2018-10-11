const expect = require("chai").expect
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
  balance[4] = web3.cmt.getBalance(web3.cmt.defaultAccount, "latest")
  logger.debug(`balance in wei: --> ${balance}`)
  return index == null ? balance : balance[index]
}

const newContract = function(deployAddress, abi, bytecode, cb) {
  let tokenContract = web3.cmt.contract(abi)
  let contractInstance = tokenContract.new(
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
  return contractInstance
}

const tokenTransfer = function(f, t, v, gasPrice, nonce) {
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)
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
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)
  let hash = tokenInstance.kill({ from: deployAdrress })
  logger.debug("token kill hash: ", hash)
  return hash
}

const getTokenBalance = () => {
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)

  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    balance[i] = tokenInstance.balanceOf(Globals.Accounts[i])
  }
  logger.debug(`token balance: --> ${balance}`)
  return balance
}

const getDelegation = (acc_index, pk_index, vals) => {
  let delegation = {
    delegate_amount: web3.toBigNumber(0),
    award_amount: web3.toBigNumber(0),
    withdraw_amount: web3.toBigNumber(0),
    pending_withdraw_amount: web3.toBigNumber(0),
    slash_amount: web3.toBigNumber(0),
    shares: web3.toBigNumber(0),
    voting_power: 0,
    comp_rate: 0
  }
  result = web3.cmt.stake.delegator.query(Globals.Accounts[acc_index], 0)
  if (result && result.data) {
    let data = result.data.find(d => d.pub_key && d.pub_key.value == Globals.PubKeys[pk_index])
    if (data)
      delegation = {
        delegate_amount: web3.toBigNumber(data.delegate_amount),
        award_amount: web3.toBigNumber(data.award_amount),
        withdraw_amount: web3.toBigNumber(data.withdraw_amount),
        pending_withdraw_amount: web3.toBigNumber(data.pending_withdraw_amount),
        slash_amount: web3.toBigNumber(data.slash_amount),
        voting_power: Number(data.voting_power),
        comp_rate: eval(data.comp_rate)
      }
    delegation.shares = delegation.delegate_amount
      .plus(delegation.award_amount)
      .minus(delegation.withdraw_amount)
      .minus(delegation.pending_withdraw_amount)
      .minus(delegation.slash_amount)
  }
  logger.debug(
    "delegation: --> ",
    `delegate_amount: ${delegation.delegate_amount.toString(10)}`,
    `award_amount: ${delegation.award_amount.toString(10)}`,
    `withdraw_amount: ${delegation.withdraw_amount.toString(10)}`,
    `pending_withdraw_amount: ${delegation.pending_withdraw_amount.toString(10)}`,
    `slash_amount: ${delegation.slash_amount.toString(10)}`,
    `shares: ${delegation.shares.toString(10)}`,
    `voting_power: ${delegation.voting_power}`,
    `comp_rate: ${delegation.comp_rate}`
  )
  return delegation
}

const vote = (proposalId, from, answer) => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  web3.cmt.governance.vote(
    {
      from: from,
      proposalId: proposalId,
      answer: answer
    },
    (err, res) => {
      if (err) {
        logger.error(err.message)
      } else {
        expectTxSuccess(res)
      }
    }
  )
}

const getProposal = proposalId => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.listProposals()
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
  let startingTime = Math.round(new Date().getTime() / 1000)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    let timeGone = Math.round(new Date().getTime() / 1000) - startingTime
    logger.debug(`Blocks Passed ${blocksGone}`)
    if (blocksGone == blocks) {
      logger.debug("waiting end. ")
      clearInterval(interval)
      done()
    }
    if (timeGone > Settings.WaitTimeout) {
      clearInterval(interval)
      logger.error(`Pending full after ${Settings.WaitTimeout} seconds`)
      process.exit(1)
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
      gasLimit = web3.toBigNumber(Globals.Params.declare_candidacy_gas)
      break
    case "updateCandidacy":
      gasLimit = web3.toBigNumber(Globals.Params.update_candidacy_gas)
      break
    case "proposeTransferFund":
      gasLimit = web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas)
      break
    case "proposeChangeParam":
      gasLimit = web3.toBigNumber(Globals.Params.change_params_proposal_gas)
      break
    case "proposeDeployLibEni":
      gasLimit = web3.toBigNumber(Globals.Params.deploy_libeni_proposal_gas)
      break
    case "setCompRate":
      gasLimit = web3.toBigNumber(Globals.Params.set_comp_rate_gas)
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
        let initAmount = 100000,
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
  // const inflation_rate = eval(Globals.Params.inflation_rate)
  // let cmts = web3.toWei(web3.toBigNumber(1000000000), "cmt")
  // let blocksYr = (365 * 24 * 3600) / 10
  // let blockAward = cmts.times(inflation_rate / 100).dividedToIntegerBy(blocksYr)
  let blockAward = web3.toBigNumber("25367833587011669203")
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
  const validator_size_threshold = Number(Globals.Params.validator_size_threshold)
  let perc1 = perc0.map(s => (s > validator_size_threshold ? validator_size_threshold : s))
  logger.debug(perc1)

  // 1st round with threshold
  let round1 = vals.map((v, idx) => award.times(perc1[idx]).dividedToIntegerBy(1))
  logger.debug("round1: ", round1.map(r => r.toString(10)))

  // 2nd round for the rest
  total = round1.reduce((s, v) => {
    return s.plus(v)
  })
  let left = award.minus(total)
  let round2 = vals.map((v, idx) => left.times(perc0[idx]).dividedToIntegerBy(1))
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

// n: number of delegators; s: delegator's current stake
const calcVotingPower = (n, s, p) => {
  logger.debug("n,s,p:", n, s, p)
  // no awards if less than 1000cmt
  if (parseInt(s / 1e18) < Number(Globals.Params.min_staking_amount)) {
    return 0
  }

  let s10 = 1,
    s90 = 1,
    t = 1 // simplfied.
  let r1 = Math.pow(s10 / s90, 2)
  let r2 = (t / 180 + 1).toFixed(2)
  let r3 = Math.pow(1 - 1 / (n * 4 + 1), 2)
  let r4 = parseInt((s / 1e18) * p)
  let x = r1 * r3 * r4
  let l = Math.log2(r2)
  let vp = Math.ceil(l * x)
  logger.debug("r1,r2,r3,r4,x,l,vp:", r1, r2, r3, r4, x, l, vp)
  return vp
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

  let result = percs.map((p, idx) => shares[idx].plus(dele.times(p)).dividedToIntegerBy(1))
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
    sig: cubeSign(delegator, nonce)
  }
  let r = web3.cmt.stake.delegator.accept(payload)
  expectTxSuccess(r)
}

const crypto = require("crypto")
// this private key is for testing only, use it together with cubeBatch "01"
const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCiWpvDnwYFTqgSWPlA3VO8u+Yv9r8QGlRaYZFszUZEXUQxquGl
FexMSVyFeqYjIokfPOEHHx2voqWgi3FKKlp6dkxwApP3T22y7Epqvtr+EfNybRta
15snccZy47dY4UcmYxbGWFTaL66tz22pCAbjFrxY3IxaPPIjDX+FiXdJWwIDAQAB
AoGAOc63XYz20Nbz4yyI+36S/UWOLY/W8f3eARxycmIY3eizilfE5koLDBKm/ePw
2dvHJTdBDI8Yu9vWy3Y7DWRNOHJKdcc1gGCR36cJFc4/h02zdaK+CK4eAaZLXhdK
H8DljEx6QAeRtxVLZGeYa4actY+3GeujYvkQ5QwNprchTSECQQDO4VMmLB+iIT/N
jnADYOuWKe3iLBoTKHmVfAaTRMMeHATMkpgyVzTLO7jMYCWy7+S0DL4wDNUTQv+P
Nna/hrAxAkEAyObfMAgjnW6s+CGoN+yWtdBC0LvDXDrzaT3KqmHxK2iCg2kQ9R6P
0vCvGJytuPxmIVZn54+KpKfR6ok6RJSbSwJAF+CRxDobfI7x2juyWfF5v18fgZct
e0CUp9gkuiKZkoQRWbshrc263ioKbiw6rahacR13ZfxVK1/0NwdGNVzKQQJBAJpw
QGpgF2DSz8z/sp0rFsA1lOd5L7ka6Dui8MUB/a9s68exYQPNtqpls3SsHS/zd19x
WPa9dcsV51zwmQZXZvkCQQChnQLBs6BbH6O85ePXSSbe7RUvHua6EEkmCNkIw+vT
3Jqmk4ecxCzmEv3xbzrCdgOhfjxqjrsqLLK6BH+lJsWS
-----END RSA PRIVATE KEY-----`

const cubeSign = (address, nonce) => {
  let message = address + "|" + nonce

  let hash = crypto
    .createHash("sha256")
    .update(message)
    .digest("hex")
  let signature = crypto.privateDecrypt(
    {
      key: privateKey,
      padding: crypto.constants.RSA_NO_PADDING
    },
    new Buffer(hash, "hex")
  )
  let signature_hex = signature.toString("hex")
  logger.debug("cube sig: ", signature_hex)
  return signature_hex
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
  delegatorAccept,
  cubeSign,
  calcVotingPower,
  getBlockAward
}
