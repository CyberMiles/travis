const Web3 = require("web3-cmt")
const async = require("async")

const provider = process.argv[2]
const web3 = new Web3(new Web3.providers.HttpProvider(provider))
console.log(web3.currentProvider)

const source = "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
web3.personal.unlockAccount(source, "1234")
const dest = "0x4044e64e49c2f5392e3885c8a6519933e7f4d790"
const value = 1
const gasPrice = web3.toBigNumber(web3.toWei("5", "gwei"))

const payload = {
  from: source,
  to: dest,
  gasPrice: gasPrice,
  value: value
}

const random = (min, max) => {
  return Math.floor(Math.random() * (max - min + 1) + min)
}

let interval = setInterval(() => {
  // random 5-10
  let times = random(5, 10)
  console.log(new Date())
  async.times(
    times,
    (n, next) => {
      web3.cmt.sendTransaction(payload, (err, result) => {
        next(err, result)
      })
    },
    function(err, result) {
      if (err) {
        clearInterval(interval)
        console.log(err)
      } else {
        console.log(result)
      }
    }
  )
}, 1000)
