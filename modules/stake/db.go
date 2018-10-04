package stake

import (
	"database/sql"
	"fmt"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/types"
	"github.com/ethereum/go-ethereum/common"
)

var (
	deliverSqlTx *sql.Tx
)

func SetDeliverSqlTx(tx *sql.Tx) {
	deliverSqlTx = tx
}

func ResetDeliverSqlTx() {
	deliverSqlTx = nil
}

func getDb() *sql.DB {
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		panic(err)
	}
	return db
}

type SqlTxWrapper struct {
	tx        *sql.Tx
	withBlock bool
}

func getSqlTxWrapper() *SqlTxWrapper {
	var wrapper = &SqlTxWrapper{
		tx:        deliverSqlTx,
		withBlock: true,
	}
	if wrapper.tx == nil {
		db := getDb()
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		wrapper.tx = tx
		wrapper.withBlock = false
	}
	return wrapper
}

func (wrapper *SqlTxWrapper) Commit() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Commit(); err != nil {
			panic(err)
		}
	}
}

func (wrapper *SqlTxWrapper) Rollback() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Rollback(); err != nil {
			panic(err)
		}
	}
}

func buildQueryClause(cond map[string]interface{}) (clause string, params []interface{}) {
	if cond == nil || len(cond) == 0 {
		return "", nil
	}

	clause = ""
	for k, v := range cond {
		s := fmt.Sprintf("%s = ?", k)

		if len(clause) == 0 {
			clause = s
		} else {
			clause = fmt.Sprintf("%s and %s", clause, s)
		}
		params = append(params, v)
	}

	if len(clause) != 0 {
		clause = fmt.Sprintf(" where %s", clause)
	}

	return
}

