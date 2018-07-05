module.exports = {
  TestMode: "cluster",

  Accounts: [],
  PubKeys: [
    "051FUvSNJmVL4UiFL7ucBr3TnGqG6a5JgUIgKf4UOIA=",
    "v0yMKq/chUKEhELdLp1HJfGAmHZJll8cEeskU5L97Mg=",
    "lmlbeRtIZLSgIvib9Emndk/W0isuGrJmBDlB+EwbYuY=",
    "GzGGwxzBnEj8RbFFAMgH+QP8bPsyRXrlTknpkt8mo5o="
  ],

  ValSizeLimit: 0.12,
  ValMinSelfStakingRatio: 0.1,

  UnstakeWaitingPeriod: (7 * 24 * 60 * 60) / 10,
  ProposalExpires: (7 * 24 * 60 * 60) / 10,

  AnnualInflation: 0.08,
  BlockAwards: (1000000000 * 0.08) / ((365 * 24 * 3600) / 10),

  GasPrice: 2e9,
  GasLimit: {
    DeclareCandidacy: 1e6,
    UpdateCandidacy: 1e6,
    GovernancePropose: 2e6
  }
}
