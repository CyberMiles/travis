const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")
const { Settings } = require("./constants")

describe("Stake Test", function() {
  function Amounts(maxAmount) {
    self_staking_ratio = eval(Globals.Params.self_staking_ratio)
    this.max = web3.toWei(maxAmount, "cmt")
    this.self = web3.toWei(maxAmount * self_staking_ratio, "cmt")
    this.dele1 = web3.toWei(maxAmount * 0.1, "cmt")
    this.dele2 = web3.toWei(maxAmount * (1 - self_staking_ratio - 0.3), "cmt")
    this.increasedMax = web3.toWei(maxAmount * 1.5, "cmt")
    this.reducedMax = web3.toWei(maxAmount * 0.8, "cmt")
  }

  let compRate = "4/5"
  let existingValidator = {}
  let amounts, balance_old, balance_new, tx_result
  let newAccount, newAccount2

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
      amounts = new Amounts(10000)
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
        it("D is not a validator yet", function(done) {
          tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
          // backup validator has voting power now
          expect(tx_result.data.voting_power).to.be.above(0)
          // expect(tx_result.data.tendermint_voting_power).to.eq(1)
          expect(tx_result.data.state).to.not.eq("Validator")
          Utils.waitBlocks(done, 1)
        })
        it("5 validators on tendermint", function(done) {
          Utils.getTMValidators((err, res) => {
            expect(err).to.be.null
            expect(res).to.be.not.null
            expect(res.result.validators.length).to.eq(5)
            let result = res.result.validators.filter(v => v.pub_key.value == Globals.PubKeys[3])
            expect(result.length).to.eq(1)
            expect(Number(result[0].voting_power)).to.eq(10)
            done()
          })
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

  describe("Deactivate and activate", function() {
    describe("Deactivate D", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[3]
        }
        tx_result = web3.cmt.stake.validator.deactivate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 1)
      })
      it("active=N, state=Candidate, vp=tvp=0", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.active).to.be.eq("N")
        expect(tx_result.data.state).to.be.eq("Candidate")
        expect(tx_result.data.voting_power).to.be.eq(0)
        // expect(tx_result.data.tendermint_voting_power).to.be.eq(0)
      })
      it("no award", function() {
        let awardInfos = web3.cmt.stake.validator.queryAwardInfos()
        let award = awardInfos.data.filter(a => a.address == Globals.Accounts[3])
        expect(award.length).to.be.eq(0)
      })
    })
    describe("Deactivate C", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[2]
        }
        tx_result = web3.cmt.stake.validator.deactivate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 1)
      })
      it("active=N, state=Candidate, vp=tvp=0", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[2], 0)
        expect(tx_result.data.active).to.be.eq("N")
        expect(tx_result.data.state).to.be.eq("Candidate")
        expect(tx_result.data.voting_power).to.be.eq(0)
        // expect(tx_result.data.tendermint_voting_power).to.be.eq(0)
      })
      it("no award", function() {
        let awardInfos = web3.cmt.stake.validator.queryAwardInfos()
        let award = awardInfos.data.filter(a => a.address == Globals.Accounts[2])
        expect(award.length).to.be.eq(0)
      })
    })
    describe("Activate D", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[3]
        }
        tx_result = web3.cmt.stake.validator.activate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 3)
      })
      it("active=Y, state=Validator, vp>1, tvp=10", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.active).to.be.eq("Y")
        expect(tx_result.data.state).to.be.eq("Validator")
        expect(tx_result.data.voting_power).to.be.gt(1)
        // expect(tx_result.data.tendermint_voting_power).to.eq(10)
      })
      it("got award", function() {
        if (Globals.TestMode == "cluster") {
          let awardInfos = web3.cmt.stake.validator.queryAwardInfos()
          let award = awardInfos.data.filter(a => a.address == Globals.Accounts[3])
          expect(award.length).to.be.eq(1)
          expect(Number(award[0].amount)).to.be.gt(0)
        }
      })
    })
    describe("Activate C", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[2]
        }
        tx_result = web3.cmt.stake.validator.activate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 3)
      })
      it("C active=Y, state=Validator, vp>1, tvp=10", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[2], 0)
        expect(tx_result.data.active).to.be.eq("Y")
        expect(tx_result.data.state).to.be.eq("Validator")
        expect(tx_result.data.voting_power).to.be.gt(1)
        // expect(tx_result.data.tendermint_voting_power).to.eq(10)
      })
      it("D active=Y, state=Backup Validator, vp>1, tvp=1", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.active).to.be.eq("Y")
        expect(tx_result.data.state).to.be.eq("Backup Validator")
        expect(tx_result.data.voting_power).to.be.gt(1)
        // expect(tx_result.data.tendermint_voting_power).to.be.eq(1)
      })
      it("C got award", function() {
        if (Globals.TestMode == "cluster") {
          let awardInfos = web3.cmt.stake.validator.queryAwardInfos()
          let award = awardInfos.data.filter(a => a.address == Globals.Accounts[2])
          expect(award.length).to.be.eq(1)
          expect(Number(award[0].amount)).to.be.gt(0)
        }
      })
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
        // expect(tx_result.data.tendermint_voting_power).to.eq(10)
        expect(tx_result.data.state).to.eq("Validator")
      })
      it("One of the genesis validators now drops off", function() {
        tx_result = web3.cmt.stake.validator.list()
        let vals = tx_result.data.filter(d => d.state == "Validator")
        expect(vals.length).to.eq(Globals.Params.max_vals)
      })
    })
  })

  describe("Block awards check after D becomes a validator", function() {
    let awardInfos
    before(function() {
      awardInfos = web3.cmt.stake.validator.queryAwardInfos()
    })
    it("sum of awards should equal to block award", function() {
      let sum = web3.toBigNumber(0)
      awardInfos.data.forEach(o => {
        sum = sum.plus(web3.toBigNumber(o.amount))
      })
      logger.debug(awardInfos)
      let diff = Math.abs(sum.minus(Utils.getBlockAward()).toNumber())
      logger.debug("sum, diff: ", sum.toString(), diff)
      expect(diff).to.be.at.most(Number(web3.toWei(1, "gwei")))
    })
    it("5 in total, 4 validators, 1 backup, D is validator", function() {
      if (Globals.TestMode == "cluster") {
        expect(awardInfos.data.length).to.be.eq(5)
        let vCount = awardInfos.data.filter(o => o.state == "Validator").length
        let bCount = awardInfos.data.filter(o => o.state == "Backup Validator").length
        expect(vCount).to.be.eq(4)
        expect(bCount).to.be.eq(1)
        let data = awardInfos.data.find(o => o.address == Globals.Accounts[3].toLowerCase())
        expect(data != null && data.state == "Validator").to.be.true
      } else {
        expect(awardInfos.data.length).to.be.eq(1)
      }
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
      let diff = Math.abs(dele_B.voting_power - vp_B)
      expect(diff).to.be.at.most(1)
    })
    it("check delegator C's voting power", function() {
      let n = val_D.num_of_delegators
      vp_C = Utils.calcVotingPower(n, dele_C.shares, p)
      let diff = Math.abs(dele_C.voting_power - vp_C)
      expect(diff).to.be.at.most(1)
    })
    it("check delegator D's voting power", function() {
      let n = val_D.num_of_delegators
      vp_D = Utils.calcVotingPower(n, dele_D.shares, p)
      let diff = Math.abs(dele_D.voting_power - vp_D)
      expect(diff).to.be.at.most(1)
    })
    it("check validator D's voting power", function() {
      let diff = Math.abs(val_D.voting_power - vp_B - vp_C - vp_D)
      expect(diff).to.be.at.most(1)
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
        delegation_after = Utils.getDelegation(1, 3)
        expect(
          delegation_after.pending_withdraw_amount
            .minus(delegation_before.pending_withdraw_amount)
            .toNumber()
        ).to.eq(Number(withdraw))
      })
    })
  })

  describe("Update Candidacy", function() {
    describe("Account D increase max amount", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(3)
        delegation_before = Utils.getDelegation(3, 3)
      })
      it("The verified status will still be true", function() {
        let payload = { from: Globals.Accounts[3], maxAmount: amounts.increasedMax }
        tx_result = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(tx_result)
        // check validator
        let r = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(r.data.max_shares).to.be.eq(amounts.increasedMax)
        expect(r.data.verified).to.be.eq("Y")
      })
      it("More refund will be issued", function() {
        delegation_after = Utils.getDelegation(3, 3)
        let refund = delegation_after.delegate_amount.minus(delegation_before.delegate_amount)
        expect(
          web3
            .toBigNumber(amounts.increasedMax)
            .times(0.1)
            .minus(delegation_before.shares)
            .minus(refund)
            .toNumber()
        ).to.eq(0)
        // balance after
        balance_new = Utils.getBalance(3)
        let gasFee = Utils.gasFee("updateCandidacy")
        expect(
          balance_old
            .minus(balance_new)
            .minus(gasFee)
            .minus(refund)
            .toNumber()
        ).to.eq(0)
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
    describe("Account D reduce max amount", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(3)
      })
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
        let compRate = "1/3"
        let pubKey = "LY3sRPcr63CE9uIJivApXlcYXKUoidtD+64mIljrYxk="
        let payload = {
          from: Globals.Accounts[3],
          compRate: compRate,
          pubKey: pubKey,
          description: {
            website: website
          }
        }
        tx_result = web3.cmt.stake.validator.update(payload)
        Utils.expectTxSuccess(tx_result)
        // check validator
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        // expect(tx_result.data.pub_key.value).to.be.eq(pubKey)
        expect(tx_result.data.comp_rate).to.be.eq(compRate)
        expect(tx_result.data.description.website).to.be.eq(website)
        expect(tx_result.data.verified).to.be.eq("N")
      })
    })
  })

  describe("Update D's account address", function() {
    let accountUpdateRequestId, accountUpdateRequestId2
    before(function() {
      newAccount = web3.personal.newAccount(Settings.Passphrase)
      web3.personal.unlockAccount(newAccount, Settings.Passphrase)
      newAccount2 = web3.personal.newAccount(Settings.Passphrase)
      web3.personal.unlockAccount(newAccount2, Settings.Passphrase)
      // balance before
      balance_old = Utils.getBalance(3)
    })

    it("fail if update to an existing delegator's address", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: Globals.Accounts[1] }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxFail(tx_result)
    })
    it("update validator's account address", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: newAccount }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxSuccess(tx_result)
      // balance after
      balance_new = Utils.getBalance(3)
      let gasFee = Utils.gasFee("updateAccount")
      expect(
        balance_old
          .minus(balance_new)
          .minus(gasFee)
          .toNumber()
      ).to.equal(0)
      accountUpdateRequestId = Number(
        Buffer.from(tx_result.deliver_tx.data, "base64").toString("utf-8")
      )
    })
    it("fail if update to an address that has been used in update_account", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: newAccount }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxFail(tx_result)
    })
    it("request to update B's address to newAccount2", function() {
      let payload = { from: Globals.Accounts[1], newCandidateAccount: newAccount2 }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxSuccess(tx_result)
      accountUpdateRequestId2 = Number(
        Buffer.from(tx_result.deliver_tx.data, "base64").toString("utf-8")
      )
    })
    describe("new account accept the update", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(3)
      })
      it("fail if newAccount2 delegate to B, then acceptAccountUpdate", function() {
        // fund
        let gasFee = Utils.gasFee("acceptAccountUpdate")
        let CMT1000 = web3.toWei(1000, "cmt")
        delegation_before = Utils.getDelegation(1, 1)
        Utils.transfer(
          web3.cmt.defaultAccount,
          newAccount2,
          delegation_before.shares.plus(gasFee).plus(CMT1000)
        )
        // delegate to D
        Utils.delegatorAccept(newAccount2, Globals.Accounts[1], CMT1000)
        // accept account update
        let payload = { from: newAccount2, accountUpdateRequestId: accountUpdateRequestId2 }
        tx_result = web3.cmt.stake.validator.acceptAccountUpdate(payload)
        Utils.expectTxFail(tx_result)
      })

      it("no enough balance - fail", function() {
        let payload = { from: newAccount, accountUpdateRequestId: accountUpdateRequestId }
        tx_result = web3.cmt.stake.validator.acceptAccountUpdate(payload)
        Utils.expectTxFail(tx_result)
      })
      it("send funds to the new account ", function() {
        let gasFee = Utils.gasFee("acceptAccountUpdate")
        delegation_before = Utils.getDelegation(3, 3)
        Utils.transfer(web3.cmt.defaultAccount, newAccount, delegation_before.shares.plus(gasFee))
      })
      it("it has enough balance - success", function() {
        let payload = { from: newAccount, accountUpdateRequestId: accountUpdateRequestId }
        tx_result = web3.cmt.stake.validator.acceptAccountUpdate(payload)
        Utils.expectTxSuccess(tx_result)
        // check balance of old account
        balance_new = Utils.getBalance(3)
        expect(
          balance_new
            .minus(balance_old)
            .minus(delegation_before.shares)
            .toNumber()
        ).to.be.equal(0)
        // check balance of new account
        balance = web3.cmt.getBalance(newAccount)
        logger.debug(`balance in wei: --> ${balance}`)
        expect(balance.toNumber()).to.be.eq(0)
      })
    })
  })

  describe("Candidate drops candidacy", function() {
    let theAccount
    before(function() {
      // balance before
      balance_old = Utils.getBalance()
      theAccount = newAccount ? newAccount : Globals.Accounts[3]
    })
    it("Withdraw Candidacy", function(done) {
      let payload = { from: theAccount }
      tx_result = web3.cmt.stake.validator.withdraw(payload)
      Utils.expectTxSuccess(tx_result)
      Utils.waitBlocks(done, 3)
    })
    it("Account D no longer a validator, and genesis validator restored", function() {
      // check validators, no theAccount
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(d => (d.owner_address = d.owner_address.toLowerCase()))
      expect(tx_result.data).to.not.containSubset([{ owner_address: theAccount }])
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
    it("4 validators on tendermint", function(done) {
      Utils.getTMValidators((err, res) => {
        expect(err).to.be.null
        expect(res).to.be.not.null
        expect(res.result.validators.length).to.eq(4)
        res.result.validators.forEach(v => {
          expect(Number(v.voting_power)).to.eq(10)
        })
        done()
      })
    })
  })

  describe("Block awards check after D withdraw candidacy", function() {
    let awardInfos, theAccount
    before(function() {
      awardInfos = web3.cmt.stake.validator.queryAwardInfos()
      theAccount = newAccount ? newAccount : Globals.Accounts[3]
    })
    it("sum of awards should equal to block award", function() {
      let sum = web3.toBigNumber(0)
      awardInfos.data.forEach(o => {
        sum = sum.plus(web3.toBigNumber(o.amount))
      })
      logger.debug(awardInfos)
      let diff = Math.abs(sum.minus(Utils.getBlockAward()).toNumber())
      logger.debug("sum, diff: ", sum.toString(), diff)
      expect(diff).to.be.at.most(Number(web3.toWei(1, "gwei")))
    })
    it("4 in total, 4 validators, D is not there", function() {
      if (Globals.TestMode == "cluster") {
        expect(awardInfos.data.length).to.be.eq(4)
        let vCount = awardInfos.data.filter(o => o.state == "Validator").length
        let bCount = awardInfos.data.filter(o => o.state == "Backup Validator").length
        expect(vCount).to.be.eq(4)
        expect(bCount).to.be.eq(0)
        let data = awardInfos.data.find(o => o.address == theAccount.toLowerCase())
        expect(data == null).to.be.true
      }
    })
  })
})
