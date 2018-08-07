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
    self_staking_ratio = Globals.Params.self_staking_ratio
    this.max = web3.toWei(maxAmount, "cmt")
    this.self = web3.toWei(maxAmount * self_staking_ratio, "cmt")
    this.dele1 = web3.toWei(maxAmount * 0.1, "cmt")
    this.dele2 = web3.toWei(maxAmount * (1 - self_staking_ratio - 0.3), "cmt")
    this.reducedMax = web3.toWei(maxAmount * 0.8, "cmt")
  }

  let compRate = "0.8"
  let existingValidator = {}
  let amounts, balance_old, balance_new, tx_result

  before(function() {
    Utils.addFakeValidators()
    amounts = new Amounts(20000) // 20000 cmt
  })

  after(function() {
    Utils.removeFakeValidators()
  })

  before(function() {
    // get existing validator
    tx_result = web3.cmt.stake.validator.list()
    expect(tx_result.data.length).be.above(0)

    logger.debug("current validators: ", JSON.stringify(tx_result.data))
    existingValidator = tx_result.data[0]
    expect(existingValidator).be.an("object")

    if (Globals.TestMode == "single") {
      amounts = new Amounts(2000) // 2000 cmt
    }
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: Globals.PubKeys[3]
      }
      tx_result = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(tx_result)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: Globals.Accounts[3],
        pubKey: existingValidator.pub_key.value
      }
      tx_result = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(tx_result)
    })

    describe(`Declare to be a validator with 20000 CMTs max and ${compRate *
      100}% compRate`, function() {
      describe("Account D does not have enough CMTs.", function() {
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
          tx_result = web3.cmt.stake.validator.declare(payload)
          Utils.expectTxFail(tx_result, 20)
        })
      })

      describe("Account D has enough CMTs.", function() {
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

        it("Succeeds, the 2000 CMTs becomes D's stake after the successful declaration", function() {
          let payload = {
            from: Globals.Accounts[3],
            pubKey: Globals.PubKeys[3],
            maxAmount: amounts.max,
            compRate: compRate
          }
          tx_result = web3.cmt.stake.validator.declare(payload)
          Utils.expectTxSuccess(tx_result)
          // balance after
          balance_new = Utils.getBalance(3)
          let gasFee = Utils.gasFee("declareCandidacy")
          expect(balance_new.minus(balance_old).toNumber()).to.equal(
            -gasFee.plus(amounts.self).toNumber()
          )
          // check deliver tx tx_result
          expect(tx_result.deliver_tx.fee.value).to.eq(gasFee.toString())
          expect(tx_result.deliver_tx.gasUsed).to.eq(
            web3.toBigNumber(Globals.Params.declare_candidacy).toString()
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
      tx_result = web3.cmt.stake.validator.verify(payload)
      Utils.expectTxSuccess(tx_result)
      // check validator's status
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(
        d => (d.owner_address = d.owner_address.toLowerCase())
      )
      expect(tx_result.data).to.containSubset([
        { owner_address: Globals.Accounts[3], verified: "Y" }
      ])
    })
  })

  describe("Query validator D. ", function() {
    it("make sure all the information are accurate.", function() {
      tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
      // check validator's information
      logger.debug(tx_result.data)
      expect(tx_result.data.owner_address.toLowerCase()).to.eq(
        Globals.Accounts[3]
      )
      expect(tx_result.data.verified).to.eq("Y")
      expect(tx_result.data.comp_rate).to.eq(compRate)
      expect(tx_result.data.pub_key.value).to.eq(Globals.PubKeys[3])
      expect(tx_result.data.max_shares).to.eq(amounts.max.toString())
      expect(
        web3
          .toBigNumber(tx_result.data.shares)
          .minus(amounts.self)
          .toNumber()
      ).to.gte(0)
      expect(tx_result.data.state).to.eq("Backup Validator")
      delegation_after = Utils.getDelegation(3, 3)
    })
  })

  describe("Stake Delegate", function() {
    describe("Account B stakes 2000 CMTs for D.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        // delegation before
        delegation_before = Utils.getDelegation(1, 3)
      })

      it("CMTs are moved from account B", function() {
        Utils.delegatorAccept(
          Globals.Accounts[1],
          Globals.Accounts[3],
          amounts.dele1
        )
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
      it("D is still a backup", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.voting_power).to.eq(0)
        expect(tx_result.data.state).to.eq("Backup Validator")
      })
    })
    describe("Account C stakes 12000 CMTs for D.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(2)
        // delegation before
        delegation_before = Utils.getDelegation(2, 3)
      })
      it("CMTs are moved from account C", function() {
        Utils.delegatorAccept(
          Globals.Accounts[2],
          Globals.Accounts[3],
          amounts.dele2
        )
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
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.voting_power).to.be.above(0)
        expect(tx_result.data.state).to.eq("Validator")
      })
      it("One of the genesis validators now drops off", function() {
        tx_result = web3.cmt.stake.validator.list()
        let drops = tx_result.data.filter(d => d.voting_power == 0)
        expect(drops.length).to.eq(1)
        expect(drops[0].owner_address.toLowerCase()).to.not.equal(
          Globals.Accounts[3]
        )
        expect(drops[0].state).to.eq("Backup Validator")
      })
    })
  })

  describe.skip("Block awards", function() {
    let blocks = 1,
      vals_expected = [],
      dele_expected = []

    before(function(done) {
      if (Globals.TestMode == "single") {
        this.skip()
      }
      Utils.waitBlocks(done, 1)
    })

    before(function(done) {
      if (Globals.TestMode == "single") {
        this.skip()
      }
      // validators
      const getShares = (vals, acc) => {
        let v = vals.find(v => v.owner_address == acc)
        if (v) {
          return web3.toBigNumber(v.shares)
        }
        return web3.toBigNumber(0)
      }
      tx_result = web3.cmt.stake.validator.list()
      console.log(tx_result)
      let d0 = getShares(tx_result.data, Globals.Accounts[3])
      vals_expected = Utils.calcAwards(tx_result.data, blocks)
      let d1 = getShares(vals_expected, Globals.Accounts[3])
      // delegators of validator D
      let award = web3.toBigNumber(d1).minus(d0)
      dele_expected = Utils.calcDeleAwards(
        award,
        compRate,
        [
          Utils.getDelegation(1, 3).shares,
          Utils.getDelegation(2, 3).shares,
          Utils.getDelegation(3, 3).shares
        ],
        blocks
      )
      // wait a few blocks
      Utils.waitBlocks(done, blocks)
    })

    it("check Validators' shares", function() {
      // validators
      let vals = {}
      tx_result = web3.cmt.stake.validator.list()
      console.log(tx_result)
      tx_result.data.forEach(r => (vals[r.owner_address] = r.shares))
      logger.debug(vals)

      vals_expected.forEach(v => {
        let diff = web3
          .fromWei(
            web3
              .toBigNumber(v.shares)
              .minus(vals[v.owner_address])
              .abs(),
            "gwei"
          )
          .toNumber()
        expect(diff).lt(0.01)
      })
    })
    it("check Delegators' shares of Validator D", function() {
      // delegators of validator D
      dele_expected.forEach((d, idx) => {
        let dele = Utils.getDelegation(idx + 1, 3).shares
        let diff = web3
          .fromWei(
            web3
              .toBigNumber(d)
              .minus(dele)
              .abs(),
            "gwei"
          )
          .toNumber()
        expect(diff).lt(0.01)
      })
    })
  })

  describe("Stake Withdraw", function() {
    describe("Account B withdraw some CMTs for D.", function() {
      let delegation_before, delegation_after
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        // delegation before
        delegation_before = Utils.getDelegation(1, 3)
      })
      it("CMTs are moved back to account B(locked)", function() {
        let withdraw = web3.toBigNumber(amounts.dele1).times(0.2)
        let payload = {
          from: Globals.Accounts[1],
          validatorAddress: Globals.Accounts[3],
          amount: withdraw
        }
        tx_result = web3.cmt.stake.delegator.withdraw(payload)
        Utils.expectTxSuccess(tx_result)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.eq(Number(0))
        // delegation after
        let delegation_after = Utils.getDelegation(1, 3)
        expect(
          delegation_after.withdraw_amount
            .minus(delegation_before.withdraw_amount)
            .toNumber()
        ).to.eq(Number(withdraw))
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
        tx_result = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(tx_result)
        // check validator
        let r = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(r.data.max_shares).to.be.eq(amounts.reducedMax)
        expect(r.data.verified).to.be.eq("Y")
      })
      it("No refund will be issued", function() {
        // balance after
        balance_new = Utils.getBalance(3)
        let gasFee = Utils.gasFee("updateCandidacy")
        expect(balance_new.minus(balance_old).toNumber()).to.eq(
          -gasFee.toNumber()
        )
        // check deliver tx tx_result
        expect(tx_result.deliver_tx.fee.value).to.eq(gasFee.toString())
        expect(tx_result.deliver_tx.gasUsed).to.eq(
          web3.toBigNumber(Globals.Params.update_candidacy).toString()
        )
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
        tx_result = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(tx_result)
        // check validator
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.description.website).to.be.eq(website)
        expect(tx_result.data.verified).to.be.eq("N")
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
      tx_result = web3.cmt.stake.validator.withdraw(payload)
      Utils.expectTxSuccess(tx_result)
      // check validators, no Globals.Accounts[3]
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(
        d => (d.owner_address = d.owner_address.toLowerCase())
      )
      expect(tx_result.data).to.not.containSubset([
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
