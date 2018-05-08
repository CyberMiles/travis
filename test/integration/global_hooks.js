const expect = require("chai").expect
const Web3 = require("web3-cmt")
const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./utils")

const initialFund = 5000 // 5000 cmt or 5000 token
const estimateCost = 2000 // at most. not so much in fact
const gasPrice = 5 //gwei

before("web3 setup", function() {
  logger.info(this.test.fullTitle())
  web3 = new Web3(new Web3.providers.HttpProvider(Settings.Providers.node1))
  if (!web3 || !web3.isConnected())
    throw new Error("cannot connect to server. ")

  // set default account
  web3.cmt.defaultAccount = web3.cmt.accounts[0]
  web3.personal.unlockAccount(web3.cmt.defaultAccount, Settings.Passphrase)
})

before("Prepare 4 accounts", function() {
  logger.info(this.test.fullTitle())
  // get or create 4 accounts. skip first 2 accounts
  let count = web3.cmt.accounts.length
  if (count > 2) {
    accounts = web3.cmt.accounts.slice(2, 6)
    logger.debug("use existing accounts: ", accounts)
  } else {
    accounts = []
  }
  for (i = 0; i < 6 - count; ++i) {
    let acc = web3.personal.newAccount(Settings.Passphrase)
    logger.debug("new account created: ", acc)
    accounts.push(acc)
  }
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

before("Transfer 2000 CMT to A, B, C, D from defaultAccount", function(done) {
  logger.info(this.test.fullTitle())

  let balances = Utils.getBalance()
  let arrFund = []
  for (i = 0; i < 4; ++i) {
    if (web3.fromWei(balances[i], "gwei") > estimateCost) continue

    let hash = Utils.transfer(
      web3.cmt.defaultAccount,
      accounts[i],
      web3.toWei(initialFund, "ether"),
      gasPrice
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

module.exports = Utils
