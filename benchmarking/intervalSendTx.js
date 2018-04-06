const Web3 = require("web3-cmt")
const async = require("async")

const provider = process.argv[2]

const source = "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
const dest = "0x4044e64e49c2f5392e3885c8a6519933e7f4d790"
const value = 1
const gasPrice = 5000000000

const payload = {
  from: source,
  to: dest,
  gasPrice: gasPrice,
  value: value
}

const random = (min, max) => {
  return Math.floor(Math.random() * (max - min + 1) + min)
}

let web3 = null
const connectWeb3 = () => {
  if (!web3) {
    web3 = new Web3(new Web3.providers.HttpProvider(provider))
    console.log(web3.currentProvider)
  }

  if (!web3.isConnected()) {
    try {
      web3 = new Web3(new Web3.providers.HttpProvider(provider))
    } catch (e) {
      console.log(e)
    }
  }
  if (web3.isConnected()) {
    web3.personal.unlockAccount(source, "1234")
  } else {
    console.log("try to connect", web3.currentProvider)
  }
}

let interval = setInterval(() => {
  if (!web3 || !web3.isConnected()) {
    connectWeb3()
    return
  }
  // random 5-10
  let times = random(5, 10)
  console.log(new Date(), times)
  async.times(
    times,
    (n, next) => {
      web3.cmt.sendTransaction(payload, (err, result) => {
        next(err, result)
      })
    },
    function(err, result) {
      if (err) {
        console.log(err)
        connectWeb3()
      }
    }
  )
}, 1000)
