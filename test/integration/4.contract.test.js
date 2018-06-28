const expect = require("chai").expect
const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Contract Test", function() {
  let token_balance_old = new Array(4),
    token_balance_new = new Array(4)
  let tokens = 1,
    gasPrice = Globals.GasPrice

  before("Transfer 1000 ETH to A, B, C, D from defaultAccount", function(done) {
    logger.info(this.test.fullTitle())

    let balances = Utils.getTokenBalance()
    let arrFund = []
    let initialFund = 1000 // 1000 ether or 1000 token
    let estimateCost = 100 // at most. not so much in fact

    for (i = 0; i < 4; ++i) {
      if (balances[i] > estimateCost) continue

      let hash = Utils.tokenTransfer(
        web3.cmt.defaultAccount,
        Globals.Accounts[i],
        initialFund,
        gasPrice
      )
      arrFund.push(hash)
    }
    if (arrFund.length > 0) {
      Utils.waitMultiple(
        arrFund,
        () => {
          Utils.getTokenBalance()
          done()
        },
        Settings.BlockTicker
      )
    } else {
      logger.debug("token fund skipped. ")
      done()
    }
  })

  beforeEach(function() {
    // token balance before
    token_balance_old = Utils.getTokenBalance()
  })

  describe("Free ETH TX from A to B to C to D, and then back", function() {
    it("from A to B to C to D", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[i],
          Globals.Accounts[i + 1],
          tokens
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(-tokens)
        expect(
          token_balance_new[1].minus(token_balance_old[1]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[2].minus(token_balance_old[2]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[3].minus(token_balance_old[3]).toNumber()
        ).to.equal(tokens)

        done()
      })
    })
    it("from D to C to B to A", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[3 - i],
          Globals.Accounts[2 - i],
          tokens
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(tokens)
        expect(
          token_balance_new[1].minus(token_balance_old[1]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[2].minus(token_balance_old[2]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[3].minus(token_balance_old[3]).toNumber()
        ).to.equal(-tokens)

        done()
      })
    })
  })

  describe("Fee ETH TX from A to B to C to D, and then back", function() {
    it("from A to B to C to D", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[i],
          Globals.Accounts[i + 1],
          tokens,
          gasPrice
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(-tokens)
        expect(
          token_balance_new[1].minus(token_balance_old[1]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[2].minus(token_balance_old[2]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[3].minus(token_balance_old[3]).toNumber()
        ).to.equal(tokens)

        done()
      })
    })
    it("from D to C to B to A", function(done) {
      let arrHash = []
      for (i = 0; i < 3; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[3 - i],
          Globals.Accounts[2 - i],
          tokens,
          gasPrice
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        expect(err).to.be.null
        expect(res.length).to.equal(3)
        expect(res).to.not.include(null)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(tokens)
        expect(
          token_balance_new[1].minus(token_balance_old[1]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[2].minus(token_balance_old[2]).toNumber()
        ).to.equal(0)
        expect(
          token_balance_new[3].minus(token_balance_old[3]).toNumber()
        ).to.equal(-tokens)

        done()
      })
    })
  })

  describe("Send free ETH TX from A to B 3 times within 10s", function() {
    it("expect only the first one will succeed", function(done) {
      let arrHash = [],
        times = 3
      for (i = 0; i < times; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[0],
          Globals.Accounts[1],
          tokens,
          0
        )
        arrHash.push(hash)
      }
      // 2nd and 3rd will fail
      expect(arrHash[1]).to.be.null
      expect(arrHash[2]).to.be.null

      Utils.waitMultiple(arrHash, (err, res) => {
        // 1st one will succeed
        expect(res.length).to.eq(1)
        expect(res[0]).to.not.be.null
        expect(res[0].blockNumber).to.be.gt(0)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(-tokens)

        done()
      })
    })
  })

  describe("Send fee ETH TX from A to B 3 times within 10s.", function() {
    it("expect all to succeed", function(done) {
      let arrHash = [],
        times = 3
      for (i = 0; i < times; ++i) {
        let hash = Utils.tokenTransfer(
          Globals.Accounts[0],
          Globals.Accounts[1],
          tokens,
          gasPrice
        )
        arrHash.push(hash)
      }

      Utils.waitMultiple(arrHash, (err, res) => {
        // all success
        expect(err).to.be.null
        expect(res.length).to.equal(times)
        expect(res).to.not.include(null)

        // token balance after
        token_balance_new = Utils.getTokenBalance()

        // check token balance change
        expect(
          token_balance_new[0].minus(token_balance_old[0]).toNumber()
        ).to.equal(-tokens * times)
        expect(
          token_balance_new[1].minus(token_balance_old[1]).toNumber()
        ).to.equal(tokens * times)

        done()
      })
    })
  })

  describe("Destroy the contract", function() {
    it("expect all to succeed", function(done) {
      let deployAdrress = web3.cmt.defaultAccount
      let hash = Utils.tokenKill(deployAdrress)

      Utils.waitInterval(hash, (err, res) => {
        expect(err).to.be.null
        expect(res).to.be.not.null

        // balance after
        token_balance_new = Utils.getTokenBalance()
        for (i = 0; i < 4; ++i) {
          expect(token_balance_new[i].toNumber()).to.eq(0)
        }
        // check code
        expect(web3.cmt.getCode(contractAddress)).to.eq("0x")

        done()
      })
    })
  })
})
