const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

let existingValidators = []
let newPubKey =
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04759"
let slotId = "1ACAF2550C2B4ED0A13896DE3C4AC136"
let sequences = []

const expectTxFail = r => {
  logger.debug(r)
  expect(r)
    .to.have.property("check_tx")
    .to.have.property("code")
  expect(r)
    .to.have.property("deliver_tx")
    .to.have.property("code")
  expect(r.check_tx.code > 0 || r.deliver_tx.code > 0).to.be.true
}

const expectTxSuccess = r => {
  logger.debug(r)
  expect(r)
    .to.have.property("height")
    .and.to.gt(0)
  expect(r)
    .to.have.property("check_tx")
    .to.have.property("code")
    .and.to.eq(0)
  expect(r)
    .to.have.property("deliver_tx")
    .to.have.property("code")
    .and.to.eq(0)
}

describe.skip("Stake Test", function() {
  before(function() {
    accounts.forEach(acc => {
      // unlock account
      web3.personal.unlockAccount(acc, Settings.Passphrase)
      // get sequence
      sequences.push(web3.cmt.getSequence(acc) + 1)
    })
    // get existing validators
    let result = web3.cmt.stake.queryValidators()
    result.data.forEach(v => {
      existingValidators.push(v)
    })
    logger.debug("current validators: ", JSON.stringify(existingValidators))
    expect(existingValidators.length).to.gt(0)
  })

  describe("Declare Candidacy", function() {
    it.skip("for an existing initial validator account — fail", function() {
      let payload = {
        from: web3.cmt.accounts[0],
        pubKey: newPubKey
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      expectTxFail(r)
    })
    it("associate to an existing validator pubkey — fail", function() {
      let payload = {
        from: accounts[0],
        pubKey: existingValidators[0].pub_key.data
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      expectTxFail(r)
    })
    it("declare for one new validator pubkey and the new account A", function() {
      let payload = {
        from: accounts[0],
        pubKey: newPubKey,
        sequence: sequences[0]++
      }
      console.log(sequences)
      let r = web3.cmt.stake.declareCandidacy(payload)
      expectTxSuccess(r)
    })
  })

  describe.skip("Propose Slot", function() {
    it("Candidate node offers a slot", function() {
      let payload = {
        from: accounts[0],
        amount: 5,
        proposedRoi: 1,
        sequence: sequences[0]++
      }
      let r = web3.cmt.stake.proposeSlot(payload)
      expectTxSuccess(r)
      slotId = r.deliver_tx.data
      expect(slotId).to.be.not.empty
    })
  })

  describe.skip("Accept & Withdraw Slot", function() {
    it("Account B stakes candidate — candidate becomes validator, and account A receives block awards", function() {
      let payload = {
        from: accounts[1],
        amount: 5,
        slotId: slotId,
        sequence: sequences[1]++
      }
      let r = web3.cmt.stake.acceptSlot(payload)
      expectTxSuccess(r)
    })

    it("Account B unbind candidate — candidate is no longer a validator", function() {
      let payload = {
        from: accounts[1],
        amount: 5,
        slotId: slotId,
        sequence: sequences[1]++
      }
      let r = web3.cmt.stake.withdrawSlot(payload)
      expectTxSuccess(r)
    })

    it("Account C stakes candidate — candidate becomes validator, and account A receives block awards", function() {
      let payload = {
        from: accounts[2],
        amount: 5,
        slotId: slotId,
        sequence: sequences[2]++
      }
      let r = web3.cmt.stake.acceptSlot(payload)
      expectTxSuccess(r)
    })
  })

  describe("Edit Candidacy", function() {
    it("Account A modify address to account D", function() {
      let payload = {
        from: accounts[0],
        newAddress: accounts[3],
        sequence: sequences[0]++
      }
      let r = web3.cmt.stake.editCandidacy(payload)
      expectTxSuccess(r)
      // check validators, include newAddress and state=Y
      let result = web3.cmt.stake.queryValidators()
      expect(result.data).to.containSubset([
        { owner_address: accounts[3], state: "Y" }
      ])
    })
  })

  describe("Candidate drops candidacy", function() {
    it("it no longer a validator", function() {
      let payload = {
        from: accounts[3],
        sequence: sequences[3]++
      }
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      expectTxSuccess(r)
      // check validators, not include accounts[0] and state=Y
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([
        { owner_address: accounts[0], state: "Y" }
      ])
    })
  })
})
