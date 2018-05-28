const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const { Settings } = require("./constants")
const Utils = require("./global_hooks")

const blockAwards = 1000000000 * 0.08 / (365 * 24 * 3600 / 10)
const calcAward = powers => {
  let total = powers.reduce((s, v) => {
    return s + v
  })
  let origin = powers.map(p => p / total)
  let round1 = origin.map(p => (p > 0.1 ? 0.1 : p))

  let left =
    1 -
    round1.reduce((s, v) => {
      return s + v
    })
  let round2 = origin.map(p => left * p)

  let final = round1.map((p, idx) => {
    return strip(p + round2[idx])
  })
  // console.log(final)

  let result = powers.map((p, idx) => p + final[idx] * blockAwards)
  // console.log(result)
  return result
}

const strip = (x, precision = 12) => parseFloat(x.toPrecision(precision))

const calcAwards = (powers, blocks) => {
  for (i = 0; i < blocks; ++i) {
    powers = calcAward(powers)
  }
  return powers
}

let existingValidator = {}
let newPubKey = [
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04751",
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04752",
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04753",
  "1135A20BACD24ACAF7779FC24839350BC0D79EDBF130F0F4EE247703CEC04754"
]

let maxAmount = 200 // 2000 cmt
let deleAmount1 = maxAmount * 0.1
let deleAmount2 = maxAmount - maxAmount * 0.1 * 2
let compRate = "0.8"

describe("Stake Test", function() {
  before(function() {
    // unlock account
    web3.personal.unlockAccount(web3.cmt.defaultAccount, Settings.Passphrase)
    accounts.forEach(acc => {
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    })
  })

  before(function() {
    // get existing validator
    let result = web3.cmt.stake.queryValidators()
    let vCount = result.data.length
    expect(vCount).be.above(0)
    if (vCount == 0) process.exit(1)

    logger.debug("current validators: ", JSON.stringify(result.data))
    existingValidator = result.data[0]
    expect(existingValidator).be.an("object")

    if (vCount == 1) {
      logger.debug("one node test, add some fake validators")
      accounts.forEach((acc, idx) => {
        // declare A, B, C
        if (idx >= 3) return
        let initAmount = 1000
        let payload = {
          from: acc,
          pubKey: newPubKey[idx],
          maxAmount: web3.toWei(initAmount, "cmt"),
          compRate: compRate
        }
        let r = web3.cmt.stake.declareCandidacy(payload)
        Utils.expectTxSuccess(r)
        logger.debug("validator added, max_amount: ", initAmount)
      })
    }
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: newPubKey[3]
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: accounts[3],
        pubKey: existingValidator.pub_key.data
      }
      let r = web3.cmt.stake.declareCandidacy(payload)
      Utils.expectTxFail(r)
    })

    describe(`Declare to be a validator with ${maxAmount} CMT max and ${compRate *
      100}% compRate`, function() {
      describe(`Account D does not have ${maxAmount * 0.1} CMTs.`, function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(accounts[3]), "cmt")
            .toNumber()
          if (balance > maxAmount * 0.1) {
            web3.cmt.sendTransaction({
              from: accounts[3],
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
            from: accounts[3],
            pubKey: newPubKey[3],
            maxAmount: web3.toWei(maxAmount, "cmt"),
            compRate: compRate
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxFail(r, 20)
        })
      })

      describe(`Account D has over ${maxAmount * 0.1} CMTs.`, function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(accounts[3]), "cmt")
            .toNumber()
          if (balance < maxAmount * 0.1) {
            web3.cmt.sendTransaction({
              from: web3.cmt.defaultAccount,
              to: accounts[3],
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
            from: accounts[3],
            pubKey: newPubKey[3],
            maxAmount: wei,
            compRate: compRate
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
        candidateAddress: accounts[3],
        verified: true
      }
      let r = web3.cmt.stake.verifyCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator's status
      let result = web3.cmt.stake.queryValidators()
      expect(result.data).to.containSubset([
        { owner_address: accounts[3], verified: "Y" }
      ])
    })
  })

  describe("Query validator D. ", function() {
    it("make sure all the information are accurate.", function() {
      let result = web3.cmt.stake.queryValidator(accounts[3], 0)
      // check validator's information
      logger.debug(result.data)
      expect(result.data.owner_address).to.eq(accounts[3])
      expect(result.data.verified).to.eq("Y")
      expect(result.data.comp_rate).to.eq(compRate)
      expect(result.data.pub_key.data).to.eq(newPubKey[3])
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
          from: accounts[1],
          validatorAddress: accounts[3],
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
        let result = web3.cmt.stake.queryDelegator(accounts[1], 0)
        let delegation = result.data.filter(d => d.pub_key.data == newPubKey[3])
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(wei)
      })
      it("D is still not a validator", function() {
        let result = web3.cmt.stake.queryValidator(accounts[3], 0)
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
          from: accounts[2],
          validatorAddress: accounts[3],
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
        let result = web3.cmt.stake.queryDelegator(accounts[2], 0)
        let delegation = result.data.filter(d => d.pub_key.data == newPubKey[3])
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(wei)
      })
      it("D is now a validator", function() {
        let result = web3.cmt.stake.queryValidator(accounts[3], 0)
        let power = result.data.voting_power
        expect(power).to.be.above(0)
      })
      it("One of the genesis validators now drops off", function() {
        let result = web3.cmt.stake.queryValidators()
        let drops = result.data.filter(d => d.voting_power == 0)
        expect(drops.length).to.eq(1)
        expect(drops[0].owner_address).to.not.equal(accounts[3])
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
      powers_new = calcAwards(powers_old, blocks)
      total_awards = powers_new[4] - powers_old[4]
      // wait a few blocks
      Utils.waitBlocks(done, blocks)
    })
    it("check Validator D's current power", function() {
      let result = web3.cmt.stake.queryValidator(accounts[3], 0)
      let power = result.data.voting_power
      console.log(result.data)
      // expect(current).to.eq(powers_new[4])
    })
    it("B: total awards * 80% * 0.1", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[1], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[3])
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.1))
    })
    it("C: total awards * 80% * 0.8", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[2], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[3])
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.8))
    })
    it("D: total awards - B - C", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[3], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[3])
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
          from: accounts[1],
          validatorAddress: accounts[3],
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
        from: accounts[3],
        maxAmount: web3.toWei(newAmount, "cmt")
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(accounts[3], 0)
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
        from: accounts[3],
        description: {
          website: website
        }
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(accounts[3], 0)
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
        from: accounts[3]
      }
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validators, not include accounts[3]
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([{ owner_address: accounts[3] }])
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
