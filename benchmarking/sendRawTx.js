const config = require("config")
const Wallet = require("ethereumjs-wallet")
const Web3 = require("web3-cmt")
const utils = require("./utils")

const web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))
const wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))

const walletAddress = wallet.getAddressString()
const initialNonce = web3.cmt.getTransactionCount(walletAddress)
const totalTxs = config.get("txs")
const blockTimeout = config.get("blockTimeout")

console.log("Current block number:", web3.cmt.blockNumber)
console.log(
  `Will send ${totalTxs} transactions and wait for ${blockTimeout} blocks`
)

let privKey = wallet.getPrivateKey()
let dest = config.get("address")
let value = 1
let gasPrice = web3.cmt.gasPrice

let cost = utils.calculateTransactionsPrice(gasPrice, value, totalTxs)
let balance = web3.cmt.getBalance(walletAddress)
let endBalance = balance.minus(cost)

if (cost.comparedTo(balance) > 0) {
  let error =
    `You don't have enough money to make ${totalTxs} transactions, ` +
    `it needs ${cost} wei, but you have ${balance}`
  throw new Error(error)
}

console.log(`Generating ${totalTxs} transactions`)
let transactions = []
for (let i = 0; i < totalTxs; i++) {
  let nonce = i + initialNonce
  let tx = utils.generateRawTransaction({
    from: walletAddress,
    to: dest,
    privKey: privKey,
    nonce: nonce,
    gasPrice: gasPrice,
    value: value
  })

  transactions.push(tx)
}
console.log("Generated.")

// Send transactions
console.log(`Starting to send raw transactions in parallel`)
const start = new Date()
console.log("start time: ", start)
utils.sendRawTransactions(web3, transactions, (err, ms) => {
  if (err) {
    console.error("Couldn't send Transactions:")
    console.error(err)
    return
  }

  utils.waitProcessedInterval(
    web3,
    walletAddress,
    endBalance,
    (err, endDate) => {
      if (err) {
        console.error("Couldn't process transactions in blocks")
        console.error(err)
        return
      }

      let sent = transactions.length
      let processed = web3.cmt.getTransactionCount(walletAddress) - initialNonce
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
