let Wallet = require("ethereumjs-wallet")
let Web3 = require("web3-cmt")
let config = require("config")
let utils = require("./utils")

let web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))
let wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))

// from,to,value and total transactions
let fromAddress = wallet.getAddressString()
let destAddress = config.get("to")
let value = 1

let totalTxs = config.get("txs")
console.log("Current block number:", web3.cmt.blockNumber)
console.log(
  `Will send ${totalTxs} transactions from ${fromAddress} to ${destAddress}`
)

// check balance
let gasPrice = web3.toBigNumber(web3.toWei("5", "gwei"))
let cost = utils.calculateTransactionsPrice(gasPrice, value, totalTxs)
let balance = web3.cmt.getBalance(fromAddress)
let endBalance = balance.minus(cost)
console.log("balance after transfer will be: ", endBalance.toString())

if (cost.comparedTo(balance) > 0) {
  let error =
    `You don't have enough money to make ${totalTxs} transactions, ` +
    `it needs ${cost} wei, but you have ${balance}`
  throw new Error(error)
}

// generate raw transactions
console.log(`Generating ${totalTxs} transactions`)
let initialNonce = web3.cmt.getTransactionCount(fromAddress)
let privKey = wallet.getPrivateKey()
let transactions = []
for (let i = 0; i < totalTxs; i++) {
  let nonce = i + initialNonce
  let tx = utils.generateRawTransaction({
    from: fromAddress,
    to: destAddress,
    privKey: privKey,
    nonce: nonce,
    gasPrice: gasPrice,
    value: value
  })

  transactions.push(tx)
}
console.log("done.")

// Send transactions
console.log(`Starting to send raw transactions in parallel`)
let startingBlock = web3.cmt.blockNumber
let start = new Date()
console.log(`start time: ${start.toISOString()}, block: ${startingBlock}`)

utils.sendRawTransactions(web3, transactions, (err, ms) => {
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
