const chai = require("chai")
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

let balance_old = new Array(4),
  balance_new = new Array(4),
  proposalId = ""

const propose = (
  { signer, transferFrom, transferTo, amount, reason, expire = 0 },
  expectSuccess = true
) => {
  let r = web3.cmt.governance.propose({
    signer,
    transferFrom,
    transferTo,
    amount,
    reason,
    expire
  })

  if (expectSuccess) {
    Utils.expectTxSuccess(r)
    proposalId = r.deliver_tx.data

    // check proposal
    let p = getProposal()
    expect(p.Amount).to.equal(amount)
    expect(p.Result).to.be.empty
    let elapse = p.ExpireBlockHeight - p.BlockHeight
    if (expire > 0) {
      expect(elapse).to.equal(expire)
    } else {
      // default to 7 days
      expect(elapse).to.equal(7 * 24 * 60 * 60 / 10)
    }

    // balance after
    balance_new = Utils.getBalance()
    expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
      Number(-amount)
    )
    expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(Number(0))
  } else {
    Utils.expectTxFail(r)
  }
}

const vote = (from, answer) => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.vote({
    from: from,
    proposalId: proposalId,
    answer: answer
  })
  Utils.expectTxSuccess(r)
}

const getProposal = () => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.queryProposals()
  expect(r.data.length).to.be.above(0)
  if (r.data.length > 0) {
    proposal = r.data.filter(d => d.Id == proposalId)
    expect(proposal.length).to.equal(1)
    return proposal[0]
  }
  return {}
}

describe("Governance Test", function() {
  before(function() {
    // unlock account
    web3.personal.unlockAccount(web3.cmt.defaultAccount, Settings.Passphrase)
    accounts.forEach(acc => {
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
          to: web3.cmt.defaultAccount,
          value: web3.toWei(balance - 1, "cmt")
        })
        Utils.waitBlocks(done)
      } else {
        done()
      }
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("The proposal TX returns an error. ", function() {
        propose(
          {
            signer: web3.cmt.defaultAccount,
            transferFrom: accounts[1],
            transferTo: accounts[2],
            amount: web3.toWei(500, "cmt"),
            reason: "Governance test"
          },
          false
        )
      })
    })
  })

  describe("Account #1 have 500 CMTs. ", function() {
    before(function(done) {
      let balance = web3
        .fromWei(web3.cmt.getBalance(accounts[1]), "cmt")
        .toNumber()
      if (balance < 5000) {
        web3.cmt.sendTransaction({
          from: web3.cmt.defaultAccount,
          to: accounts[1],
          value: web3.toWei(5000 - balance, "cmt")
        })
        Utils.waitBlocks(done)
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
        propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
      })
      describe("Validators A, B, and C votes for the proposal. The total vote (A+B+C) now exceeds 2/3. ", function() {
        it("Verify that the 500 CMTs are transfered to account #2. ", function() {
          // vote(accounts[0], "Y")
          // vote(accounts[1], "Y")
          // vote(accounts[2], "Y")
          vote(web3.cmt.defaultAccount, "Y")
          // check proposal
          let p = getProposal()
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
        propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
      })
      describe("Validator A votes for the proposal, but B and C vote against the proposal. The total vote (B+C) now exceeds 2/3.", function() {
        it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
          // vote(accounts[0], "Y")
          // vote(accounts[1], "N")
          // vote(accounts[2], "N")
          vote(web3.cmt.defaultAccount, "N")
          // check proposal
          let p = getProposal()
          expect(p.Result).to.equal("Rejected")
          // balance after
          balance_new = Utils.getBalance()
          expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
            Number(web3.toWei(0, "cmt"))
          )
          expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(
            Number(web3.toWei(0, "cmt"))
          )
        })
      })
    })

    describe("Validator A proposes to move 500 CMTs from account #1 to #2. And he specifies a short expiration date (5 blocks).  ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        propose({
          signer: web3.cmt.defaultAccount,
          transferFrom: accounts[1],
          transferTo: accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test",
          expire: 5
        })
      })

      describe("Validator A votes for the proposal, but no one else votes.", function() {
        it("Wait for 5 blocks", function(done) {
          // vote(web3.cmt.defaultAccount, "Y")
          Utils.waitBlocks(done, 5)
        })

        it("Verify that the 500 CMTs are transfered back to account #1 after 5 blocks. ", function() {
          // check proposal
          let p = getProposal()
          expect(p.Result).to.equal("Expired")
          // balance after
          balance_new = Utils.getBalance()
          expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
            Number(web3.toWei(0, "cmt"))
          )
          expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(
            Number(web3.toWei(0, "cmt"))
          )
        })
      })
    })
  })
})
