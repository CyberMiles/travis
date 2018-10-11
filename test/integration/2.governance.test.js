const chai = require("chai")
const expect = chai.expect

const logger = require("./logger")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

let V = web3.cmt.defaultAccount
describe("Governance Test", function() {
  let proposalId = ""
  let balance_old, balance_new, tx_result
  const EXPIRE_BLOCKS = 5

  describe("Account #1 does not have 500 CMTs.", function() {
    before(function() {
      balance = Utils.getBalance(1)
      Utils.transfer(Globals.Accounts[1], web3.cmt.defaultAccount, balance)
    })
    after(function() {
      Utils.transfer(web3.cmt.defaultAccount, Globals.Accounts[1], balance)
      tx_result = web3.cmt.stake.validator.list()
      logger.debug(tx_result.data)
    })

    describe("Validator V proposes to move 500 CMTs from account #1 to #2. ", function() {
      it("The proposal TX returns an error. ", function() {
        tx_result = web3.cmt.governance.proposeRecoverFund({
          from: V,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: web3.toWei(500, "cmt"),
          reason: "Governance test"
        })
        Utils.expectTxFail(tx_result)
      })
    })
  })

  describe("Account #1 have enough CMTs. ", function() {
    before(function(done) {
      let balance = Utils.getBalance(1)
      if (web3.fromWei(balance, "cmt") < 500) {
        let hash = Utils.transfer(
          web3.cmt.defaultAccount,
          Globals.Accounts[1],
          web3.toWei(500, "cmt")
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

    let old_params, new_params
    describe("Change System Parameters. ", function() {
      before(function() {
        // current system parameters
        old_params = web3.cmt.governance.getParams()
        logger.debug(old_params)
        // balance before
        balance_old = Utils.getBalance()
      })

      it("Validators V propose to double max_slash_blocks. ", function() {
        tx_result = web3.cmt.governance.proposeChangeParam({
          from: V,
          name: "max_slash_blocks",
          value: (old_params.data.max_slash_blocks * 2).toString()
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data
      })
      it("max_slash_blocks won't change before vote. ", function() {
        new_params = web3.cmt.governance.getParams()
        expect(new_params.data.max_slash_blocks).to.equal(old_params.data.max_slash_blocks)
      })
      it("vote the proposal. ", function(done) {
        // vote
        if (Globals.TestMode == "cluster") {
          Utils.vote(proposalId, Globals.Accounts[0], "Y")
          Utils.vote(proposalId, Globals.Accounts[1], "Y")
          Utils.vote(proposalId, Globals.Accounts[2], "Y")
        } else {
          Utils.vote(proposalId, web3.cmt.defaultAccount, "Y")
        }
        Utils.waitBlocks(done, 2)
      })
      it("Proposal passed. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Result).to.equal("Approved")

        it("Verify gasfee. ", function() {
          // balance after
          balance_new = Utils.getBalance()
          let gasFee = Utils.gasFee("proposeChangeParam")
          expect(balance_new[4].minus(balance_old[4]).toNumber()).to.eq(-gasFee.toNumber())
        })
      })
      it("Verify the max_slash_blocks is doubled. ", function() {
        new_params = web3.cmt.governance.getParams()
        logger.debug(new_params)
        expect(new_params.data.max_slash_blocks).to.equal(old_params.data.max_slash_blocks * 2)
        Globals.Params.max_slash_blocks = old_params.data.max_slash_blocks * 2
      })
    })

    describe("Validator V proposes to move 500 CMTs from account #1 to #2. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        tx_result = web3.cmt.governance.proposeRecoverFund({
          from: V,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test"
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.amount).to.equal(amount)
        expect(p.Result).to.be.empty

        // default to 7 days(7*24*360 blocks)
        let elapse = p.ExpireBlockHeight - p.BlockHeight
        expect(elapse).to.equal(Globals.Params.proposal_expire_period)

        // balance after
        balance_new = Utils.getBalance()
        let gasFee = Utils.gasFee("proposeTransferFund")
        expect(balance_new[4].minus(balance_old[4]).toNumber()).to.eq(-gasFee.toNumber())
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(Number(-amount))
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        // check deliver tx tx_result
        // let tag = tx_result.deliver_tx.tags.find(
        //   t => t.key == Globals.GasFeeKey
        // )
        // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
        //   gasFee.toString()
        // )
        // expect(tx_result.deliver_tx.gasUsed).to.eq(
        //   web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas).toString()
        // )
      })
      describe("Validators V, B, and C votes for the proposal. The total vote (V+B+C) now exceeds 2/3. ", function() {
        before(function(done) {
          Utils.vote(proposalId, V, "Y")
          if (Globals.TestMode == "cluster") {
            Utils.vote(proposalId, Globals.Accounts[1], "Y")
            Utils.vote(proposalId, Globals.Accounts[2], "Y")
          }
          Utils.waitBlocks(done, 1)
        })
        it("Verify that the 500 CMTs are transfered to account #2. ", function() {
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

    describe("Validator V proposes to move 500 CMTs from account #1 to #2. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        tx_result = web3.cmt.governance.proposeRecoverFund({
          from: V,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test"
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.amount).to.equal(amount)
        expect(p.Result).to.be.empty

        // default to 7 days
        let elapse = p.ExpireBlockHeight - p.BlockHeight
        expect(elapse).to.equal(Globals.Params.proposal_expire_period)

        // balance after
        balance_new = Utils.getBalance()
        let gasFee = Utils.gasFee("proposeTransferFund")
        expect(balance_new[4].minus(balance_old[4]).toNumber()).to.eq(-gasFee.toNumber())
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(Number(-amount))
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        // check deliver tx tx_result
        // let tag = tx_result.deliver_tx.tags.find(
        //   t => t.key == Globals.GasFeeKey
        // )
        // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
        //   gasFee.toString()
        // )
        // expect(tx_result.deliver_tx.gasUsed).to.eq(
        //   web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas).toString()
        // )
      })
      describe("Validator V votes for the proposal, but A, B and C vote against the proposal. The total vote (A+B+C) now exceeds 2/3.", function() {
        before(function(done) {
          if (Globals.TestMode == "cluster") {
            Utils.vote(proposalId, V, "Y")
            Utils.vote(proposalId, Globals.Accounts[0], "N")
            Utils.vote(proposalId, Globals.Accounts[1], "N")
            Utils.vote(proposalId, Globals.Accounts[2], "N")
          } else {
            Utils.vote(proposalId, V, "N")
          }
          Utils.waitBlocks(done, 1)
        })
        it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
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

    describe("Validator V proposes to move 500 CMTs from account #1 to #2. And he specifies expireTimeStamp. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        let expire = web3.cmt.getBlock("latest").timestamp + EXPIRE_BLOCKS * 10

        tx_result = web3.cmt.governance.proposeRecoverFund({
          from: V,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test",
          expireTimestamp: expire
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.amount).to.equal(amount)
        expect(p.Result).to.be.empty
        expect(p.ExpireTimestamp).to.equal(expire)

        // balance after
        balance_new = Utils.getBalance()
        let gasFee = Utils.gasFee("proposeTransferFund")
        expect(balance_new[4].minus(balance_old[4]).toNumber()).to.eq(-gasFee.toNumber())
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(Number(-amount))
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        // check deliver tx tx_result
        // let tag = tx_result.deliver_tx.tags.find(
        //   t => t.key == Globals.GasFeeKey
        // )
        // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
        //   gasFee.toString()
        // )
        // expect(tx_result.deliver_tx.gasUsed).to.eq(
        //   web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas).toString()
        // )
      })

      it("Validator V votes for the proposal, but no one else votes(wait for several blocks).", function(done) {
        if (Globals.TestMode == "cluster") {
          Utils.vote(proposalId, V, "Y")
        }
        Utils.waitBlocks(done, EXPIRE_BLOCKS)
      })

      it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Result).to.equal("Expired")
        // balance after
        balance_new = Utils.getBalance()
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(0)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
      })
    })

    describe("Validator V proposes to move 500 CMTs from account #1 to #2. And he specifies expireBlockHeight. ", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance()
      })
      it("Verify that 500 CMTs are removed from account #1 and show up as frozen amount for this account. ", function() {
        let amount = web3.toWei(500, "cmt")
        let expire = web3.cmt.blockNumber + EXPIRE_BLOCKS

        tx_result = web3.cmt.governance.proposeRecoverFund({
          from: V,
          transferFrom: Globals.Accounts[1],
          transferTo: Globals.Accounts[2],
          amount: amount,
          reason: "Governance test",
          expireBlockHeight: expire
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data

        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.amount).to.equal(amount)
        expect(p.Result).to.be.empty
        expect(p.ExpireBlockHeight).to.equal(expire)

        // balance after
        balance_new = Utils.getBalance()
        let gasFee = Utils.gasFee("proposeTransferFund")
        expect(balance_new[4].minus(balance_old[4]).toNumber()).to.eq(-gasFee.toNumber())
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(Number(-amount))
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(0)
        // check deliver tx tx_result
        // let tag = tx_result.deliver_tx.tags.find(
        //   t => t.key == Globals.GasFeeKey
        // )
        // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
        //   gasFee.toString()
        // )
        // expect(tx_result.deliver_tx.gasUsed).to.eq(
        //   web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas).toString()
        // )
      })

      it("Validator V votes for the proposal, but no one else votes(wait for several blocks).", function(done) {
        if (Globals.TestMode == "cluster") {
          Utils.vote(proposalId, V, "Y")
        }
        Utils.waitBlocks(done, EXPIRE_BLOCKS)
      })

      it("Verify that the 500 CMTs are transfered back to account #1. ", function() {
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
