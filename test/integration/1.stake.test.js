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
    this.self = web3
      .toBigNumber(this.max * Globals.ValMinSelfStakingRatio)
      .toString(10)
    this.dele1 = web3.toBigNumber(this.max * 0.1).toString(10)
    this.dele2 = web3
      .toBigNumber(this.max - this.self - this.dele1)
      .toString(10)
    this.reducedMax = web3.toWei(maxAmount - 1, "cmt")
  }
  let amounts = new Amounts(2000) // 2000 cmt
  let compRate = "0.8"

  let existingValidator = {}
  let balance_old, balance_new

  before(function() {
    // get existing validator
    let result = web3.cmt.stake.queryValidators()
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

    describe(`Declare to be a validator with ${web3.fromWei(
      amounts.max,
      "cmt"
    )} CMTs max and ${compRate * 100}% compRate`, function() {
      describe(`Account D does not have ${web3.fromWei(
        amounts.self,
        "cmt"
      )} CMTs.`, function() {
        before(function() {
          let balance = Utils.getBalance(3)
          Utils.transfer(Globals.Accounts[3], web3.cmt.defaultAccount, balance)
        })

        it("Fails", function() {
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: amounts.max,
            compRate: compRate
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxFail(r, 20)
        })
        after(function(done) {
          Utils.waitBlocks(done)
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
          let r = web3.cmt.stake.declareCandidacy(payload)
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
      })

      it("CMTs are moved from account B", function() {
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele1
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(
          Number(-amounts.dele1)
        )
      })
      it("CMTs show up as staked balance for B", function() {
        let result = web3.cmt.stake.queryDelegator(Globals.Accounts[1], 0)
        let delegation = result.data.filter(
          d => d.pub_key.value == Globals.PubKeys[3]
        )
        expect(delegation.length).to.eq(1)
        expect(delegation[0].delegate_amount).to.eq(amounts.dele1.toString())
      })
      it("D is still not a validator", function() {
        let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
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
      })
      it("CMTs are moved from account C", function() {
        let payload = {
          from: Globals.Accounts[2],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele2
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(2)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(
          Number(-amounts.dele2)
        )
      })
      it("CMTs show up as staked balance for C", function() {
        let result = web3.cmt.stake.queryDelegator(Globals.Accounts[2], 0)
        let delegation = result.data.filter(
          d => d.pub_key.value == Globals.PubKeys[3]
        )
        expect(delegation.length).to.eq(1)
        expect(delegation[0].delegate_amount).to.eq(amounts.dele2.toString())
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

  // todo: The withdraw request reduces the stake and voting power immediately, but has
  // a 7 day waiting period (as measured in block heights) before the fund is actually available
  // in delegator's account to trade. It puts funds into unstaked_waiting status.
  describe.skip("Stake Withdraw", function() {
    describe(`Account B withdraw ${web3.fromWei(
      amounts.dele1,
      "cmt"
    )} CMTs for D.`, function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
      })
      it("CMTs are moved back to account B", function() {
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: amounts.dele1
        }
        let r = web3.cmt.stake.withdraw(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.eq(
          Number(amounts.dele1)
        )
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
        let r = web3.cmt.stake.updateCandidacy(payload)
        Utils.expectTxSuccess(r)
        // check validator
        let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
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
        let r = web3.cmt.stake.updateCandidacy(payload)
        Utils.expectTxSuccess(r)
        // check validator
        let result = web3.cmt.stake.queryValidator(Globals.Accounts[3], 0)
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
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validators, no Globals.Accounts[3]
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([
        { owner_address: Globals.Accounts[3] }
      ])
    })
    // todo: All its staked CMTs will become unstaked_waiting, and be refunded to their source accounts after 7 days.
    it("All staked tokens will be distributed back to delegator addresses", function() {
      // balance after
      balance_new = Utils.getBalance()
      // account[1] has withdrawed, refund some interests
      expect(balance_new[1].minus(balance_old[1]).toNumber() >= 0).to.be.true
      // account[2] refund delegate amount and interests
      expect(
        balance_new[2].minus(balance_old[2]).toNumber() >= Number(amounts.dele2)
      ).to.be.true
    })
    it("Self-staked CMTs will be refunded back to the validator address", function() {
      expect(
        balance_new[3].minus(balance_old[3]).toNumber() >= Number(amounts.self)
      ).to.be.true
    })
  })
})
