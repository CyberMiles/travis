package stake

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
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
	rows, err := txWrapper.tx.Query("select pub_key, address, shares, voting_power, pending_voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, block_height, rank, state, num_of_delegators, created_at, updated_at from candidates"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

func DeactivateValidators(pubKeys []string) {
	updateValidatorsActive(pubKeys, "N")
}

func ActivateValidators(pubKeys []string) {
	updateValidatorsActive(pubKeys, "Y")
}

func updateValidatorsActive(pubKeys []string, active string) {
	if pubKeys == nil || len(pubKeys) == 0 {
		return
	}

	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidates set active = ?, updated_at = ? where pub_key in (?" + strings.Repeat(",?", len(pubKeys)-1) + ")")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	args := []interface{}{active, utils.GetNow()}
	for _, pk := range pubKeys {
		args = append(args, pk)
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		panic(err)
	}
}

func composeCandidateResults(rows *sql.Rows) (candidates Candidates) {
	for rows.Next() {
		var pubKey, address, createdAt, updatedAt, shares, maxShares, name, website, location, profile, email, state, verified, active, compRate string
		var votingPower, pendingVotingPower, blockHeight, rank, numOfDelegators int64
		err := rows.Scan(&pubKey, &address, &shares, &votingPower, &pendingVotingPower, &maxShares, &compRate, &name, &website, &location, &profile, &email, &verified, &active, &blockHeight, &rank, &state, &numOfDelegators, &createdAt, &updatedAt)
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
			UpdatedAt:          updatedAt,
			Active:             active,
			BlockHeight:        blockHeight,
			Rank:               rank,
			State:              state,
			NumOfDelegator:     numOfDelegators,
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

	stmt, err := txWrapper.tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, pending_voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, hash, block_height, rank, state, num_of_delegators, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
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
		candidate.NumOfDelegator,
		candidate.CreatedAt,
		candidate.UpdatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func updateCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, pending_voting_power = ?, max_shares = ?, comp_rate = ?, name =?, website = ?, location = ?, profile = ?, email = ?, verified = ?, active = ?, hash = ?, rank = ?, state = ?, num_of_delegators = ?, updated_at = ? where pub_key = ?")
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
		candidate.NumOfDelegator,
		candidate.UpdatedAt,
		types.PubKeyString(candidate.PubKey),
	)
	if err != nil {
		panic(err)
	}
}

func cleanCandidates() {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("delete from candidates where shares = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("0")
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

	stmt, err := txWrapper.tx.Prepare("insert into delegations(delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, hash, voting_power, state, block_height, average_staking_date, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), types.PubKeyString(d.PubKey), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.PendingWithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.BlockHeight, d.AverageStakingDate, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveDelegation(id int64) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update delegations set state = ?, updated_at = ? where id = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("N", utils.GetNow(), id)
	if err != nil {
		panic(err)
	}
}

func UpdateDelegation(d *Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()
	stmt, err := txWrapper.tx.Prepare("update delegations set delegator_address = ?, delegate_amount = ?, award_amount =?, withdraw_amount = ?, pending_withdraw_amount = ?, slash_amount = ?, comp_rate = ?, hash = ?, voting_power = ?, state = ?, average_staking_date = ?, updated_at = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(d.DelegatorAddress.String(), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.PendingWithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.AverageStakingDate, d.UpdatedAt, d.Id)
	if err != nil {
		panic(err)
	}
}

func GetDelegation(delegatorAddress common.Address, pubKey types.PubKey) *Delegation {
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	cond["pub_key"] = types.PubKeyString(pubKey)
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

func GetDelegationsByPubKey(pubKey types.PubKey, state string) (delegations []*Delegation) {
	cond := make(map[string]interface{})
	cond["pub_key"] = types.PubKeyString(pubKey)
	if state != "" {
		cond["state"] = state
	}

	return getDelegationsInternal(cond)
}

func getDelegationsInternal(cond map[string]interface{}) (delegations []*Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, voting_power, state, block_height, average_staking_date, created_at, updated_at from delegations"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	delegations = composeDelegationResults(rows)
	return
}

func composeDelegationResults(rows *sql.Rows) (delegations []*Delegation) {
	for rows.Next() {
		var delegatorAddress, pubKey, delegateAmount, awardAmount, withdrawAmount, pendingWithdrawAmount, slashAmount, compRate, state, createdAt, updatedAt string
		var id, votingPower, blockHeight, averageStakingDate int64
		err := rows.Scan(&id, &delegatorAddress, &pubKey, &delegateAmount, &awardAmount, &withdrawAmount, &pendingWithdrawAmount, &slashAmount, &compRate, &votingPower, &state, &blockHeight, &averageStakingDate, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, err := types.GetPubKey(pubKey)
		if err != nil {
			return
		}

		c, _ := sdk.NewRatFromString(compRate)
		delegation := &Delegation{
			Id:                    id,
			DelegatorAddress:      common.HexToAddress(delegatorAddress),
			PubKey:                pk,
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
			UpdatedAt:             updatedAt,
		}
		delegations = append(delegations, delegation)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func saveDelegateHistory(delegateHistory *DelegateHistory) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into delegate_history(delegator_address, pub_key, amount, op_code, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		delegateHistory.DelegatorAddress.String(),
		types.PubKeyString(delegateHistory.PubKey),
		delegateHistory.Amount.String(),
		delegateHistory.OpCode,
		delegateHistory.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func saveSlash(slash *Slash) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into slashes(pub_key, slash_ratio, slash_amount, reason, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		types.PubKeyString(slash.PubKey),
		slash.SlashRatio.String(),
		slash.SlashAmount.String(),
		slash.Reason,
		slash.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func saveUnstakeRequest(req *UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into unstake_requests(delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, hash, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.DelegatorAddress.String(),
		types.PubKeyString(req.PubKey),
		req.InitiatedBlockHeight,
		req.PerformedBlockHeight,
		req.Amount,
		req.State,
		common.Bytes2Hex(req.Hash()),
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func GetUnstakeRequests(height int64) (reqs []*UnstakeRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	rows, err := txWrapper.tx.Query("select id, delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, created_at, updated_at from unstake_requests where state = ? and performed_block_height <= ?", "PENDING", height)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	reqs = composeUnstakeRequestResults(rows)
	return
}

func composeUnstakeRequestResults(rows *sql.Rows) (reqs []*UnstakeRequest) {
	for rows.Next() {
		var delegatorAddress, pubKey, state, amount, createdAt, updatedAt string
		var id, initiatedBlockHeight, performedBlockHeight int64
		err := rows.Scan(&id, &delegatorAddress, &pubKey, &initiatedBlockHeight, &performedBlockHeight, &amount, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := types.GetPubKey(pubKey)
		req := &UnstakeRequest{
			Id:                   id,
			DelegatorAddress:     common.HexToAddress(delegatorAddress),
			PubKey:               pk,
			InitiatedBlockHeight: initiatedBlockHeight,
			PerformedBlockHeight: performedBlockHeight,
			State:                state,
			Amount:               amount,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
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
	rows, err := txWrapper.tx.Query("select id, delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, created_at, updated_at from unstake_requests"+clause, params...)
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

	stmt, err := txWrapper.tx.Prepare("update unstake_requests set delegator_address = ?, pub_key = ?, initiated_block_height = ?, performed_block_height = ?, amount = ?, state = ?, hash=?, updated_at = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.DelegatorAddress.String(),
		types.PubKeyString(req.PubKey),
		req.InitiatedBlockHeight,
		req.PerformedBlockHeight,
		req.Amount,
		req.State,
		common.Bytes2Hex(req.Hash()),
		req.UpdatedAt,
		req.Id,
	)
	if err != nil {
		panic(err)
	}
}

func SaveCandidateDailyStake(cds *CandidateDailyStake) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidate_daily_stakes(pub_key, amount, created_at) values(?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(types.PubKeyString(cds.PubKey), cds.Amount, cds.CreatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveCandidateDailyStakes(pubKey types.PubKey, startDate string) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("delete from candidate_daily_stakes where pub_key = ? and created_at < ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(types.PubKeyString(pubKey), startDate)
	if err != nil {
		panic(err)
	}
}

func GetCandidateDailyStakeMaxValue(pubKey types.PubKey, startDate string) (res sdk.Int) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select max(amount) from candidate_daily_stakes where pub_key = ? and created_at >= ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var maxAmount string
	err = stmt.QueryRow(types.PubKeyString(pubKey), startDate).Scan(&maxAmount)

	if err != nil {
		panic(err)
	}

	res, _ = sdk.NewIntFromString(maxAmount)
	return
}
