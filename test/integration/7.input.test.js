const Tx = require("ethereumjs-tx")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")
const logger = require("./logger")

let A, B, C, D
let proposalId
let accounts = [
  {
    addr: "0x7EFF122B94897EA5B0E2A9ABF47B86337FAFEBDC",
    pkey: "0ce9f0b80483fbae111ac7df48527d443594a902b00fc797856e35eb7b12b4be"
  },
  {
    addr: "0x283ed77f880d87dbde8721259f80517a38ae5b4f",
    pkey: "f329b593cf31ee9b5cbd940f966473961acd8c09a3708565747e7004b6117834"
  },
  {
    addr: "0xb736deeba456120a706f19b92f77aeac56fd8fb1",
    pkey: "2ab880ede97305a45dcdf7cf0ac8fe31133f7e30a9fb9e55e9fd9f85a6258521"
  },
  {
    addr: "0x38d7b32e7b5056b297baf1a1e950abbaa19ce949",
    pkey: "2588cb6702d296c6f2db8eef7860426058bc646de8d3debc313ef9716e973d97"
  },
  {
    addr: "0xf616dd410161faa94737a88943cf582f95f0b630",
    pkey: "1921567765af66367389d1dbbcfa173791623255bde6841040f1b36ea2da5fab"
  },
  {
    addr: "0xc156a70e4e3421c06f8b4780e03c97935a2564ed",
    pkey: "b7271283e65b2e4a457b4641c01cd1c8ea7046a5ba3319b744b77ccf931907d6"
  },
  {
    addr: "0x51a063839e9f09fd41709782dbcb90e9192f01a0",
    pkey: "6754e0c56cddc05a29e30ad6af7ee852bda627cdb1f98e9ec710ee98c97ee8ee"
  },
  {
    addr: "0xfed52a089b540f97b48451eebdfff4b6f9f3c2d6",
    pkey: "713818b29dc78f3bbd4b54d2d4bfa2fb722c930f8b7e9c2679df03a50811169a"
  }
]

