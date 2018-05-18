module.exports = {
  TestMode: "cluster",

  Accounts: [],
  PubKeys: [
    "051FUvSNJmVL4UiFL7ucBr3TnGqG6a5JgUIgKf4UOIA=",
    "v0yMKq/chUKEhELdLp1HJfGAmHZJll8cEeskU5L97Mg=",
    "lmlbeRtIZLSgIvib9Emndk/W0isuGrJmBDlB+EwbYuY=",
    "RqKjPhMuo/PkFkSJJabpqfys18kp9Rnl1WyccrcY5w4="
  ],

  ValSizeLimit: 0.1,
  ValMinSelfStakingRatio: 0.1,
  UnstakeWaitingPeriod: 7 * 24 * 60 * 60 / 10,
  AnnualInflation: 0.08,

  BlockAwards: 1000000000 * 0.08 / (365 * 24 * 3600 / 10),
  ProposalExpires: 7 * 24 * 60 * 60 / 10
}
