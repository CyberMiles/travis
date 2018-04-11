let Web3 = require("web3-cmt")
let config = require("config")
let utils = require("./utils")

let web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))

// from,to,value, contract address and total transactions
let fromAddress = config.get("wallet").address
let destAddress = config.get("to")
let value = config.get("value")
const contractAddress = config.get("contractAddress")

let totalTxs = config.get("txs")
console.log("Current block number:", web3.cmt.blockNumber)
console.log(
  `Will send ${totalTxs} token transactions from ${fromAddress} to ${destAddress}`
)

// check balance
let gasPrice = web3.toBigNumber(web3.toWei("5", "gwei"))
let cost = utils.calculateTokenTransactionsPrice(gasPrice, totalTxs)
let balance = web3.cmt.getBalance(fromAddress)
let endBalance = balance.minus(cost)
console.log("balance after transfer will be(estimate): ", endBalance.toString())

if (cost.comparedTo(balance) > 0) {
  let error =
    `You don't have enough money to make ${totalTxs} transactions, ` +
    `it needs ${cost} wei, but you have ${balance}`
  throw new Error(error)
}

// check token balance
let balanceToken = utils.getTokenBalance(web3, fromAddress)
let costToken = web3.toBigNumber(value).times(totalTxs)
let endBalanceToken = balanceToken.minus(costToken)
console.log("token balance before transfer: ", balanceToken.toString())
console.log(
  "token balance after transfer will be: ",
  endBalanceToken.toString()
)

if (costToken.comparedTo(balanceToken) > 0) {
  let error =
    `You don't have enough token to make ${totalTxs} transactions, ` +
    `it needs ${costToken} wei, but you have ${balanceToken}`
  throw new Error(error)
}

// generate transactions
console.log(`Generating ${totalTxs} transactions`)
let transactions = []
for (let i = 0; i < totalTxs; i++) {
  let tx = {
    to: destAddress,
    value: value
  }

  transactions.push(tx)
}
console.log("done.")

// token instance
const fs = require("fs")
const abi = JSON.parse(fs.readFileSync("TestToken.json").toString())["abi"]
const tokenContract = web3.cmt.contract(abi)
const tokenInstance = tokenContract.at(contractAddress)
web3.cmt.defaultAccount = fromAddress

// unlock fromAddress
console.log(`Unlock account ${fromAddress}`)
web3.personal.unlockAccount(fromAddress, config.get("password"))
console.log("done.")

// Send transactions
console.log(`Starting to send transactions in parallel`)
let startingBlock = web3.cmt.blockNumber
let initialNonce = web3.cmt.getTransactionCount(fromAddress)
let start = new Date()
console.log(`start time: ${start.toISOString()}, block: ${startingBlock}`)

utils.tokenTransfer(web3, tokenInstance, transactions, (err, ms) => {
  if (err) {
    console.error("Couldn't send Transactions:")
    console.error(err)
    return
  }

  // wait for transactions finish processing
  utils.waitProcessedInterval(
    web3,
    startingBlock,
    fromAddress,
    initialNonce,
    totalTxs,
    (err, endDate) => {
      if (err) {
        console.error("Couldn't process transactions in blocks")
        console.error(err)
        return
      }

      let sent = transactions.length
      let processed = web3.cmt.getTransactionCount(fromAddress) - initialNonce
      let timePassed = (endDate - start) / 1000
      let perSecond = processed / timePassed

      console.log("end time: ", endDate)
      console.log(
        `Processed ${processed} of ${sent} transactions ` +
          `from one account in ${timePassed}s, ${perSecond} tx/s`
      )
    }
  )
})
