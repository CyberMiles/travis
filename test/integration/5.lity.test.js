const chai = require("chai")
const expect = chai.expect

const logger = require("./logger")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Lity Test", function() {
  const EXPIRE_BLOCKS = 5

  before(function() {
    if (process.platform == "darwin") {
      // skips current and all nested describes
      logger.debug("mac os is not supported. ")
      this.test.parent.pending = true
      this.skip()
    }
  })

  describe("Reverse hello world", function() {
    let contractInstance
    it("new reverse contract", function(done) {
      contractInstance = Utils.newContract(
        web3.cmt.defaultAccount,
        Globals.Reverse.abi,
        Globals.Reverse.bytecode,
        addr => {
          Globals.Reverse.contractAddress = addr
          done()
        }
      )
    })
    it("reverse hello", function() {
      test = contractInstance.reverse.call("hello", {
        from: web3.cmt.defaultAccount
      })
      expect(test).to.equal("olleh")
    })
  })
  describe("Dogecoin", function() {
    let contractInstance
    it("new DogecoinVerifier contract", function(done) {
      contractInstance = Utils.newContract(
        web3.cmt.defaultAccount,
        Globals.Dogecoin.abi,
        Globals.Dogecoin.bytecode,
        addr => {
          Globals.Dogecoin.contractAddress = addr
          done()
        }
      )
    })
    it("verify block", function() {
      test = contractInstance.verifyBlock.call(
        1,
        "82bc68038f6034c0596b6e313729793a887fded6e92a31fbdf70863f89d9bea2",
        "3b14b76d22a3f2859d73316002bc1b9bfc7f37e2c3393be9b722b62bbd786983",
        1386474933,
        "1e0ffff0",
        3404207872
      )
      expect(test).to.equal(true)
      test = contractInstance.verifyBlock.call(
        1,
        "82bc68038f6034c0596b6e313729793a887fded6e92a31fbdf70863f89d9bea2",
        "3b14b76d22a3f2859d73316002bc1b9bfc7f37e2c3393be9b722b62bbd786983",
        1386474933,
        "1e0ffff0",
        3404207871
      )
      expect(test).to.equal(false)
    })
  })
  describe("ProposeDeployLibEni", function() {
    let proposalId = ""
    let balance_old, balance_new, tx_result
    let expireBlocks

    describe("Propose to deploy reverse 0.9.0. ", function() {
      before(function() {
        // balance before
        balance_old = web3.cmt.getBalance(web3.cmt.defaultAccount, "latest")
      })
      it("The proposal TX returns an error if bad version format. ", function() {
        expireBlocks = web3.cmt.blockNumber + EXPIRE_BLOCKS
        tx_result = web3.cmt.governance.proposeDeployLibEni({
          from: web3.cmt.defaultAccount,
          name: "reverse",
          version: "0.9.0",
          deployBlockHeight: expireBlocks,
          fileUrl: Globals.LibEni.FileUrl,
          md5: Globals.LibEni.MD5
        })
        Utils.expectTxFail(tx_result)
      })
      it("The proposal TX returns proposal id. ", function() {
        expireBlocks = web3.cmt.blockNumber + EXPIRE_BLOCKS
        tx_result = web3.cmt.governance.proposeDeployLibEni({
          from: web3.cmt.defaultAccount,
          name: "reverse",
          version: "v0.9.0",
          deployBlockHeight: expireBlocks,
          fileUrl: Globals.LibEni.FileUrl,
          md5: Globals.LibEni.MD5
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data
      })
      it("Verify gasfee. ", function() {
        // balance after
        balance_new = web3.cmt.getBalance(web3.cmt.defaultAccount, "latest")
        let gasFee = Utils.gasFee("proposeDeployLibEni")
        expect(balance_new.minus(balance_old).toNumber()).to.eq(-gasFee.toNumber())
      })
      it("Vote the proposal. ", function(done) {
        if (Globals.TestMode == "cluster") {
          Utils.vote(proposalId, Globals.Accounts[0], "Y")
          Utils.vote(proposalId, Globals.Accounts[1], "Y")
          Utils.vote(proposalId, Globals.Accounts[2], "Y")
        } else {
          Utils.vote(proposalId, web3.cmt.defaultAccount, "Y")
        }
        Utils.waitBlocks(done, 1)
      })
      it("Proposal vote passed. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Result).to.equal("Approved")
      })
      it("Returns an error if proposes another version of the same lib. ", function() {
        expireBlocks = web3.cmt.blockNumber + EXPIRE_BLOCKS
        tx_result = web3.cmt.governance.proposeDeployLibEni({
          from: web3.cmt.defaultAccount,
          name: "reverse",
          version: "v1.0.0",
          deployBlockHeight: expireBlocks,
          fileUrl: Globals.LibEni.FileUrl,
          md5: Globals.LibEni.MD5
        })
        Utils.expectTxFail(tx_result)
      })
      it("Wait for serveral blocks.", function(done) {
        Utils.waitBlocks(done, expireBlocks - web3.cmt.blockNumber + 1)
      })
      it("The library has been deployed. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.status).to.equal("deployed")
      })
    })
    describe("Propose to upgrade reverse. ", function() {
      it("The proposal TX returns an error if version <= 0.9.0. ", function() {
        expireBlocks = web3.cmt.blockNumber + EXPIRE_BLOCKS
        tx_result = web3.cmt.governance.proposeDeployLibEni({
          from: web3.cmt.defaultAccount,
          name: "reverse",
          version: "0.8.0",
          deployBlockHeight: expireBlocks,
          fileUrl: Globals.LibEni.FileUrl,
          md5: Globals.LibEni.MD5
        })
        Utils.expectTxFail(tx_result)
      })
      it("The proposal TX returns proposal id if version > 0.9.0. ", function() {
        expireBlocks = web3.cmt.blockNumber + EXPIRE_BLOCKS
        tx_result = web3.cmt.governance.proposeDeployLibEni({
          from: web3.cmt.defaultAccount,
          name: "reverse",
          version: "v1.1.0",
          deployBlockHeight: expireBlocks,
          fileUrl: Globals.LibEni.FileUrl,
          md5: Globals.LibEni.MD5
        })
        Utils.expectTxSuccess(tx_result)
        proposalId = tx_result.deliver_tx.data
      })
      it("Vote the proposal. ", function(done) {
        if (Globals.TestMode == "cluster") {
          Utils.vote(proposalId, Globals.Accounts[0], "Y")
          Utils.vote(proposalId, Globals.Accounts[1], "Y")
          Utils.vote(proposalId, Globals.Accounts[2], "Y")
        } else {
          Utils.vote(proposalId, web3.cmt.defaultAccount, "Y")
        }
        Utils.waitBlocks(done, 1)
      })
      it("Proposal vote passed. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Result).to.equal("Approved")
      })
      it("Wait for serveral blocks.", function(done) {
        Utils.waitBlocks(done, expireBlocks - web3.cmt.blockNumber + 1)
      })
      it("The library has been deployed. ", function() {
        // check proposal
        let p = Utils.getProposal(proposalId)
        expect(p.Detail.status).to.equal("deployed")
      })
    })
  })
})
