const chai = require("chai")
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Governance Test", function() {
  let proposalId = ""
  let balance_old, balance_new

  describe("Account #1 does not have 500 CMTs.", function() {
    before(function() {
      let balance = Utils.getBalance(1)
      if (web3.fromWei(balance, "cmt") > 500) {
        Utils.transfer(Globals.Accounts[1], web3.cmt.defaultAccount, balance)
      }
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("The proposal TX returns an error. ", function() {
        let r = web3.cmt.governance.propose({
          from: web3.cmt.defaultAccount,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxFail(r)
      })
    })
    after(function(done) {
      Utils.waitBlocks(done)
    })
  })

  describe("Account #1 have enough CMTs. ", function() {
    before(function(done) {
      let balance = Utils.getBalance(1)
      if (web3.fromWei(balance, "cmt") < 5000) {
        let hash = Utils.transfer(
          web3.cmt.defaultAccount,
          Globals.Accounts[1],
          web3.toWei(5000, "cmt")
        )
        Utils.waitInterval(hash, (err, res) => {
          expect(err).to.be.null
          expect(res).to.be.not.null
          done()
        })
      } else {
        done()
      }
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        let r = web3.cmt.governance.propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test"
        })
        Utils.expectTxSuccess(r)
        proposalId = r.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Amount).to.equal(amount)
        expect(p.Result).to.be.empty
        let elapse = p.ExpireBlockHeight - p.BlockHeight
        // default to 7 days
        expect(elapse).to.equal(Globals.ProposalExpires)

        // balance after
        balance_new = Utils.getBalance()
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          Number(-amount)
        )
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
      })
      describe("Validators A, B, and C votes for the proposal. The total vote (A+B+C) now exceeds 2/3. ", function() {
        it("Verify that the 500 CMTs are transfered to account #2. ", function() {
          if (Globals.TestMode == "cluster") {
            Utils.vote(proposalId, Globals.Accounts[0], "Y")
            Utils.vote(proposalId, Globals.Accounts[1], "Y")
            Utils.vote(proposalId, Globals.Accounts[2], "Y")
          } else {
            Utils.vote(proposalId, web3.cmt.defaultAccount, "Y")
          }
          // check proposal
          let p = Utils.getProposal(proposalId)
          expect(p.Result).to.equal("Approved")
          // balance after
          balance_new = Utils.getBalance()
          expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
            Number(web3.toWei(-500, "cmt"))
          )
          expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(
            Number(web3.toWei(500, "cmt"))
          )
        })
      })
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        let r = web3.cmt.governance.propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test"
        })
        Utils.expectTxSuccess(r)
        proposalId = r.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Amount).to.equal(amount)
        expect(p.Result).to.be.empty
        let elapse = p.ExpireBlockHeight - p.BlockHeight
        // default to 7 days
        expect(elapse).to.equal(Globals.ProposalExpires)

        // balance after
        balance_new = Utils.getBalance()
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          Number(-amount)
        )
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
      })
      describe("Validator A votes for the proposal, but defaultAccount, B and C vote against the proposal. The total vote (default+B+C) now exceeds 2/3.", function() {
        it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
          Utils.vote(proposalId, Globals.Accounts[0], "Y")
          Utils.vote(proposalId, web3.cmt.defaultAccount, "N")
          if (Globals.TestMode == "cluster") {
            Utils.vote(proposalId, Globals.Accounts[1], "N")
            Utils.vote(proposalId, Globals.Accounts[2], "N")
          }
          // check proposal
          let p = Utils.getProposal(proposalId)
          expect(p.Result).to.equal("Rejected")
          // balance after
          balance_new = Utils.getBalance()
          expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(0)
          expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        })
      })
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. And he specifies a short expiration date (5 blocks).  ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt"),
          expire = 5
        let r = web3.cmt.governance.propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test",
          expire: expire
        })
        Utils.expectTxSuccess(r)
        proposalId = r.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Amount).to.equal(amount)
        expect(p.Result).to.be.empty
        let elapse = p.ExpireBlockHeight - p.BlockHeight
        expect(elapse).to.equal(expire)

        // balance after
        balance_new = Utils.getBalance()
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          Number(-amount)
        )
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
      })

      describe("Validator A votes for the proposal, but no one else votes.", function() {
        it("Wait for 5 blocks", function(done) {
          Utils.vote(proposalId, Globals.Accounts[0], "Y")
          Utils.waitBlocks(done, 5)
        })

        it("Verify that the 500 CMTs are transfered back to account #1 after 5 blocks. ", function() {
          // check proposal
          let p = Utils.getProposal(proposalId)
          expect(p.Result).to.equal("Expired")
          // balance after
          balance_new = Utils.getBalance()
          expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(0)
          expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        })
      })
    })
  })
})
