const expect = require("chai").expect
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

describe("Transaction Test", function() {
  let balance_old = new Array(4),
    balance_new = new Array(4)
  let value = 1000, // gwei
    gasLimit = 21000,
    gasPrice = 5, // gwei
    gas = gasLimit * gasPrice

  beforeEach(function() {
    // balance before
    balance_old = Utils.getBalance()
    // unlock accounts
    accounts.forEach(acc =>
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    )
  })

  describe("Free CMT TX from A to B to C to D, and then back", function() {
    it("From A to B to C to D", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.transfer(accounts[i], accounts[i + 1], value)
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // balance after
        balance_new = Utils.getBalance()
        // check balance change
        expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(-value)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(0)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        expect(balance_new[3].minus(balance_old[3]).toNumber()).to.equal(value)

        done()
      })
    })

    it("From D to C to B to A", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.transfer(accounts[3 - i], accounts[2 - i], value)
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // balance after
        balance_new = Utils.getBalance()
        // check balance change
        expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(value)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(0)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        expect(balance_new[3].minus(balance_old[3]).toNumber()).to.equal(-value)

        done()
      })
    })
  })

  describe("Fee CMT TX from A to B to C to D, and then back", function() {
    it("From A to B to C to D", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.transfer(accounts[i], accounts[i + 1], value, gasPrice)
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // balance after
        balance_new = Utils.getBalance()
        // check balance change
        expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(
          -gas - value
        )
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(-gas)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(-gas)
        expect(balance_new[3].minus(balance_old[3]).toNumber()).to.equal(value)

        done()
      })
    })

    it("From D to C to B to A", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.transfer(
          accounts[3 - i],
          accounts[2 - i],
          value,
          gasPrice
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // balance after
        balance_new = Utils.getBalance()
        // check balance change
        expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(value)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(-gas)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(-gas)
        expect(balance_new[3].minus(balance_old[3]).toNumber()).to.equal(
          -gas - value
        )

        done()
      })
    })
  })

  describe("Send free CMT TX from A to B 3 times within 10s", function() {
    it("expect only the first one will succeed", function(done) {
      let arrHash = [],
        times = 2 // TODO: times=3
      let nonce = web3.cmt.getTransactionCount(accounts[0])
      for (i = 0; i < times; ++i) {
        let hash = Utils.transfer(accounts[0], accounts[1], value, 0, nonce++)
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(res.length).to.gt(1)
        expect(res[0]).to.not.be.null
        expect(res[0].blockNumber).to.be.gt(0)
        if (res.length === 2) {
          expect(!res[1] || res[1].blockNumber > res[0].blockNumber).to.be.true
        }
        if (res.length === 3) {
          expect(
            !res[2] ||
              (!res[1] && res[2].blockNumber > res[0].blockNumber) ||
              (res[1] && res[2].blockNumber > res[1].blockNumber)
          ).to.be.true
        }

        // balance after
        balance_new = Utils.getBalance()

        done()
      })
    })
  })

  describe("Send fee CMT TX from A to B 3 times within 10s", function() {
    it("expect all to succeed", function(done) {
      let arrHash = [],
        times = 3
      let nonce = web3.cmt.getTransactionCount(accounts[0])
      for (i = 0; i < times; ++i) {
        let hash = Utils.transfer(
          accounts[0],
          accounts[1],
          value,
          gasPrice,
          nonce++
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // balance after
        balance_new = Utils.getBalance()
        // check balance change
        expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(
          -(gas + value) * times
        )
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          value * times
        )

        done()
      })
    })
  })
})
