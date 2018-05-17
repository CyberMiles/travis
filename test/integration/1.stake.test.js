const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Stake Test", function() {
  let existingValidator = {}

  let maxAmount = 2000 // 2000 cmt
  let deleAmount1 = maxAmount * 0.1
  let deleAmount2 = maxAmount - maxAmount * 0.1 * 2
  let cut = "0.8"

  let balance_old = new Array(4),
    balance_new = new Array(4)

  before(function() {
    // unlock account
    web3.personal.unlockAccount(web3.cmt.defaultAccount, Settings.Passphrase)
    Globals.Accounts.forEach(acc => {
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    })
  })

  before(function() {
    // get existing validator
    let result = web3.cmt.stake.queryValidators()
    expect(result.data.length).be.above(0)

    logger.debug("current validators: ", JSON.stringify(result.data))
    existingValidator = result.data[0]
    expect(existingValidator).be.an("object")

    if (Globals.TestMode == "single") {
      maxAmount = 200
      deleAmount1 = maxAmount * 0.1
      deleAmount2 = maxAmount - maxAmount * 0.1 * 2
    }
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: Globals.PubKeys[3]
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: Globals.Accounts[3],
        pubKey: existingValidator.pub_key.value
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    describe(`Declare to be a validator with ${maxAmount} CMT max and ${cut *
      100}% cut`, function() {
      describe(`Account D does not have ${maxAmount * 0.1} CMTs.`, function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(Globals.Accounts[3]), "cmt")
            .toNumber()
          if (balance > maxAmount * 0.1) {
            web3.cmt.sendTransaction({
              from: Globals.Accounts[3],
              to: web3.cmt.defaultAccount,
              value: web3.toWei(balance - 1, "cmt")
            })
            Utils.waitBlocks(done)
          } else {
            done()
          }
        })

        it("Fails", function() {
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: web3.toWei(maxAmount, "cmt"),
            cut: cut
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxFail(r, 20)
        })
      })

      describe(`Account D has over ${maxAmount * 0.1} CMTs.`, function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(Globals.Accounts[3]), "cmt")
            .toNumber()
          if (balance < maxAmount * 0.1) {
            web3.cmt.sendTransaction({
              from: web3.cmt.defaultAccount,
              to: Globals.Accounts[3],
              value: web3.toWei(maxAmount * 0.1, "cmt")
            })
            Utils.waitBlocks(done)
          } else {
            done()
          }
        })
        before(function() {
          // balance before
          balance_old = Utils.getBalance(3)
        })

        it(`Succeeds, the ${maxAmount *
          0.1} CMTs becomes D's stake after the successful declaration`, function() {
          let wei = web3.toWei(maxAmount, "cmt")
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: wei,
            cut: cut
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxSuccess(r)
          // balance after
          balance_new = Utils.getBalance(3)
          expect(balance_new[3].minus(balance_old[3]).toNumber()).to.equal(
            Number(-wei * 0.1)
          )
        })
      })
    })
  })

  describe("The foundation account verifies account D. ", function() {
    it("Update the verified status to Y", function() {
      let payload = {
        from: web3.cmt.defaultAccount,
        candidateAddress: Globals.Accounts[3],
        verified: true
      }
      let r = web3.cmt.stake.verifyCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator's status
      let result = web3.cmt.stake.queryValidators()
      expect(result.data).to.containSubset([
        { owner_address: Globals.Accounts[3], verified: "Y" }
      ])
    })
  })

  describe("Query validator D. ", function() {
    it("make sure all the information are accurate.", function() {
      let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
      // check validator's information
      logger.debug(result.data)
      expect(result.data.owner_address).to.eq(Globals.Accounts[3])
      expect(result.data.verified).to.eq("Y")
      expect(result.data.cut).to.eq(cut)
      expect(result.data.pub_key.value).to.eq(Globals.PubKeys[3])
      // todo
      // expect(result.data.max_shares).to.eq(maxAmount)
      // expect(result.data.shares).to.eq(maxAmount * 0.1)
    })
  })

  describe("Stake Delegate", function() {
    describe(`Account B stakes ${deleAmount1} CMTs for D.`, function() {
      let wei
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        wei = web3.toWei(deleAmount1, "cmt")
      })

      it("CMTs are moved from account B", function() {
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: wei
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          Number(-wei)
        )
      })
      it("CMTs show up as staked balance for B", function() {
        let result = web3.cmt.stake.queryDelegator(Globals.Accounts[1], 0)
        let delegation = result.data.filter(
          d => d.pub_key.value == Globals.PubKeys[3]
        )
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(wei)
      })
      it("D is still not a validator", function() {
        let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
        let power = result.data.voting_power
        expect(power).to.eq(0)
      })
    })
    describe(`Account C stakes ${deleAmount2} CMTs for D.`, function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(2)
        wei = web3.toWei(deleAmount2, "cmt")
      })
      it("CMTs are moved from account C", function() {
        let payload = {
          from: Globals.Accounts[2],
          validatorAddress: Globals.Accounts[3],
          amount: wei
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(2)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(
          Number(-wei)
        )
      })
      it("CMTs show up as staked balance for C", function() {
        let result = web3.cmt.stake.queryDelegator(Globals.Accounts[2], 0)
        let delegation = result.data.filter(
          d => d.pub_key.value == Globals.PubKeys[3]
        )
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(wei)
      })
      it("D is now a validator", function() {
        let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
        let power = result.data.voting_power
        expect(power).to.be.above(0)
      })
      it("One of the genesis validators now drops off", function() {
        let result = web3.cmt.stake.queryValidators()
        let drops = result.data.filter(d => d.voting_power == 0)
        expect(drops.length).to.eq(1)
        expect(drops[0].owner_address).to.not.equal(Globals.Accounts[3])
      })
    })
  })

  //todo
  describe.skip("Block awards", function() {
    let blocks = 2,
      powers_old = [],
      powers_new = [],
      total_awards = 0
    before(function(done) {
      // get current powers
      let result = web3.cmt.stake.queryValidators()
      powers_old = result.data.map(d => d.voting_power)
      // calc awards
      powers_new = Utils.calcAwards(powers_old, blocks)
      total_awards = powers_new[4] - powers_old[4]
      // wait a few blocks
      Utils.waitBlocks(done, blocks)
    })
    it("check Validator D's current power", function() {
      let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
      let power = result.data.voting_power
      console.log(result.data)
      // expect(current).to.eq(powers_new[4])
    })
    it("B: total awards * 80% * 0.1", function() {
      let result = web3.cmt.stake.queryDelegator(Globals.Accounts[1], 0)
      let delegation = result.data.filter(
        d => d.pub_key.value == Globals.PubKeys[3]
      )
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.1))
    })
    it("C: total awards * 80% * 0.8", function() {
      let result = web3.cmt.stake.queryDelegator(Globals.Accounts[2], 0)
      let delegation = result.data.filter(
        d => d.pub_key.value == Globals.PubKeys[3]
      )
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.8))
    })
    it("D: total awards - B - C", function() {
      let result = web3.cmt.stake.queryDelegator(Globals.Accounts[3], 0)
      let delegation = result.data.filter(
        d => d.pub_key.value == Globals.PubKeys[3]
      )
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards - total_awards * 0.8 * 0.9))
    })
  })

  describe.skip("Stake Withdraw", function() {
    describe(`Account B withdraw ${deleAmount1} CMTs for D.`, function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
      })
      it("CMTs are moved back to account B", function() {
        let wei = web3.toWei(deleAmount1, "cmt")
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: wei
        }
        let r = web3.cmt.stake.withdraw(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.eq(
          Number(wei)
        )
      })
    })
  })

  describe("Update Candidacy", function() {
    before(function() {
      // balance before
      balance_old = Utils.getBalance(3)
    })
    it("Account D reduce max amount", function() {
      let newAmount = maxAmount - 10
      let payload = {
        from: Globals.Accounts[3],
        maxAmount: web3.toWei(newAmount, "cmt")
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
      // todo
      // expect(result.data.max_amount).to.be.eq(newAmount)
      expect(result.data.verified).to.be.eq("Y")
      // balance after
      balance_new = Utils.getBalance(3)
      expect(balance_new[3].minus(balance_old[3]).toNumber()).to.eq(0)
    })
    it("Account D modify other information", function() {
      let website = "http://aaa.com"
      let payload = {
        from: Globals.Accounts[3],
        description: {
          website: website
        }
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
      expect(result.data.description.website).to.be.eq(website)
      expect(result.data.verified).to.be.eq("N")
    })
  })

  describe("Candidate drops candidacy", function() {
    before(function() {
      // balance before
      balance_old = Utils.getBalance()
    })

    it("Account D no longer a validator", function() {
      let payload = {
        from: Globals.Accounts[3]
      }
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validators, not include Globals.Accounts[3]
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([
        { owner_address: Globals.Accounts[3] }
      ])
    })
    it("All staked tokens will be distributed back to delegator addresses", function() {
      // balance after
      balance_new = Utils.getBalance()
      // account[1] has withdrawed, refund some interests
      expect(balance_new[1].minus(balance_old[1]).toNumber() >= 0).to.be.true
      // account[2] refund delegate amount and interests
      expect(
        balance_new[2].minus(balance_old[2]).toNumber() >=
          Number(web3.toWei(deleAmount2, "cmt"))
      ).to.be.true
    })
    it("Self-staked CMTs will be refunded back to the validator address", function() {
      expect(
        balance_new[3].minus(balance_old[3]).toNumber() >=
          Number(web3.toWei(maxAmount * 0.1, "cmt"))
      ).to.be.true
    })
  })
})
