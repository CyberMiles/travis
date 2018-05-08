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
let maxAmount = "210000000000000000000" // 210 cmt
let deleAmount1 = "20000000000000000000" // 20 cmt
let deleAmount2 = "160000000000000000000" // 160 cmt
let cut = 8000

describe("Stake Test", function() {
  before(function() {
    // get existing validator
    let result = web3.cmt.stake.queryValidators()
    let vCount = result.data.length
    expect(vCount).be.above(0)
    if (vCount == 0) process.exit(1)

    existingValidator = result.data[0]
    logger.debug("current validator: ", JSON.stringify(existingValidator))
    expect(existingValidator).be.an("object")

    // unlock accounts
    accounts.forEach((acc, idx) => {
      web3.personal.unlockAccount(acc, Settings.Passphrase)
    })

    if (vCount == 1) {
      accounts.forEach((acc, idx) => {
        // declare B, C, D
        if (idx == 0) return
        let payload = {
          from: acc,
          pubKey: newPubKey[idx],
          maxAmount: web3.toWei("1000", "cmt"),
          cut: cut
        }
        let r = web3.cmt.stake.declareCandidacy(payload)
        Utils.expectTxSuccess(r)
      })
    }
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: newPubKey[0]
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

    describe("Declare to be a validator with 2000 CMT max and 80% cut", function() {
      describe("Account A does not have 200 CMTs.", function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(accounts[0]), "cmt")
            .toNumber()
          if (balance > 200) {
            web3.cmt.sendTransaction({
              from: accounts[0],
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
            from: accounts[0],
            pubKey: newPubKey[0],
            maxAmount: maxAmount,
            cut: cut
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxFail(r, 20)
        })
      })

      describe("Account A has over 200 CMTs.", function() {
        before(function(done) {
          let balance = web3
            .fromWei(web3.cmt.getBalance(accounts[0]), "cmt")
            .toNumber()
          if (balance < 200) {
            web3.cmt.sendTransaction({
              from: web3.cmt.defaultAccount,
              to: accounts[0],
              value: web3.toWei(200, "cmt")
            })
            Utils.waitBlocks(done)
          } else {
            done()
          }
        })
        before(function() {
          // balance before
          balance_old = Utils.getBalance(0)
        })

        it("Succeeds, the 200 CMTs becomes A's stake after the successful declaration", function() {
          let payload = {
            from: accounts[0],
            pubKey: newPubKey[0],
            maxAmount: maxAmount,
            cut: cut
          }
          let r = web3.cmt.stake.declareCandidacy(payload)
          Utils.expectTxSuccess(r)
          // balance after
          balance_new = Utils.getBalance(0)
          expect(balance_new[0].minus(balance_old[0]).toNumber()).to.equal(
            Number(-maxAmount * 0.1)
          )
        })
      })
    })
  })

  describe("The foundation account verifies account A. ", function() {
    it("Update the verified status to Y", function() {
      let payload = {
        from: web3.cmt.defaultAccount,
        candidateAddress: accounts[0],
        verified: true
      }
      let r = web3.cmt.stake.verifyCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator's status
      let result = web3.cmt.stake.queryValidators()
      expect(result.data).to.containSubset([
        { owner_address: accounts[0], verified: "Y" }
      ])
    })
  })

  describe("Query validator A. ", function() {
    it("make sure all the information are accurate.", function() {
      let result = web3.cmt.stake.queryValidator(accounts[0], 0)
      // check validator's information
      logger.debug(result.data)
      expect(result.data.owner_address).to.eq(accounts[0])
      expect(result.data.verified).to.eq("Y")
      expect(result.data.cut).to.eq(cut)
      expect(result.data.pub_key.data).to.eq(newPubKey[0])
      // todo
      // expect(result.data.max_shares).to.eq(maxAmount)
      // expect(result.data.shares).to.eq(maxAmount * 0.1)
    })
  })

  describe("Stake Delegate", function() {
    describe("Account B stakes 200 CMTs for A.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
      })

      it("CMTs are moved from account B", function() {
        let payload = {
          from: accounts[1],
          validatorAddress: accounts[0],
          amount: deleAmount1
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.equal(
          Number(-deleAmount1)
        )
      })
      it("CMTs show up as staked balance for B", function() {
        let result = web3.cmt.stake.queryDelegator(accounts[1], 0)
        let delegation = result.data.filter(d => d.pub_key.data == newPubKey[0])
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(deleAmount1)
      })
      it("A is still not a validator", function() {
        // todo
      })
    })
    describe("Account C stakes 1600 CMTs for A.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(2)
      })
      it("CMTs are moved from account C", function() {
        let payload = {
          from: accounts[2],
          validatorAddress: accounts[0],
          amount: deleAmount2
        }
        let r = web3.cmt.stake.delegate(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(2)
        expect(balance_new[2].minus(balance_old[2]).toNumber()).to.equal(
          Number(-deleAmount2)
        )
      })
      it("CMTs show up as staked balance for C", function() {
        let result = web3.cmt.stake.queryDelegator(accounts[2], 0)
        let delegation = result.data.filter(d => d.pub_key.data == newPubKey[0])
        expect(delegation.length).to.eq(1)
        // todo
        // expect(delegation[0].shares).to.eq(deleAmount2)
      })
      it("A is now a validator", function() {
        // todo
      })
      it("One of the genesis validators now drops off", function() {
        // todo
      })
    })
  })

  //todo
  describe("Block awards", function() {
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
    it("check Validator A's current power", function() {
      let result = web3.cmt.stake.queryValidator(accounts[0], 0)
      let power = result.data.voting_power
      console.log(result.data)
      // expect(current).to.eq(powers_new[4])
    })
    it("B: total awards * 80% * 0.1", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[1], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[0])
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.1))
    })
    it("C: total awards * 80% * 0.8", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[2], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[0])
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards * 0.8 * 0.8))
    })
    it("A: total awards - B - C", function() {
      let result = web3.cmt.stake.queryDelegator(accounts[0], 0)
      let delegation = result.data.filter(d => d.pub_key.data == newPubKey[0])
      console.log(delegation)
      expect(delegation.length).to.eq(1)
      // expect((delegation[0].awards = total_awards - total_awards * 0.8 * 0.9))
    })
  })

  describe("Stake Withdraw", function() {
    describe("Account B withdraw 200 CMTs for A.", function() {
      before(function() {
        // balance before
        balance_old = Utils.getBalance(1)
      })
      it("CMTs are moved back to account B", function() {
        let payload = {
          from: accounts[1],
          validatorAddress: accounts[0],
          amount: deleAmount1
        }
        let r = web3.cmt.stake.withdraw(payload)
        Utils.expectTxSuccess(r)
        // balance after
        balance_new = Utils.getBalance(1)
        expect(balance_new[1].minus(balance_old[1]).toNumber()).to.eq(
          Number(deleAmount1)
        )
      })
    })
  })

  describe("Update Candidacy", function() {
    it("Account A modify address to account A", function() {
      let newAddress = accounts[0]
      let payload = {
        from: accounts[0],
        maxAmount: maxAmount,
        newAddress: newAddress
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(newAddress, 0)
      expect(result.data.owner_address).to.be.eq(newAddress)
      expect(result.data.verified).to.be.eq("Y")
    })
    it("Account A modify other information", function() {
      let website = "http://aaa.com"
      let payload = {
        from: accounts[0],
        description: {
          website: website
        }
      }
      let r = web3.cmt.stake.updateCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validator
      let result = web3.cmt.stake.queryValidator(accounts[0], 0)
      expect(result.data.description.website).to.be.eq(website)
      expect(result.data.verified).to.be.eq("N")
    })
  })

  describe("Candidate drops candidacy", function() {
    before(function() {
      // balance before
      balance_old = Utils.getBalance()
    })

    it("Account A no longer a validator", function() {
      let payload = {
        from: accounts[0]
      }
      let r = web3.cmt.stake.withdrawCandidacy(payload)
      Utils.expectTxSuccess(r)
      // check validators, not include accounts[0]
      let result = web3.cmt.stake.queryValidators()
      logger.debug(result.data)
      expect(result.data).to.not.containSubset([{ owner_address: accounts[0] }])
    })
    it("All staked tokens will be distributed back to delegator addresses", function() {
      // balance after
      balance_new = Utils.getBalance()
      expect(balance_new[2].minus(balance_old[2]).toNumber()).to.be.above(
        Number(deleAmount2)
      )
    })
    it("Self-staked CMTs will be refunded back to the validator address", function() {
      expect(balance_new[0].minus(balance_old[0]).toNumber()).to.be.above(
        Number(maxAmount * 0.1)
      )
    })
  })
})
