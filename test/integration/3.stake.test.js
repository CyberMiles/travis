const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

let existingValidator = {}
let newPubKey =
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04759"
let slotId = "1ACAF2550C2B4ED0A13896DE3C4AC136"

describe.skip("Stake Test", function() {
  before(function() {
    accounts.forEach(acc => {
      // unlock account
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    })
    // get existing validator
    let result = web3.cmt.stake.queryValidators()
    if (result.data.length > 0) {
      existingValidator = result.data[0]
    }
    logger.debug("current validator: ", JSON.stringify(existingValidator))
    // expect(existingValidators.length).to.gt(0)
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: newPubKey
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: accounts[0],
        pubKey: existingValidator.pub_key.data
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    it("declare for one new validator pubkey and the new account A", function() {
      // get sequence
      console.log(web3.cmt.getSequence(accounts[0]))
      let payload = {
        from: accounts[0],
        pubKey: newPubKey
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxSuccess(r)
    })
  })

  describe("Edit Candidacy", function() {
    it("Account A modify address to account D", function() {
      let payload = {
        from: accounts[0],
        newAddress: accounts[3]
      }
      let r = web3.cmt.stake.editCandidacy(payload)
      Utils.expectTxSuccess(r)
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
        from: accounts[3]
      }
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validators, not include accounts[0] and state=Y
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([
        { owner_address: accounts[0], state: "Y" }
      ])
    })
  })
})
