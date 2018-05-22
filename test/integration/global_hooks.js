const expect = require("chai").expect
const Web3 = require("web3-cmt")
const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./utils")
const Globals = require("./global_vars")

// web3 setup before all
web3 = new Web3(new Web3.providers.HttpProvider(Settings.Providers.node1))
if (!web3 || !web3.isConnected()) throw new Error("cannot connect to server. ")

before("Set default account", function() {
  logger.info(this.test.fullTitle())
  // set default account
  web3.cmt.defaultAccount = web3.cmt.accounts[0]
})

before("Prepare 4 accounts", function() {
  logger.info(this.test.fullTitle())
  // get or create 4 accounts. skip first 2 accounts
  let count = web3.cmt.accounts.length
  if (count > 2) {
    Globals.Accounts = web3.cmt.accounts.slice(2, 6)
    logger.debug("use existing accounts: ", Globals.Accounts)
  } else {
    Globals.Accounts = []
  }
  for (i = 0; i < 6 - count; ++i) {
    let acc = web3.personal.newAccount(Settings.Passphrase)
    logger.debug("new account created: ", acc)
    Globals.Accounts.push(acc)
  }
})
before("Unlock all accounts", function() {
  logger.info(this.test.fullTitle())
  // unlock account
  web3.personal.unlockAccount(
    web3.cmt.defaultAccount,
    Settings.Passphrase,
    3000
  )
  Globals.Accounts.forEach(acc => {
    web3.personal.unlockAccount(acc, Settings.Passphrase, 3000)
  })
})

before("Setup a ERC20 Smart contract called ETH", function(done) {
  logger.info(this.test.fullTitle())
  // check if contract already exists
  let first = "b6b29ef90120bec597939e0eda6b8a9164f75deb"
  if (web3.cmt.getCode(first) === "0x") {
    let deployAddress = web3.cmt.accounts[0]
    Utils.newContract(deployAddress, addr => {
      contractAddress = addr
      done()
    })
  } else {
    contractAddress = first
    logger.debug("create new contract skipped. ")
    done()
  }
})

before("Transfer 5000 CMT to A, B, C, D from defaultAccount", function(done) {
  logger.info(this.test.fullTitle())
  let balances = Utils.getBalance()
  let arrFund = []
  for (i = 0; i < 4; ++i) {
    // 2000 cmt should be far enough for the testing
    if (web3.fromWei(balances[i], "ether") > 2000) continue

    let hash = Utils.transfer(
      web3.cmt.defaultAccount,
      Globals.Accounts[i],
      web3.toWei(5000, "ether"),
      5 //gwei
    )
    arrFund.push(hash)
  }
  if (arrFund.length > 0) {
    Utils.waitMultiple(
      arrFund,
      () => {
        Utils.getBalance()
        done()
      },
      Settings.BlockTicker
    )
  } else {
    logger.debug("fund skipped. ")
    done()
  }
})

before("Add some fake validators if it's a single node", function() {
  logger.info(this.test.fullTitle())
  if (web3.net.peerCount == 0) {
    Globals.TestMode = "single"

    let result = web3.cmt.stake.queryValidators()
    let valsToAdd = 4 - result.data.length

    if (valsToAdd > 0) {
      Globals.Accounts.forEach((acc, idx) => {
        if (idx >= valsToAdd) return
        let initAmount = 1000,
          cut = "0.8"
        let payload = {
          from: acc,
          pubKey: Globals.PubKeys[idx],
          maxAmount: web3.toWei(initAmount, "cmt"),
          cut: cut
        }
        let r = web3.cmt.stake.declareCandidacy(payload)
        Utils.expectTxSuccess(r)
        logger.debug(`validator ${acc} added, max_amount: ${initAmount} cmt`)
      })
    }
  }
  logger.debug("test mode: ", Globals.TestMode)
})

after("Remove fake validators for single node", function() {
  logger.info(this.test.fullTitle())
  if (web3.net.peerCount == 0) {
    let result = web3.cmt.stake.queryValidators()
    result.data.forEach((val, idx) => {
      // skip the first one
      if (idx == 0) return
      // remove all others
      let acc = val.owner_address
      let r = web3.cmt.stake.withdrawCandidacy({ from: acc })
      Utils.expectTxSuccess(r)
      logger.debug(`validator ${acc} removed`)
    })
  }
})

module.exports = Utils
