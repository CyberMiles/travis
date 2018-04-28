const chai = require("chai")
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

let balance_old = new Array(4),
  balance_new = new Array(4),
  proposeId = ""

const vote = (from, answer) => {
  let r = web3.cmt.governance.vote({
    from: from,
    proposeId: proposeId,
    answer: answer
  })
  Utils.expectTxSuccess(r)
}

describe.skip("Governance Test", function() {
  before(function() {
    accounts.forEach(acc => {
      // unlock account
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    })
  })
  describe("Account #1 does not have 500 CMTs.", function() {
    before(function(done) {
      let balance = web3
        .fromWei(web3.cmt.getBalance(accounts[1]), "cmt")
        .toNumber()
      if (balance > 500) {
        web3.cmt.sendTransaction({
          from: accounts[1],
          to: accounts[0],
          value: web3.toWei(balance - 1, "cmt")
        })
        Utils.waitBlocks(done)
      } else {
        done()
      }
    })

    before(function() {
      let balance = web3
        .fromWei(web3.cmt.getBalance(accounts[1]), "cmt")
        .toNumber()
      logger.debug(`balance of accounts[1] is ${balance}`)
      expect(balance).to.be.lt(500)
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("The proposal TX returns an error. ", function() {
        let r = web3.cmt.governance.propose({
          from: accounts[0],
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxFail(r)
      })
    })
  })

  describe("Account #1 have 500 CMTs. ", function() {
    before(function(done) {
      let balance = web3
        .fromWei(web3.cmt.getBalance(accounts[1]), "cmt")
        .toNumber()
      if (balance < 500) {
        web3.cmt.sendTransaction({
          from: accounts[0],
          to: accounts[1],
          value: web3.toWei(500, "cmt")
        })
        Utils.waitBlocks(done)
      } else {
        done()
      }
    })

    before(function() {
      let balance = web3
        .fromWei(web3.cmt.getBalance(accounts[1]), "cmt")
        .toNumber()
      logger.debug(`balance of accounts[1] is ${balance}`)
      expect(balance).to.be.gt(500)
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let r = web3.cmt.governance.propose({
          from: accounts[0],
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxSuccess(r)
        // proposeId = r.
      })
      describe("Validators A, B, and C votes for the proposal. The total vote (A+B+C) now exceeds 2/3. ", function() {
        it("Verify that the 500 CMTs are transfered to account #2. ", function() {
          vote(accounts[0], "Y")
          vote(accounts[1], "Y")
          vote(accounts[2], "Y")
        })
      })
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let r = web3.cmt.governance.propose({
          from: accounts[0],
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxSuccess(r)
        // proposeId =
      })
      describe("Validator A votes for the proposal, but B and C vote against the proposal. The total vote (B+C) now exceeds 2/3.", function() {
        it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
          vote(accounts[0], "Y")
          vote(accounts[1], "N")
          vote(accounts[2], "N")
        })
      })
    })
    describe("Validator A proposes to move 500 CMTs from account #1 to #2. And he specifies a short expiration date (5 blocks).  ", function() {
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let r = web3.cmt.governance.propose({
          from: accounts[0],
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxSuccess(r)
        // proposeId =
      })
      describe("Validator A votes for the proposal, but no one else votes.", function() {
        it("Verify that the 500 CMTs are transfered back to account #1 after 5 blocks. ", function(done) {
          vote(accounts[0], "Y")
          Utils.waitBlocks(done, 5)
        })
      })
    })
  })
})