describe("API Input Parameter Test", function() {
  before(function() {
    A = accounts[0]
    B = accounts[1]
    C = accounts[2]
    D = accounts[7]
  })
  after(function(done) {
    web3.cmt.stake.validator.query(D.addr, 0, (err, res) => {
      if (!err) {
        sendTx(D, "withdrawCandidacy", [], Utils.expectTxSuccess, done)
      } else {
        done()
      }
    })
  })
  describe("junk requests", function() {
    it("fail if junk", function(done) {
      sendTx(A, "junk", [], Utils.expectTxFail, done)
    })
  })

  describe("stake/declareCandidacy", function() {
    it("fail if empty input", function(done) {
      sendTx(D, "declare", [], Utils.expectTxFail, done)
    })
    it("fail if bad pub key", function(done) {
      sendTx(D, "declare", ["abc", "11", "0.15"], Utils.expectTxFail, done)
    })
    it("fail if no max_amount specified", function(done) {
      sendTx(D, "declare", [Globals.PubKeys[3]], Utils.expectTxFail, done)
    })
    it("fail if bad max_amount format", function(done) {
      sendTx(D, "declare", [Globals.PubKeys[3], "A"], Utils.expectTxFail, done)
    })
    it("fail if max_amount<=0", function(done) {
      sendTx(D, "declare", [Globals.PubKeys[3], "-1"], Utils.expectTxFail, done)
    })
    it("fail if no comp_rate specified", function(done) {
      sendTx(D, "declare", [Globals.PubKeys[3], "1"], Utils.expectTxFail, done)
    })
    it("fail if bad comp_rate format", function(done) {
      sendTx(
        D,
        "declare",
        [Globals.PubKeys[3], "11", "a"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if wrong comp_rate scope", function(done) {
      sendTx(
        D,
        "declare",
        [Globals.PubKeys[3], "11", "-1"],
        Utils.expectTxFail,
        done
      )
    })
  })
  describe("stake/updateCandidacy", function() {
    it("success if empty input(nothing changed)", function(done) {
      sendTx(A, "update", [], Utils.expectTxSuccess, done)
    })
    it("fail if max_amount<=0", function(done) {
      sendTx(A, "update", ["-1"], Utils.expectTxFail, done)
    })
    it("fail if bad max_amount format", function(done) {
      sendTx(A, "update", ["A"], Utils.expectTxFail, done)
    })
  })
  describe("stake/set-comprate", function() {
    it("fail if empty input", function(done) {
      sendTx(A, "compRate", [], Utils.expectTxFail, done)
    })
    it("fail if bad validator", function(done) {
      sendTx(D, "compRate", [A.addr, "0.1"], Utils.expectTxFail, done)
    })
    it("fail if bad delegator", function(done) {
      sendTx(A, "compRate", [C.addr, "0.1"], Utils.expectTxFail, done)
    })
    it("fail if no comp_rate specified", function(done) {
      sendTx(A, "compRate", [A.addr], Utils.expectTxFail, done)
    })
    it("fail if bad comp_rate format", function(done) {
      sendTx(A, "compRate", [A.addr, "A"], Utils.expectTxFail, done)
    })
    it("fail if wrong comp_rate scope", function(done) {
      sendTx(A, "compRate", [A.addr, "-1"], Utils.expectTxFail, done)
    })
  })
  describe("stake/verify", function() {
    it("fail if empty input", function(done) {
      sendTx(A, "verify", [], Utils.expectTxFail, done)
    })
    it("fail if not foundation", function(done) {
      sendTx(D, "verify", [A.addr], Utils.expectTxFail, done)
    })
    it("fail if bad validator", function(done) {
      sendTx(D, "verify", [D.addr], Utils.expectTxFail, done)
    })
    it("fail if bad verified format", function(done) {
      sendTx(A, "verify", [A.addr, "A"], Utils.expectTxFail, done)
    })
    it("success if no verifed specified(default to false)", function(done) {
      sendTx(A, "verify", [A.addr], Utils.expectTxSuccess, done)
    })
  })
  describe("stake/accept", function() {
    it("fail if empty input", function(done) {
      sendTx(D, "accept", [], Utils.expectTxFail, done)
    })
    it("fail if bad cube batch", function(done) {
      sendTx(D, "accept", [A.addr, "10", "AA"], Utils.expectTxFail, done)
    })
    it("fail if bad amount format", function(done) {
      sendTx(D, "accept", [A.addr, "A", "01"], Utils.expectTxFail, done)
    })
    it("fail if amount<=0", function(done) {
      sendTx(D, "accept", [A.addr, "-1", "01"], Utils.expectTxFail, done)
    })
  })
  describe("stake/withdraw", function() {
    before(function(done) {
      let balance = web3.cmt.getBalance(D.addr)
      if (balance < 1) Utils.transfer(A.addr, D.addr, 1)
      sendTx(D, "accept", [A.addr, "1", "01"], Utils.expectTxSuccess, done)
    })
    it("fail if empty input", function(done) {
      sendTx(D, "withdraw", [], Utils.expectTxFail, done)
    })
    it("fail if bad validator", function(done) {
      sendTx(D, "withdraw", [D.addr], Utils.expectTxFail, done)
    })
    it("fail if bad amount format", function(done) {
      sendTx(D, "withdraw", [A.addr, "A"], Utils.expectTxFail, done)
    })
    it("fail if amount<=0", function(done) {
      sendTx(D, "withdraw", [A.addr, "-1"], Utils.expectTxFail, done)
    })
    it("fail if bad delegator", function(done) {
      sendTx(B, "withdraw", [A.addr, "1"], Utils.expectTxFail, done)
    })
    it("success if all set", function(done) {
      sendTx(D, "withdraw", [A.addr, "1"], Utils.expectTxSuccess, done)
    })
  })
  describe("gov/transFund", function() {
    it("fail if empty input", function(done) {
      sendTx(A, "transFund", [], Utils.expectTxFail, done)
    })
    it("fail if no proposer", function(done) {
      sendTx(A, "transFund", [null, A.addr, B.addr], Utils.expectTxFail, done)
    })
    it("fail if no from/to", function(done) {
      sendTx(A, "transFund", [A.addr], Utils.expectTxFail, done)
    })
    it("fail if bad amount format", function(done) {
      sendTx(
        A,
        "transFund",
        [A.addr, A.addr, B.addr, "A"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if amount<=0", function(done) {
      sendTx(
        A,
        "transFund",
        [A.addr, A.addr, B.addr, "-1"],
        Utils.expectTxFail,
        done
      )
    })
  })
  describe("gov/changeParam", function() {
    it("fail if empty input", function(done) {
      sendTx(A, "changeParam", [], Utils.expectTxFail, done)
    })
    it("fail if no name&value", function(done) {
      sendTx(A, "changeParam", [A.addr], Utils.expectTxFail, done)
    })
    it("fail if bad format", function(done) {
      sendTx(
        A,
        "changeParam",
        [A.addr, "max_vals", "A"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if bad expire timestamp", function(done) {
      sendTx(
        A,
        "changeParam",
        [A.addr, "max_vals", "4", null, -1],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if bad expire block", function(done) {
      sendTx(
        A,
        "changeParam",
        [A.addr, "max_vals", "4", null, null, -1],
        Utils.expectTxFail,
        done
      )
    })
    it("success if all set", function(done) {
      sendTx(
        A,
        "changeParam",
        [A.addr, "max_vals", "4"],
        Utils.expectTxSuccess,
        done
      )
    })
  })
  describe("gov/deployLibEni", function() {
    before(function() {
      if (process.platform == "darwin") {
        logger.debug("mac os is not supported. ")
        this.skip()
      }
    })
    it("fail if empty input", function(done) {
      sendTx(A, "deployLibEni", [], Utils.expectTxFail, done)
    })
    it("fail if no other parameters", function(done) {
      sendTx(A, "deployLibEni", [A.addr], Utils.expectTxFail, done)
    })
    it("fail if bad version format", function(done) {
      sendTx(
        A,
        "deployLibEni",
        [A.addr, "test", "aa"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if no fileurl&md5", function(done) {
      sendTx(
        A,
        "deployLibEni",
        [A.addr, "test", "v1.0.0"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if empty fileurl", function(done) {
      sendTx(
        A,
        "deployLibEni",
        [A.addr, "test", "v1.0.0", "{}"],
        Utils.expectTxFail,
        done
      )
    })
    it("fail if empty md5", function(done) {
      sendTx(
        A,
        "deployLibEni",
        [A.addr, "test", "v1.0.0", Globals.LibEni.FileUrl, "{}"],
        Utils.expectTxFail,
        done
      )
    })
  })
  describe("gov/vote", function() {
    before(function() {
      let r = web3.cmt.governance.listProposals()
      if (r.data && r.data.length > 0) {
        proposalId = r.data[r.data.length - 1].Id
      }
    })
    it("fail if empty input", function(done) {
      sendTx(A, "vote", [], Utils.expectTxFail, done)
    })
    it("fail if bad proposal id", function(done) {
      sendTx(A, "vote", ["A", A.addr], Utils.expectTxFail, done)
    })
    it("success if no answer set(default to empty string)", function(done) {
      if (proposalId) {
        sendTx(A, "vote", [proposalId, A.addr], Utils.expectTxSuccess, done)
      } else done()
    })
  })
})

function sendTx(account, op, data, fnExpect, done) {
  const privateKey = new Buffer(account.pkey, "hex")
  const nonce = web3.cmt.getTransactionCount(account.addr)
  let txInner
  switch (op) {
    case "declare":
      txInner = {
        type: "stake/declareCandidacy",
        data: { pub_key: data[0], max_amount: data[1], comp_rate: data[2] }
      }
      break
    case "update":
      txInner = { type: "stake/updateCandidacy", data: { max_amount: data[0] } }
      break
    case "compRate":
      txInner = {
        type: "stake/set-comprate",
        data: { delegator_address: data[0], comp_rate: data[1] }
      }
      break
    case "verify":
      txInner = {
        type: "stake/verifyCandidacy",
        data: { candidate_address: data[0], verified: data[1] }
      }
      break
    case "withdrawCandidacy":
      txInner = { type: "stake/withdrawCandidacy", data: {} }
      break
    case "activate":
      txInner = { type: "stake/activateCandidacy", data: {} }
      break
    case "accept":
      txInner = {
        type: "stake/delegate",
        data: {
          validator_address: data[0],
          amount: data[1],
          cube_batch: data[2],
          sig: Utils.cubeSign(account.addr, nonce)
        }
      }
      break
    case "withdraw":
      txInner = {
        type: "stake/withdraw",
        data: { validator_address: data[0], amount: data[1] }
      }
      break
    case "transFund":
      txInner = {
        type: "governance/propose/transfer_fund",
        data: {
          proposer: data[0],
          from: data[1],
          to: data[2],
          amount: data[3],
          reason: data[4],
          expire_timestamp: data[5],
          expire_block_height: data[6]
        }
      }
      break
    case "changeParam":
      txInner = {
        type: "governance/propose/change_param",
        data: {
          proposer: data[0],
          name: data[1],
          value: data[2],
          reason: data[3],
          expire_timestamp: data[4],
          expire_block_height: data[5]
        }
      }
      break
    case "deployLibEni":
      txInner = {
        type: "governance/propose/deploy_libeni",
        data: {
          proposer: data[0],
          name: data[1],
          version: data[2],
          fileurl: data[3],
          md5: data[4],
          reason: data[5],
          expire_timestamp: data[6],
          expire_block_height: data[7]
        }
      }
      break
    case "vote":
      txInner = {
        type: "governance/vote",
        data: { proposal_id: data[0], answer: data[2] }
      }
      break
    default:
      // junk
      txInner = { type: "abc", data: {} }
      break
  }
  logger.debug(txInner)
  const hexData = "0x" + new Buffer(JSON.stringify(txInner)).toString("hex")

  // client side sign
  const rawTx = {
    nonce: "0x" + nonce.toString(16),
    from: account.addr,
    data: hexData,
    chainId: web3.net.id
  }
  const tx = new Tx(rawTx)
  tx.sign(privateKey)
  const signed = "0x" + tx.serialize().toString("hex")

  web3.cmt.sendRawTx(signed, function(err, res) {
    fnExpect(res)
    done()
  })
}
