
# 改造点
1. 字段名称
ChainId -> ChainID
ApiBackend -> APIBackend

2. GasLimit及gas类型改为了uint64，相关计算需要更改
tx.Gas()

3. stateDB CommitTo -> Commit


....



对go-ethereum本身的修改：
1.

在consensus/ethash/consensus.go中增加以下方法
// TODO:
func AccumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain progression
	accumulateRewards(config, state, header, uncles)
}
