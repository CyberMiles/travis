let Wallet = require("ethereumjs-wallet")
let async = require("async")
let Web3 = require("web3-cmt")
let config = require("config")
let utils = require("./utils")

let web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))
let wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))
let walletAddress = wallet.getAddressString()
let chainId = web3.net.id

// to,value and total accounts * transactions
let destAddress = config.get("to")
let value = config.get("value")
let totalAccounts = config.get("concurrency")
let txsPerAccount = parseInt(config.get("txs") / totalAccounts)
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

  let tx = utils.generateTransaction({
    gasPrice: gasPrice,
    from: walletAddress,
    to: wallets[i].getAddressString(),
    value: costPerAccount
  })
  transactions.push(tx)
}

// unlock walletAddress
console.log(`Unlock account ${walletAddress}`)
web3.personal.unlockAccount(walletAddress, config.get("password"))
console.log("done.")

// send fund transactions
let startingBlock = web3.cmt.blockNumber
let initialNonce = web3.cmt.getTransactionCount(walletAddress)
utils.sendTransactions(web3, transactions, (err, ms) => {
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
      let transactions = {},
        accounts = []
      console.log(
        `Generating ${txsPerAccount} transactions for ${totalAccounts} accounts`
      )
      wallets.forEach(w => {
        let addr = w.getAddressString()
        accounts.push(addr)
        initialNonces[addr] = web3.cmt.getTransactionCount(addr)
        let txs = []
        for (let i = 0; i < txsPerAccount; i++) {
          let tx = utils.generateRawTransaction(
            {
              nonce: initialNonces[addr] + i,
              gasPrice: gasPrice,
              from: addr,
              to: destAddress,
              privKey: w.getPrivateKey(),
              value: value
            },
            chainId
          )
          txs.push(tx)
        }
        transactions[addr] = txs
      })
      console.log(`done.`)

      // send transactions
      console.log(`Starting to send transactions in parallel`)
      startingBlock = web3.cmt.blockNumber
      let start = new Date()
      console.log(`start time: ${start.toISOString()}, block: ${startingBlock}`)

      async.parallelLimit(
        accounts.map(addr => {
          return utils.sendRawTransactionsSeries.bind(
            null,
            web3,
            transactions[addr]
          )
        }),
        accounts.length,
        (err, res) => {
          if (err) {
            console.log("Error when sending transactions:", err)
            return
          }

          // wait for transactions finish processing
          let sent = txsPerAccount * totalAccounts
          utils.waitMultipleProcessed(
            web3,
            startingBlock,
            accounts,
            sent,
            (err, res) => {
              if (err) {
                console.error("Couldn't process transactions in blocks")
                console.error(err)
                return
              }

              let timePassed = (res.endDate - start) / 1000
              let perSecond = res.processed / timePassed

              console.log("end time: ", res.endDate)
              console.log(
                `Processed ${res.processed} of ${sent} transactions ` +
                  `from ${totalAccounts} account in ${timePassed}s, ${perSecond} tx/s`
              )
            }
          )
        }
      )
    }
  )
})
