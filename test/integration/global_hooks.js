const expect = require("chai").expect
const fs = require("fs")
const Web3 = require("web3-cmt")
const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./utils")
const Globals = require("./global_vars")

// web3 setup before all
web3 = new Web3(new Web3.providers.HttpProvider(Settings.Providers.node1))
if (!web3 || !web3.isConnected()) throw new Error("cannot connect to server. ")

// test mode
if (web3.net.peerCount == 0) {
  Globals.TestMode = "single"
}
logger.debug("test mode: ", Globals.TestMode)

before("Set default account", function() {
  logger.info(this.test.fullTitle())
  // set default account
  web3.cmt.defaultAccount = web3.cmt.accounts[0]
})

before("Prepare 4 accounts", function() {
  logger.info(this.test.fullTitle())
  let count = web3.cmt.accounts.length
  if (count > 2) {
    // get 3 accounts. skip first 2 accounts
    Globals.Accounts = web3.cmt.accounts.slice(2, 5)
    // last account
    if (count > 7) Globals.Accounts.push(web3.cmt.accounts[count - 1])
    logger.debug("use existing accounts: ", Globals.Accounts)
  } else {
    Globals.Accounts = []
  }
  // create more accounts to get 4 in total
  let newCount = 4 - Globals.Accounts.length
  for (i = 0; i < newCount; ++i) {
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

before("Load system parameters", function(done) {
  web3.cmt.governance.getParams((err, res) => {
    if (err) {
      logger.error(err)
      done(err)
    }
    Globals.Params = res.data
    logger.debug(Globals.Params)
    done()
  })
})

before("Setup a ERC20 Smart contract called ETH", function(done) {
  logger.info(this.test.fullTitle())
  // check if contract already exists
  let tokenFile = "./ETHToken.json"
  let tokenJSON = JSON.parse(fs.readFileSync(tokenFile).toString())
  Globals.ETH.abi = tokenJSON["abi"]
  Globals.ETH.bytecode = tokenJSON["bytecode"]

  let first = Globals.ETH.contractAddress
  if (web3.cmt.getCode(first) === "0x") {
    let deployAddress = web3.cmt.accounts[0]
    Utils.newContract(
      deployAddress,
      Globals.ETH.abi,
      Globals.ETH.bytecode,
      addr => {
        Globals.ETH.contractAddress = addr
        done()
      }
    )
  } else {
    logger.debug("create new contract skipped. ")
    done()
  }
})

before("Transfer 5000000 CMT to A, B, C, D from defaultAccount", function(
  done
) {
  logger.info(this.test.fullTitle())
  let balances = Utils.getBalance()
  let arrFund = []
  for (i = 0; i < 4; ++i) {
    // 2000000 cmt should be far enough for the testing
    if (web3.fromWei(balances[i], "cmt") > 2000000) continue

    let hash = Utils.transfer(
      web3.cmt.defaultAccount,
      Globals.Accounts[i],
      web3.toWei(5000000, "cmt"),
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

module.exports = Utils
