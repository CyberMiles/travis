let Wallet = require("ethereumjs-wallet")
let Web3 = require("web3-cmt")
let config = require("config")
let utils = require("./utils")

let web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))
let wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))
let walletAddress = wallet.getAddressString()

// to,value and total accounts * transactions
let destAddress = config.get("to")
let value = config.get("value")
let totalAccounts = 5
let txsPerAccount = config.get("txs")
console.log("Current block number:", web3.cmt.blockNumber)
console.log(
  `Will send ${txsPerAccount} transactions from each account, total accounts: ${totalAccounts}`
)

// check balance
let gasPrice = web3.toBigNumber(web3.toWei("5", "gwei"))
let balance = web3.cmt.getBalance(walletAddress)
let costPerAccount = utils.calculateTransactionsPrice(
  gasPrice,
  value,
  txsPerAccount
)
let totalCost = utils.calculateTransactionsPrice(
  gasPrice,
  costPerAccount,
  totalAccounts
)
let endBalance = balance.minus(totalCost)

console.log(`Cost of each account txs: ${costPerAccount}`)
console.log(`Cost of all: ${totalCost}`)
if (totalCost.comparedTo(balance) > 0) {
  let error =
    `${walletAddress} don't have enough money to fund ${totalAccounts} to make ${txsPerAccount} transactions, ` +
    `it needs ${totalCost} wei, but you have ${balance}`
  throw new Error(error)
}
console.log(
  `balance after transfer of account ${walletAddress} will be: ${endBalance}`
)

// generate fund transactions
console.log("Generating wallets and sending funds")
let wallets = []
let nonce = web3.cmt.getTransactionCount(walletAddress)
let privKey = wallet.getPrivateKey()
let transactions = []
for (let i = 0; i < totalAccounts; i++) {
  wallets.push(Wallet.generate())

  let tx = utils.generateRawTransaction({
    nonce: nonce + i,
    gasPrice: gasPrice,
    from: walletAddress,
    to: wallets[i].getAddressString(),
    value: costPerAccount,
    privKey: privKey
  })
  transactions.push(tx)
}

// send fund transactions
let startingBlock = web3.cmt.blockNumber
let initialNonce = web3.cmt.getTransactionCount(walletAddress)
utils.sendRawTransactions(web3, transactions, (err, ms) => {
  if (err) {
    console.error("Couldn't send Transactions:")
    console.error(err)
    return
  }

  // Initial Fund distribution
  utils.waitProcessedInterval(
    web3,
    startingBlock,
    walletAddress,
    initialNonce,
    totalAccounts,
    (err, endDate) => {
      if (err) {
        console.error(err)
        return
      }
      console.log("Distributed Funds.")

      // Generate Transactions
      let initialNonces = {}
      let transactions = [],
        accounts = []
      console.log(
        `Generating ${txsPerAccount} transactions for ${totalAccounts} accounts`
      )
      wallets.forEach(w => {
        let addr = w.getAddressString()
        accounts.push(addr)
        initialNonces[addr] = web3.cmt.getTransactionCount(addr)
        for (let i = 0; i < txsPerAccount; i++) {
          let tx = utils.generateRawTransaction({
            nonce: initialNonces[addr] + i,
            gasPrice: gasPrice,
            from: addr,
            to: destAddress,
            privKey: w.getPrivateKey(),
            value: value
          })
          transactions.push(tx)
        }
      })
      console.log(`done.`)

      // send transactions
      console.log(`Starting to send transactions in parallel`)
      startingBlock = web3.cmt.blockNumber
      let start = new Date()
      console.log(`start time: ${start.toISOString()}, block: ${startingBlock}`)
      utils.sendRawTransactions(web3, transactions, (err, res) => {
        if (err) {
          console.log("Error on transactions:", err)
          return
        }

        // wait for transactions finish processing
        utils.waitMultipleProcessed(
          web3,
          startingBlock,
          accounts,
          txsPerAccount,
          (err, endDate) => {
            if (err) {
              console.error("Couldn't process transactions in blocks")
              console.error(err)
              return
            }

            let sent = txsPerAccount * totalAccounts
            let processed = Object.keys(initialNonces).reduce((sum, addr) => {
              return (
                sum + (web3.cmt.getTransactionCount(addr) - initialNonces[addr])
              )
            }, 0)
            let timePassed = (endDate - start) / 1000
            let perSecond = processed / timePassed

            console.log("end time: ", endDate)
            console.log(
              `Processed ${processed} of ${sent} transactions ` +
                `from ${totalAccounts} account in ${timePassed}s, ${perSecond} tx/s`
            )
          }
        )
      })
    }
  )
})
