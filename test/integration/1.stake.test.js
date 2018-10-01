const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")

describe("Stake Test", function() {
  function Amounts(maxAmount) {
    self_staking_ratio = eval(Globals.Params.self_staking_ratio)
    this.max = web3.toWei(maxAmount, "cmt")
    this.self = web3.toWei(maxAmount * self_staking_ratio, "cmt")
    this.dele1 = web3.toWei(maxAmount * 0.1, "cmt")
    this.dele2 = web3.toWei(maxAmount * (1 - self_staking_ratio - 0.3), "cmt")
    this.reducedMax = web3.toWei(maxAmount * 0.8, "cmt")
  }

  let compRate = "4/5"
  let existingValidator = {}
  let amounts, balance_old, balance_new, tx_result

  before(function() {
    Utils.addFakeValidators()
    amounts = new Amounts(2000000)
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
      amounts = new Amounts(20000)
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

    describe(`Declare to be a validator with 2000000 CMTs max and ${compRate} compRate`, function() {
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
            let hash = Utils.transfer(web3.cmt.defaultAccount, Globals.Accounts[3], amounts.self)
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

        it("Succeeds, the 200000 CMTs becomes D's stake after the successful declaration", function() {
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
          // let tag = tx_result.deliver_tx.tags.find(
          //   t => t.key == Globals.GasFeeKey
          // )
          // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
          //   gasFee.toString()
          // )
          // expect(tx_result.deliver_tx.gasUsed).to.eq(
          //   web3.toBigNumber(Globals.Params.declare_candidacy).toString()
          // )
        })
        it("D is not a validator yet", function() {
          tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
          // backup validator has voting power now
          expect(tx_result.data.voting_power).to.be.above(0)
          expect(tx_result.data.state).to.not.eq("Validator")
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
      tx_result.data.forEach(d => (d.owner_address = d.owner_address.toLowerCase()))
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
      expect(tx_result.data.owner_address.toLowerCase()).to.eq(Globals.Accounts[3])
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
      expect(tx_result.data.state).to.not.eq("Validator")
      delegation_after = Utils.getDelegation(3, 3)
    })
  })

  describe("Stake Delegate", function() {
    describe("Account B stakes 200000 CMTs for D.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
        // delegation before
        delegation_before = Utils.getDelegation(1, 3)
      })

      it("CMTs are moved from account B", function() {
        Utils.delegatorAccept(Globals.Accounts[1], Globals.Accounts[3], amounts.dele1)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(Number(-amounts.dele1))
      })
      it("CMTs show up as staked balance for B", function() {
        let delegation_after = Utils.getDelegation(1, 3)
        expect(
          delegation_after.delegate_amount.minus(delegation_before.delegate_amount).toNumber()
        ).to.eq(Number(amounts.dele1))
      })
    })
    describe("Account C stakes 1200000 CMTs for D.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(2)
        // delegation before
        delegation_before = Utils.getDelegation(2, 3)
      })
      it("CMTs are moved from account C", function(done) {
        Utils.delegatorAccept(Globals.Accounts[2], Globals.Accounts[3], amounts.dele2)
        // balance after
        balance_new = Utils.getBalance(2)
        expect(balance_new.minus(balance_old).toNumber()).to.equal(Number(-amounts.dele2))
        Utils.waitBlocks(done, 1)
      })
      it("CMTs show up as staked balance for C", function() {
        let delegation_after = Utils.getDelegation(2, 3)
        expect(
          delegation_after.delegate_amount.minus(delegation_before.delegate_amount).toNumber()
        ).to.eq(Number(amounts.dele2))
      })
      it("D is now a validator", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.voting_power).to.be.above(0)
        expect(tx_result.data.state).to.eq("Validator")
      })
      it("One of the genesis validators now drops off", function() {
        tx_result = web3.cmt.stake.validator.list()
        let vals = tx_result.data.filter(d => d.state == "Validator")
        expect(vals.length).to.eq(Globals.Params.max_vals)
      })
    })
  })
  describe("Voting Power", function() {
    let val_D, dele_B, dele_C, dele_D
    let p = 1
    let vp_B, vp_C, vp_D
    before(function() {
      // get validators
      let vals = web3.cmt.stake.validator.list()
      let totalShares = parseInt(
        web3.fromWei(
          vals.data.reduce((s, v) => {
            return s.plus(v.shares)
          }, web3.toBigNumber(0)),
          "cmt"
        )
      )
      // get validator D
      val_D = web3.cmt.stake.validator.query(Globals.Accounts[3], 0).data
      // calc share percentage
      let shares = parseInt(web3.fromWei(val_D.shares, "cmt"))
      let threshold = eval(Globals.Params.validator_size_threshold)
      if (shares / totalShares > threshold) {
        p = threshold / (shares / totalShares)
      }
      // get delegator B, C, D of D
      dele_B = Utils.getDelegation(1, 3)
      dele_C = Utils.getDelegation(2, 3)
      dele_D = Utils.getDelegation(3, 3)
    })
    it("check delegator B's voting power", function() {
      let n = val_D.num_of_delegators
      vp_B = Utils.calcVotingPower(n, dele_B.shares, p)
      expect(dele_B.voting_power).to.eq(vp_B)
    })
    it("check delegator C's voting power", function() {
      let n = val_D.num_of_delegators
      vp_C = Utils.calcVotingPower(n, dele_C.shares, p)
      expect(dele_C.voting_power).to.eq(vp_C)
    })
    it("check delegator D's voting power", function() {
      let n = val_D.num_of_delegators
      vp_D = Utils.calcVotingPower(n, dele_D.shares, p)
      expect(dele_D.voting_power).to.eq(vp_D)
    })
    it("check validator D's voting power", function() {
      expect(val_D.voting_power).to.eq(vp_B + vp_C + vp_D)
    })
  })

  describe("Block awards", function() {
    let blocks = 1,
      vals_expected = [],
      dele_expected = []
    let val_D, dele_B, dele_C, dele_D

    before(function() {
      if (Globals.TestMode == "single") {
        // skips current and all nested describes
        this.test.parent.pending = true
        this.skip()
      }
    })

    it("check awards", function() {
      // get validators
      let vals = web3.cmt.stake.validator.list().data
      totalVotingPower = vals.reduce((s, v) => {
        return s + v.voting_power
      }, 0)
      console.log("totalVotingPower: ", totalVotingPower)
      // get delegator B, C, D of D
      dele_B = Utils.getDelegation(1, 3)
      dele_C = Utils.getDelegation(2, 3)
      dele_D = Utils.getDelegation(3, 3)

      awardInfos = web3.cmt.stake.validator.queryAwardInfos()
      console.log(awardInfos)

      let v_ratio = eval(Globals.Params.validators_block_award_ratio)
      let blockAward = Utils.getBlockAward()
      let getAwards = dele => {
        let awards = [
          blockAward
            .times(
              (
                (Number(dele.voting_power) * (1 - dele.comp_rate) * v_ratio) /
                totalVotingPower
              ).toFixed(12)
            )
            .dividedToIntegerBy(1),
          blockAward
            .times(
              ((Number(dele.voting_power) * dele.comp_rate * v_ratio) / totalVotingPower).toFixed(
                12
              )
            )
            .dividedToIntegerBy(1)
        ]
        awards.forEach(a => console.log(a.toString()))
        return awards
      }

      award_B = getAwards(dele_B)
      award_C = getAwards(dele_C)
      award_D = getAwards(dele_D)
      award_V = award_B[0]
        .plus(award_C[0])
        .plus(award_D[0])
        .plus(award_B[1])
        .plus(award_C[1])
        .plus(award_D[1])
      console.log(award_V.toString())
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
          delegation_after.pending_withdraw_amount
            .minus(delegation_before.pending_withdraw_amount)
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
        expect(balance_new.minus(balance_old).toNumber()).to.eq(-gasFee.toNumber())
        // check deliver tx tx_result
        // let tag = tx_result.deliver_tx.tags.find(
        //   t => t.key == Globals.GasFeeKey
        // )
        // expect(Buffer.from(tag.value, "base64").toString()).to.eq(
        //   gasFee.toString()
        // )
        // expect(tx_result.deliver_tx.gasUsed).to.eq(
        //   web3.toBigNumber(Globals.Params.update_candidacy).toString()
        // )
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
    it("Withdraw Candidacy", function(done) {
      let payload = { from: Globals.Accounts[3] }
      tx_result = web3.cmt.stake.validator.withdraw(payload)
      Utils.expectTxSuccess(tx_result)
      Utils.waitBlocks(done, 1)
    })
    it("Account D no longer a validator, and genesis validator restored", function() {
      // check validators, no Globals.Accounts[3]
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(d => (d.owner_address = d.owner_address.toLowerCase()))
      expect(tx_result.data).to.not.containSubset([{ owner_address: Globals.Accounts[3] }])
      // check validators restored
      let vals = tx_result.data.filter(d => d.state == "Validator")
      expect(vals.length).to.eq(Globals.Params.max_vals)
    })
    it("account balance no change", function() {
      // balance after
      balance_new = Utils.getBalance()
      for (i = 1; i < 4; ++i) {
        expect(balance_new[i].minus(balance_old[i]).toNumber()).to.eq(0)
      }
    })
    it("All its staked CMTs will move to pending_withdraw_amount, and be refunded later", function() {
      for (i = 1; i < 4; ++i) {
        d = Utils.getDelegation(i, 3)
        expect(d).to.be.not.null
        expect(d.shares.toNumber()).to.eq(0)
      }
    })
  })
})
