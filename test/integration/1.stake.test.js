const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Stake Test", function() {
  function Amounts(maxAmount) {
    this.max = web3.toWei(maxAmount, "cmt")
    this.self = web3.toWei(maxAmount * Globals.ValMinSelfStakingRatio, "cmt")
    this.dele1 = web3.toWei(maxAmount * 0.1, "cmt")
    this.dele2 = web3.toWei(
      maxAmount * (1 - Globals.ValMinSelfStakingRatio - 0.1),
      "cmt"
    )
    this.reducedMax = web3.toWei(maxAmount * 0.8, "cmt")
  }
  let amounts = new Amounts(2000) // 2000 cmt
  let compRate = "0.8"

  let existingValidator = {}
  let balance_old, balance_new

  before(function() {
    // get existing validator
    let result = web3.cmt.stake.validator.list()
    expect(result.data.length).be.above(0)

    logger.debug("current validators: ", JSON.stringify(result.data))
    existingValidator = result.data[0]
    expect(existingValidator).be.an("object")

    if (Globals.TestMode == "single") {
      amounts = new Amounts(200) // 200 cmt
    }
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: Globals.PubKeys[3]
      }
      let r = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(r)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: Globals.Accounts[3],
        pubKey: existingValidator.pub_key.value
      }
      let r = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(r)
    })

    describe(`Declare to be a validator with ${web3.fromWei(
      amounts.max,
      "cmt"
    )} CMTs max and ${compRate * 100}% compRate`, function() {
      describe(`Account D does not have ${web3.fromWei(
        amounts.self,
        "cmt"
      )} CMTs.`, function() {
        before(function() {
          balance = Utils.getBalance(3)
          Utils.transfer(Globals.Accounts[3], web3.cmt.defaultAccount, balance)
        })
        after(function() {
          Utils.transfer(web3.cmt.defaultAccount, Globals.Accounts[3], balance)
        })

        it("Fails", function() {
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: amounts.max,
            compRate: compRate
          }
          let r = web3.cmt.stake.validator.declare(payload)
          Utils.expectTxFail(r, 20)
        })
      })

      describe(`Account D has over ${web3.fromWei(
        amounts.self,
        "cmt"
      )} CMTs.`, function() {
        before(function(done) {
          let balance = Utils.getBalance(3)
          if (balance.minus(amounts.self) < 0) {
            let hash = Utils.transfer(
              web3.cmt.defaultAccount,
              Globals.Accounts[3],
              amounts.self
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
        before(function() {
          // balance before
          balance_old = Utils.getBalance(3)
        })

        it(`Succeeds, the ${web3.fromWei(
          amounts.self,
          "cmt"
        )} CMTs becomes D's stake after the successful declaration`, function() {
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: amounts.max,
            compRate: compRate
          }
          let r = web3.cmt.stake.validator.declare(payload)
          Utils.expectTxSuccess(r)
          // balance after
          balance_new = Utils.getBalance(3)
          expect(balance_new.minus(balance_old).toNumber()).to.equal(
            Number(-amounts.self)
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
      let r = web3.cmt.stake.validator.verify(payload)
      Utils.expectTxSuccess(r)
      // check validator's status
      let result = web3.cmt.stake.validator.list()
      expect(result.data).to.containSubset([
        { owner_address: Globals.Accounts[3], verified: "Y" }
      ])
    })
  })

  describe("Query validator D. ", function() {
    it("make sure all the information are accurate.", function() {
      let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
      // check validator's information
      logger.debug(result.data)
      expect(result.data.owner_address).to.eq(Globals.Accounts[3])
      expect(result.data.verified).to.eq("Y")
      expect(result.data.comp_rate).to.eq(compRate)
      expect(result.data.pub_key.value).to.eq(Globals.PubKeys[3])
      expect(result.data.max_shares).to.eq(amounts.max.toString())
      expect(result.data.shares).to.eq(amounts.self.toString())
    })
  })

  describe("Stake Delegate", function() {
    describe(`Account B stakes ${web3.fromWei(
      amounts.dele1,
      "cmt"
    )} CMTs for D.`, function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        // delegation before
        delegation_before = Utils.getDelegation(1, 3)
      })

      it("CMTs are moved from account B", function() {
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele1
        }
        let r = web3.cmt.stake.delegator.accept(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(
          Number(-amounts.dele1)
        )
      })
      it("CMTs show up as staked balance for B", function() {
        let delegation_after = Utils.getDelegation(1, 3)
        expect(
          delegation_after.delegate_amount
            .minus(delegation_before.delegate_amount)
            .toNumber()
        ).to.eq(Number(amounts.dele1))
      })
      it("D is still not a validator", function() {
        let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        let power = result.data.voting_power
        expect(power).to.eq(0)
      })
    })
    describe(`Account C stakes ${web3.fromWei(
      amounts.dele2,
      "cmt"
    )} CMTs for D.`, function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(2)
        // delegation before
        delegation_before = Utils.getDelegation(2, 3)
      })
      it("CMTs are moved from account C", function() {
        let payload = {
          from: Globals.Accounts[2],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele2
        }
        let r = web3.cmt.stake.delegator.accept(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(2)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(
          Number(-amounts.dele2)
        )
      })
      it("CMTs show up as staked balance for C", function() {
        let delegation_after = Utils.getDelegation(2, 3)
        expect(
          delegation_after.delegate_amount
            .minus(delegation_before.delegate_amount)
            .toNumber()
        ).to.eq(Number(amounts.dele2))
      })
      it("D is now a validator", function() {
        let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        let power = result.data.voting_power
        expect(power).to.be.above(0)
      })
      it("One of the genesis validators now drops off", function() {
        let result = web3.cmt.stake.validator.list()
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
      let result = web3.cmt.stake.validator.list()
      powers_old = result.data.map(d => d.voting_power)
      // calc awards
      powers_new = Utils.calcAwards(powers_old, blocks)
      total_awards = powers_new[4] - powers_old[4]
      // wait a few blocks
      Utils.waitBlocks(done, blocks)
    })
    it("check Validator D's current power", function() {
      let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
      let power = result.data.voting_power
      console.log(result.data)
      // expect(current).to.eq(powers_new[4])
    })
    it("B: total awards * 80% * 0.1", function() {
      let delegation = Utils.getDelegation(1, 3)
      console.log(delegation)
      // expect((delegation.awards = total_awards * 0.8 * 0.1))
    })
    it("C: total awards * 80% * 0.8", function() {
      let delegation = Utils.getDelegation(2, 3)
      console.log(delegation)
      // expect((delegation.awards = total_awards * 0.8 * 0.8))
    })
    it("D: total awards - B - C", function() {
      let delegation = Utils.getDelegation(3, 3)
      console.log(delegation)
      // expect((delegation.awards = total_awards - total_awards * 0.8 * 0.9))
    })
  })

  describe("Stake Withdraw", function() {
    describe(`Account B withdraw ${web3.fromWei(
      amounts.dele1,
      "cmt"
    )} CMTs for D.`, function() {
      let delegation_before, delegation_after
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        // delegation before
        delegation_before = Utils.getDelegation(1, 3)
      })
      it("CMTs are moved back to account B(locked)", function() {
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele1
        }
        let r = web3.cmt.stake.delegator.withdraw(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.eq(Number(0))
        // delegation after
        let delegation_after = Utils.getDelegation(1, 3)
        expect(
          delegation_after.withdraw_amount
            .minus(delegation_before.withdraw_amount)
            .toNumber()
        ).to.eq(Number(amounts.dele1))
      })
    })
  })

  describe("Update Candidacy", function() {
    before(function() {
      // balance before
      balance_old = Utils.getBalance(3)
    })
    describe("Account D reduce max amount", function() {
      it("The verified status will still be true", function() {
        let payload = {
          from: Globals.Accounts[3],
          maxAmount: amounts.reducedMax
        }
        let r = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(r)
        // check validator
        let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(result.data.max_shares).to.be.eq(amounts.reducedMax)
        expect(result.data.verified).to.be.eq("Y")
      })
      it("No refund will be issued", function() {
        // balance after
        balance_new = Utils.getBalance(3)
        expect(balance_new.minus(balance_old).toNumber()).to.eq(0)
      })
    })
    describe("Account D modify other information", function() {
      it("The verified status will set to false", function() {
        let website = "http://aaa.com"
        let payload = {
          from: Globals.Accounts[3],
          description: {
            website: website
          }
        }
        let r = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(r)
        // check validator
        let result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(result.data.description.website).to.be.eq(website)
        expect(result.data.verified).to.be.eq("N")
      })
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
      let r = web3.cmt.stake.validator.withdraw(payload)
      Utils.expectTxSuccess(r)
      // check validators, no Globals.Accounts[3]
      let result = web3.cmt.stake.validator.list()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([
        { owner_address: Globals.Accounts[3] }
      ])
    })
    it("account balance no change", function() {
      // balance after
      balance_new = Utils.getBalance()
      for (i = 1; i < 4; ++i) {
        expect(balance_new[i].minus(balance_old[i]).toNumber()).to.eq(0)
      }
    })
    it("All its staked CMTs will move to withdraw_amount, and be refunded later", function() {
      for (i = 1; i < 4; ++i) {
        d = Utils.getDelegation(i, 3)
        expect(d).to.be.not.null
        expect(
          d.delegate_amount
            .plus(d.award_amount)
            .minus(d.slash_amount)
            .minus(d.withdraw_amount)
            .toNumber()
        ).to.eq(0)
      }
    })
  })
})
