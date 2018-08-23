let Web3 = require("web3-cmt")
let config = require("config")
let Tx = require("ethereumjs-tx")
let fs = require("fs")
let Wallet = require("ethereumjs-wallet")

let fileName = "TestToken.json"
let abi = JSON.parse(fs.readFileSync(fileName).toString())["abi"]
let bytecode = JSON.parse(fs.readFileSync(fileName).toString())["bytecode"]

let web3 = new Web3(new Web3.providers.HttpProvider(config.get("provider")))

let wallet = Wallet.fromV3(config.get("wallet"), config.get("password"))
let deployAdrress = config.get("wallet").address
var privateKey = wallet.getPrivateKey()

let contract = web3.cmt.contract(abi)
// Get contract data
let contractData = contract.new.getData({
  data: bytecode
})
let gasPrice = web3.toWei(2, "gwei")
let gasPriceHex = web3.toHex(gasPrice)
let gasLimitHex = web3.toHex(4700000)

let nonce = web3.cmt.getTransactionCount(deployAdrress)
let nonceHex = web3.toHex(nonce)

let rawTx = {
  nonce: nonceHex,
  gasPrice: gasPriceHex,
  gasLimit: gasLimitHex,
  data: contractData,
  from: deployAdrress,
  chainId: web3.net.id
}

let tx = new Tx(rawTx)
tx.sign(privateKey)
let serializedTx = tx.serialize()

web3.cmt.sendRawTransaction(
  "0x" + serializedTx.toString("hex"),
  (err, hash) => {
    if (err) {
      console.log(err)
    } else {
      console.log("contract creation tx: " + hash)
      waitForTransactionReceipt(hash)
    }
  }
)

function waitForTransactionReceipt(hash) {
  let receipt = web3.cmt.getTransactionReceipt(hash)
  // If no receipt, try again in 1s
  if (receipt == null) {
    setTimeout(() => {
      waitForTransactionReceipt(hash)
    }, 1000)
  } else {
    console.log(
      "Contract mined! address: " +
        receipt.contractAddress +
        " transactionHash: " +
        receipt.transactionHash
    )
    let tokenContract = web3.cmt.contract(abi).at(receipt.contractAddress)
    let decimal = tokenContract.decimals()
    let balance = tokenContract.balanceOf(deployAdrress)
    let tokenName = tokenContract.name()
    let tokenSymbol = tokenContract.symbol()

    console.log(
      balance.toString(),
      decimal.toString(),
      tokenSymbol + " (" + tokenName + ")"
    )
  }
}
