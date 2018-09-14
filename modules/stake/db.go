package stake

import (
	"database/sql"
	"fmt"
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

func removeCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("delete from candidates where address = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(candidate.OwnerAddress)
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

func SaveDelegator(delegator *Delegator) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into delegators(address, created_at) values(?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(delegator.Address.String(), delegator.CreatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveDelegator(delegator *Delegator) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("delete from delegators where address = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(delegator.Address.String())
	if err != nil {
		panic(err)
	}
}

func GetDelegator(address string) *Delegator {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()
	stmt, err := txWrapper.tx.Prepare("select address, created_at from delegators where address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var updatedAt string
	err = stmt.QueryRow(address).Scan(&updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return &Delegator{common.HexToAddress(address), updatedAt}
}

func SaveDelegation(d *Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into delegations(delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, comp_rate, hash, voting_power, state, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), types.PubKeyString(d.PubKey), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveDelegation(delegatorAddress common.Address, pubKey types.PubKey) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update delegations set state = ?, updated_at = ? where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("N", utils.GetNow(), delegatorAddress.String(), types.PubKeyString(pubKey))
	if err != nil {
		panic(err)
	}
}

func UpdateDelegation(d *Delegation) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()
	stmt, err := txWrapper.tx.Prepare("update delegations set delegate_amount = ?, award_amount =?, withdraw_amount = ?, slash_amount = ?, comp_rate = ?, hash = ?, voting_power = ?, state = ?, updated_at = ? where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.UpdatedAt, d.DelegatorAddress.String(), types.PubKeyString(d.PubKey))
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
	rows, err := txWrapper.tx.Query("select delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, comp_rate, voting_power, state, created_at, updated_at from delegations"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	delegations = composeDelegationResults(rows)
	return
}

func composeDelegationResults(rows *sql.Rows) (delegations []*Delegation) {
	for rows.Next() {
		var delegatorAddress, pubKey, delegateAmount, awardAmount, withdrawAmount, slashAmount, compRate, state, createdAt, updatedAt string
		var votingPower int64
		err := rows.Scan(&delegatorAddress, &pubKey, &delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &compRate, &votingPower, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, err := types.GetPubKey(pubKey)
		if err != nil {
			return
		}

		c, _ := sdk.NewRatFromString(compRate)
		delegation := &Delegation{
			DelegatorAddress: common.HexToAddress(delegatorAddress),
			PubKey:           pk,
			DelegateAmount:   delegateAmount,
			AwardAmount:      awardAmount,
			WithdrawAmount:   withdrawAmount,
			SlashAmount:      slashAmount,
			CompRate:         c,
			VotingPower:      votingPower,
			State:            state,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
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

func savePunishHistory(punishHistory *PunishHistory) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into punish_history(pub_key, slashing_ratio, slash_amount, reason, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		types.PubKeyString(punishHistory.PubKey),
		punishHistory.SlashingRatio.String(),
		punishHistory.SlashAmount.String(),
		punishHistory.Reason,
		punishHistory.CreatedAt,
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
	db := getDb()

	rows, err := db.Query("select id, delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, created_at, updated_at from unstake_requests where state = ? and performed_block_height <= ?", "PENDING", height)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var delegatorAddress, pubKey, state, amount, createdAt, updatedAt string
		var id, initiatedBlockHeight, performedBlockHeight int64
		err = rows.Scan(&id, &delegatorAddress, &pubKey, &initiatedBlockHeight, &performedBlockHeight, &amount, &state, &createdAt, &updatedAt)
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
		// If no records found
		res = sdk.E18Int
		return
	}

	res, _ = sdk.NewIntFromString(maxAmount)
	return
}
