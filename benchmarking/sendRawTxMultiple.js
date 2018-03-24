const config = require("config")
const Web3 = require("web3")
const Wallet = require("ethereumjs-wallet")
const utils = require("./util")
const async = require("async")
const Web3pool = require("./web3pool")

const web3p = new Web3pool(config.get("providers"))
const web3 = web3p.web3
const wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))
const totalAccounts = config.get("accounts")

const walletAddress = wallet.getAddressString()

const gasPrice = web3.eth.gasPrice
const totalTxs = config.get("txs")
const balance = web3.eth.getBalance(walletAddress)
const costPerAccount = utils.calculateTransactionsPrice(gasPrice, totalTxs)
const distributingCost = utils.calculateTransactionsPrice(
  gasPrice,
  totalAccounts
)
const totalCost = distributingCost.plus(costPerAccount.times(totalAccounts))

console.log(
  `Send ${totalTxs} transactions from each account, accounts: ${totalAccounts}`
)
console.log(
  `Cost of each account txs: ${web3.fromWei(costPerAccount, "ether")}`
)
console.log(`Distributing cost: ${web3.fromWei(distributingCost, "ether")}`)
console.log(`Cost of all: ${web3.fromWei(totalCost, "ether")}`)

if (totalCost.comparedTo(balance) > 0) {
  throw new Error(`Unsufficient funds: ${web3.fromWei(balance, "ether")}`)
}

let privKey = wallet.getPrivateKey()

console.log("Generating wallets and sending funds")
let wallets = []
let nonce = web3.eth.getTransactionCount(walletAddress)
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

utils.sendRawTransactions(web3p, transactions, (err, ms) => {
  if (err) {
    console.error("Couldn't send Transactions:")
    console.error(err)
    return
  }

  // wait for 200 blocks this case. Initial Fund distribution
  utils.waitMultipleProcessedInterval(web3p, 100, 200, (err, endDate) => {
    if (err) {
      console.error(err)
      return
    }
    console.log("Distributed Funds.")

    // Generate Transactions
    let dest = config.get("address")
    let initialNonces = {}
    let txsPerAccount = config.get("txs")
    let transactions = []
    console.log(
      `Generating ${txsPerAccount} transactions for ${totalAccounts} accounts`
    )
    wallets.forEach(w => {
      let addr = w.getAddressString()
      initialNonces[addr] = web3.eth.getTransactionCount(addr)
      for (let i = 0; i < txsPerAccount; i++) {
        let tx = utils.generateRawTransaction({
          nonce: initialNonces[addr] + i,
          gasPrice: gasPrice,
          from: addr,
          to: dest,
          privKey: w.getPrivateKey()
        })
        transactions.push(tx)
      }
    })
    console.log(`Generated`)

    console.log(`Starting to send transactions in parallel`)
    const start = new Date()
    console.log("start time: ", start)
    const blockTimeout = config.get("blockTimeout")
    utils.sendRawTransactions(web3p, transactions, (err, res) => {
      if (err) {
        console.log("Error on transactions:", err)
        return
      }

      utils.waitMultipleProcessedInterval(
        web3p,
        100,
        blockTimeout,
        (err, endDate) => {
          if (err) {
            console.error("Couldn't process transactions in blocks")
            console.error(err)
            return
          }

          let sent = totalTxs * totalAccounts
          let processed = Object.keys(initialNonces).reduce((sum, addr) => {
            return (
              sum + (web3.eth.getTransactionCount(addr) - initialNonces[addr])
            )
          }, 0)
          console.log("end time: ", endDate)
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
  })
})