func GetCandidateById(id int64) *Candidate {
	cond := make(map[string]interface{})
	cond["id"] = id
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidateByAddress(address common.Address) *Candidate {
	cond := make(map[string]interface{})
	cond["address"] = address.String()
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidateByPubKey(pubKey types.PubKey) *Candidate {
	cond := make(map[string]interface{})
	cond["pub_key"] = types.PubKeyString(pubKey)
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidates() (candidates Candidates) {
	cond := make(map[string]interface{})
	candidates = getCandidatesInternal(cond)
	return candidates
}

func GetActiveCandidates() (candidates Candidates) {
	cond := make(map[string]interface{})
	cond["active"] = "Y"
	candidates = getCandidatesInternal(cond)
	return candidates
}

func GetBackupValidators() (candidates Candidates) {
	cond := make(map[string]interface{})
	cond["state"] = "Backup Validator"
	cond["active"] = "Y"
	return getCandidatesInternal(cond)
}

func getCandidatesInternal(cond map[string]interface{}) (candidates Candidates) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, pub_key, address, shares, voting_power, pending_voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, block_height, rank, state, num_of_delegators, created_at from candidates"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

func composeCandidateResults(rows *sql.Rows) (candidates Candidates) {
	for rows.Next() {
		var pubKey, address, createdAt, shares, maxShares, name, website, location, profile, email, state, verified, active, compRate string
		var id, votingPower, pendingVotingPower, blockHeight, rank, numOfDelegators int64
		err := rows.Scan(&id, &pubKey, &address, &shares, &votingPower, &pendingVotingPower, &maxShares, &compRate, &name, &website, &location, &profile, &email, &verified, &active, &blockHeight, &rank, &state, &numOfDelegators, &createdAt)
		if err != nil {
			panic(err)
		}
		pk, _ := types.GetPubKey(pubKey)
		description := Description{
			Name:     name,
			Website:  website,
			Location: location,
			Profile:  profile,
			Email:    email,
		}
		c, _ := sdk.NewRatFromString(compRate)
		candidate := &Candidate{
			Id:                 id,
			PubKey:             pk,
			OwnerAddress:       address,
			Shares:             shares,
			VotingPower:        votingPower,
			PendingVotingPower: pendingVotingPower,
			MaxShares:          maxShares,
			CompRate:           c,
			Description:        description,
			Verified:           verified,
			CreatedAt:          createdAt,
			Active:             active,
			BlockHeight:        blockHeight,
			Rank:               rank,
			State:              state,
			NumOfDelegators:    numOfDelegators,
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func SaveCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, pending_voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, hash, block_height, rank, state, num_of_delegators, created_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		types.PubKeyString(candidate.PubKey),
		candidate.OwnerAddress,
		candidate.Shares,
		candidate.VotingPower,
		candidate.PendingVotingPower,
		candidate.MaxShares,
		candidate.CompRate.String(),
		candidate.Description.Name,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Profile,
		candidate.Description.Email,
		candidate.Verified,
		candidate.Active,
		common.Bytes2Hex(candidate.Hash()),
		candidate.BlockHeight,
		candidate.Rank,
		candidate.State,
		candidate.NumOfDelegators,
		candidate.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func updateCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, pending_voting_power = ?, max_shares = ?, comp_rate = ?, name =?, website = ?, location = ?, profile = ?, email = ?, verified = ?, active = ?, hash = ?, rank = ?, state = ?, num_of_delegators = ? where id = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.OwnerAddress,
		candidate.Shares,
		candidate.VotingPower,
		candidate.PendingVotingPower,
		candidate.MaxShares,
		candidate.CompRate.String(),
		candidate.Description.Name,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Profile,
		candidate.Description.Email,
		candidate.Verified,
		candidate.Active,
		common.Bytes2Hex(candidate.Hash()),
		candidate.Rank,
		candidate.State,
		candidate.NumOfDelegators,
		candidate.Id,
	)
	if err != nil {
		panic(err)
	}
}

func cleanCandidates() {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidates set active = ? where shares = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("N", "0")
	if err != nil {
		panic(err)
	}
}

func GetCandidatesTotalShares() (res sdk.Int) {
	res = sdk.ZeroInt
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	rows, err := txWrapper.tx.Query("select shares from candidates where active = 'Y' and state in ('Validator', 'Backup Validator')")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var shares string
		err = rows.Scan(&shares)
		if err != nil {
			panic(err)
		}

		s, ok := sdk.NewIntFromString(shares)
		if !ok {
			panic(err)
		}
		res = res.Add(s)
	}

	return
}

func SaveDelegation(d *Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into delegations(delegator_address, candidate_id, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, hash, voting_power, state, block_height, average_staking_date, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), d.CandidateId, d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.PendingWithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.BlockHeight, d.AverageStakingDate, d.CreatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveDelegation(id int64) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update delegations set state = ? where id = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("N", id)
	if err != nil {
		panic(err)
	}
}

func UpdateDelegation(d *Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()
	stmt, err := txWrapper.tx.Prepare("update delegations set delegator_address = ?, delegate_amount = ?, award_amount =?, withdraw_amount = ?, pending_withdraw_amount = ?, slash_amount = ?, comp_rate = ?, hash = ?, voting_power = ?, state = ?, average_staking_date = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(d.DelegatorAddress.String(), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.PendingWithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.AverageStakingDate, d.Id)
	if err != nil {
		panic(err)
	}
}

func GetDelegation(delegatorAddress common.Address, candidateId int64) *Delegation {
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	cond["candidate_id"] = candidateId
	delegations := getDelegationsInternal(cond)
	if len(delegations) == 0 {
		return nil
	} else {
		return delegations[0]
	}
}

func GetDelegations(state string) (delegations []*Delegation) {
	cond := make(map[string]interface{})
	if state != "" {
		cond["state"] = state
	}
	return getDelegationsInternal(cond)
}

func GetDelegationsByCandidate(candidateId int64, state string) (delegations []*Delegation) {
	cond := make(map[string]interface{})
	cond["candidate_id"] = candidateId
	if state != "" {
		cond["state"] = state
	}

	return getDelegationsInternal(cond)
}

func getDelegationsInternal(cond map[string]interface{}) (delegations []*Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, delegator_address, candidate_id, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, voting_power, state, block_height, average_staking_date, created_at from delegations"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	delegations = composeDelegationResults(rows)
	return
}

func composeDelegationResults(rows *sql.Rows) (delegations []*Delegation) {
	for rows.Next() {
		var delegatorAddress, delegateAmount, awardAmount, withdrawAmount, pendingWithdrawAmount, slashAmount, compRate, state, createdAt string
		var id, votingPower, blockHeight, averageStakingDate, candidateId int64
		err := rows.Scan(&id, &delegatorAddress, &candidateId, &delegateAmount, &awardAmount, &withdrawAmount, &pendingWithdrawAmount, &slashAmount, &compRate, &votingPower, &state, &blockHeight, &averageStakingDate, &createdAt)
		if err != nil {
			panic(err)
		}

		c, _ := sdk.NewRatFromString(compRate)
		delegation := &Delegation{
			Id:                    id,
			DelegatorAddress:      common.HexToAddress(delegatorAddress),
			CandidateId:           candidateId,
			DelegateAmount:        delegateAmount,
			AwardAmount:           awardAmount,
			WithdrawAmount:        withdrawAmount,
			PendingWithdrawAmount: pendingWithdrawAmount,
			SlashAmount:           slashAmount,
			CompRate:              c,
			VotingPower:           votingPower,
			State:                 state,
			BlockHeight:           blockHeight,
			AverageStakingDate:    averageStakingDate,
			CreatedAt:             createdAt,
		}
		delegations = append(delegations, delegation)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func saveDelegateHistory(h *DelegateHistory) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into delegate_history(delegator_address, candidate_id, amount, op_code, block_height, hash) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		h.DelegatorAddress.String(),
		h.CandidateId,
		h.Amount.String(),
		h.OpCode,
		h.BlockHeight,
		common.Bytes2Hex(h.Hash()),
	)
	if err != nil {
		panic(err)
	}
}

func saveSlash(slash *Slash) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into slashes(candidate_id, slash_ratio, slash_amount, reason, created_at, block_height, hash) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		slash.CandidateId,
		slash.SlashRatio.String(),
		slash.SlashAmount.String(),
		slash.Reason,
		slash.CreatedAt,
		slash.BlockHeight,
		common.Bytes2Hex(slash.Hash()),
	)
	if err != nil {
		panic(err)
	}
}

func saveUnstakeRequest(req *UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into unstake_requests(delegator_address, candidate_id, initiated_block_height, performed_block_height, amount, state, hash) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.DelegatorAddress.String(),
		req.CandidateId,
		req.InitiatedBlockHeight,
		req.PerformedBlockHeight,
		req.Amount,
		req.State,
		common.Bytes2Hex(req.Hash()),
	)
	if err != nil {
		panic(err)
	}
}

func GetUnstakeRequests(height int64) (reqs []*UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	rows, err := txWrapper.tx.Query("select id, delegator_address, candidate_id, initiated_block_height, performed_block_height, amount, state from unstake_requests where state = ? and performed_block_height <= ?", "PENDING", height)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	reqs = composeUnstakeRequestResults(rows)
	return
}

func composeUnstakeRequestResults(rows *sql.Rows) (reqs []*UnstakeRequest) {
	for rows.Next() {
		var delegatorAddress, state, amount string
		var id, initiatedBlockHeight, performedBlockHeight, candidateId int64
		err := rows.Scan(&id, &delegatorAddress, &candidateId, &initiatedBlockHeight, &performedBlockHeight, &amount, &state)
		if err != nil {
			panic(err)
		}

		req := &UnstakeRequest{
			Id:                   id,
			DelegatorAddress:     common.HexToAddress(delegatorAddress),
			CandidateId:          candidateId,
			InitiatedBlockHeight: initiatedBlockHeight,
			PerformedBlockHeight: performedBlockHeight,
			State:                state,
			Amount:               amount,
		}
		reqs = append(reqs, req)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func getUnstakeRequestsInternal(cond map[string]interface{}) (reqs []*UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, delegator_address, candidate_id, initiated_block_height, performed_block_height, amount, state from unstake_requests"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	reqs = composeUnstakeRequestResults(rows)
	return
}

func GetUnstakeRequestsByDelegator(delegatorAddress common.Address) []*UnstakeRequest {
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	cond["state"] = "PENDING"
	return getUnstakeRequestsInternal(cond)
}

func updateUnstakeRequest(req *UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update unstake_requests set delegator_address = ?, candidate_id = ?, initiated_block_height = ?, performed_block_height = ?, amount = ?, state = ?, hash=? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.DelegatorAddress.String(),
		req.CandidateId,
		req.InitiatedBlockHeight,
		req.PerformedBlockHeight,
		req.Amount,
		req.State,
		common.Bytes2Hex(req.Hash()),
		req.Id,
	)
	if err != nil {
		panic(err)
	}
}

func SaveCandidateDailyStake(cds *CandidateDailyStake) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidate_daily_stakes(candidate_id, amount, block_height, hash) values(?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(cds.CandidateId, cds.Amount, cds.BlockHeight, common.Bytes2Hex(cds.Hash()))
	if err != nil {
		panic(err)
	}
}

func RemoveExpiredCandidateDailyStakes(blockHeight int64) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("delete from candidate_daily_stakes where block_height < ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(blockHeight)
	if err != nil {
		panic(err)
	}
}

func GetCandidateDailyStakeMaxValue(candidateId int64, startBlockHeight int64) (res sdk.Int) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select max(amount) from candidate_daily_stakes where candidate_id = ? and block_height >= ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var maxAmount string
	err = stmt.QueryRow(candidateId, startBlockHeight).Scan(&maxAmount)

	if err != nil {
		panic(err)
	}

	res, _ = sdk.NewIntFromString(maxAmount)
	return
}

func saveCandidateAccountUpdateRequest(req *CandidateAccountUpdateRequest) int64 {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidate_account_update_requests(candidate_id, from_address, to_address, created_block_height, accepted_block_height, state, hash) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		req.CandidateId,
		req.FromAddress.String(),
		req.ToAddress.String(),
		req.CreatedBlockHeight,
		req.AcceptedBlockHeight,
		req.State,
		common.Bytes2Hex(req.Hash()),
	)
	if err != nil {
		panic(err)
	}

	lastInsertId, _ := result.LastInsertId()
	return lastInsertId
}

func getCandidateAccountUpdateRequest(id int64) *CandidateAccountUpdateRequest {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select candidate_id, from_address, to_address, created_block_height, accepted_block_height, state from candidate_account_update_requests where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var candidateId, createdBlockHeight, acceptedBlockHeight int64
	var fromAddress, toAddress, state string
	err = stmt.QueryRow(id).Scan(&candidateId, &fromAddress, &toAddress, &createdBlockHeight, &acceptedBlockHeight, &state)
	if err != nil {
		//panic(err)
		return nil
	}

	res := &CandidateAccountUpdateRequest{
		Id:                  id,
		CandidateId:         candidateId,
		FromAddress:         common.HexToAddress(fromAddress),
		ToAddress:           common.HexToAddress(toAddress),
		CreatedBlockHeight:  createdBlockHeight,
		AcceptedBlockHeight: acceptedBlockHeight,
		State:               state,
	}

	return res
}

func updateCandidateAccountUpdateRequest(req *CandidateAccountUpdateRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidate_account_update_requests set accepted_block_height = ?, state = ?, hash = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.AcceptedBlockHeight,
		req.State,
		common.Bytes2Hex(req.Hash()),
		req.Id,
	)
	if err != nil {
		panic(err)
	}
}
