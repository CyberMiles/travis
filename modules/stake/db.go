package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"path"

	"fmt"
	"github.com/CyberMiles/travis/types"
)

func getDb() *sql.DB {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := path.Join(rootDir, "data", "travis.db")

	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func composeQueryClause(cond map[string]interface{}) string {
	if cond == nil || len(cond) == 0 {
		return ""
	}

	clause := ""
	for k, v := range cond {
		s := ""
		switch v.(type) {
		case string:
			s = fmt.Sprintf("%s = '%s'", k, v)
		default:
			s = fmt.Sprintf("%s = %v", k, v)
		}

		if len(clause) == 0 {
			clause = s
		} else {
			clause = fmt.Sprintf("%s and %s", clause, s)
		}
	}

	if len(clause) != 0 {
		clause = fmt.Sprintf(" where %s", clause)
	}

	return clause
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

func GetCandidateByPubKey(pubKey string) *Candidate {
	cond := make(map[string]interface{})
	cond["pub_key"] = pubKey
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidates() (candidates Candidates) {
	cond := make(map[string]interface{})
	return getCandidatesInternal(cond)
}

func GetBackupValidators() (candidates Candidates) {
	cond := make(map[string]interface{})
	cond["state"] = "Backup Validator"
	cond["active"] = "Y"
	return getCandidatesInternal(cond)
}

func getCandidatesInternal(cond map[string]interface{}) (candidates Candidates) {
	db := getDb()
	defer db.Close()

	clause := composeQueryClause(cond)
	rows, err := db.Query("select pub_key, address, shares, voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, block_height, rank, state, created_at, updated_at from candidates" + clause)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, address, createdAt, updatedAt, shares, maxShares, name, website, location, profile, email, state, verified, active, compRate string
		var votingPower, blockHeight, rank int64
		err = rows.Scan(&pubKey, &address, &shares, &votingPower, &maxShares, &compRate, &name, &website, &location, &profile, &email, &verified, &active, &blockHeight, &rank, &state, &createdAt, &updatedAt)
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
		candidate := &Candidate{
			PubKey:       pk,
			OwnerAddress: address,
			Shares:       shares,
			VotingPower:  votingPower,
			MaxShares:    maxShares,
			CompRate:     compRate,
			Description:  description,
			Verified:     verified,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			Active:       active,
			BlockHeight:  blockHeight,
			Rank:         rank,
			State:        state,
		}
		candidates = append(candidates, candidate)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func SaveCandidate(candidate *Candidate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, hash, block_height, rank, state, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		types.PubKeyString(candidate.PubKey),
		candidate.OwnerAddress,
		candidate.Shares,
		candidate.VotingPower,
		candidate.MaxShares,
		candidate.CompRate,
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
		candidate.CreatedAt,
		candidate.UpdatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func updateCandidate(candidate *Candidate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, max_shares = ?, comp_rate = ?, name =?, website = ?, location = ?, profile = ?, email = ?, verified = ?, active = ?, hash = ?, rank = ?, state = ?, updated_at = ? where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.OwnerAddress,
		candidate.Shares,
		candidate.VotingPower,
		candidate.MaxShares,
		candidate.CompRate,
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
		candidate.UpdatedAt,
		types.PubKeyString(candidate.PubKey),
	)
	if err != nil {
		panic(err)
	}
}

func removeCandidate(candidate *Candidate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from candidates where address = ?")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from candidates where shares = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec("0")
	if err != nil {
		panic(err)
	}
}

func SaveDelegator(delegator *Delegator) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into delegators(address, created_at) values(?, ?)")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from delegators where address = ?")
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
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select address, created_at from delegators where address = ?")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into delegations(delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, comp_rate, hash, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), types.PubKeyString(d.PubKey), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, d.CompRate, common.Bytes2Hex(d.Hash()), d.CreatedAt, d.UpdatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveDelegation(delegation *Delegation) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from delegations where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(delegation.DelegatorAddress.String(), types.PubKeyString(delegation.PubKey))
	if err != nil {
		panic(err)
	}
}

func UpdateDelegation(d *Delegation) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update delegations set delegate_amount = ?, award_amount =?, withdraw_amount = ?, slash_amount = ?, hash = ?, updated_at = ? where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, common.Bytes2Hex(d.Hash()), d.UpdatedAt, d.DelegatorAddress.String(), types.PubKeyString(d.PubKey))
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

func GetDelegationsByPubKey(pubKey types.PubKey) (delegations []*Delegation) {
	cond := make(map[string]interface{})
	cond["pub_key"] = types.PubKeyString(pubKey)
	return getDelegationsInternal(cond)
}

func GetDelegationsByDelegator(delegatorAddress common.Address) (delegations []*Delegation) {
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	return getDelegationsInternal(cond)
}

func getDelegationsInternal(cond map[string]interface{}) (delegations []*Delegation) {
	db := getDb()
	defer db.Close()

	clause := composeQueryClause(cond)
	rows, err := db.Query("select delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, comp_rate, created_at, updated_at from delegations" + clause)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var delegatorAddress, pubKey, delegateAmount, awardAmount, withdrawAmount, slashAmount, compRate, createdAt, updatedAt string
		err = rows.Scan(&delegatorAddress, &pubKey, &delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &compRate, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, err := types.GetPubKey(pubKey)
		if err != nil {
			return
		}

		delegation := &Delegation{
			DelegatorAddress: common.HexToAddress(delegatorAddress),
			PubKey:           pk,
			DelegateAmount:   delegateAmount,
			AwardAmount:      awardAmount,
			WithdrawAmount:   withdrawAmount,
			SlashAmount:      slashAmount,
			CompRate:         compRate,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
		delegations = append(delegations, delegation)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return
}

func saveDelegateHistory(delegateHistory *DelegateHistory) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into delegate_history(delegator_address, pub_key, amount, op_code, created_at) values(?, ?, ?, ?, ?)")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into punish_history(pub_key, slashing_ratio, slash_amount, reason, created_at) values(?, ?, ?, ?, ?)")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into unstake_requests(id, delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, hash, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.Id,
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
	defer db.Close()

	rows, err := db.Query("select id, delegator_address, pub_key, initiated_block_height, performed_block_height, amount, state, created_at, updated_at from unstake_requests where state = ? and performed_block_height <= ?", "PENDING", height)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, delegatorAddress, pubKey, state, amount, createdAt, updatedAt string
		var initiatedBlockHeight, performedBlockHeight int64
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

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func updateUnstakeRequest(req *UnstakeRequest) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update unstake_requests set delegator_address = ?, pub_key = ?, initiated_block_height = ?, performed_block_height = ?, amount = ?, state = ?, hash=?, updated_at = ? where id = ?")
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
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into candidate_daily_stakes(id, pub_key, amount, created_at) values(?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(cds.Id, types.PubKeyString(cds.PubKey), cds.Amount, cds.CreatedAt)
	if err != nil {
		panic(err)
	}
}

func RemoveCandidateDailyStakes(pubKey types.PubKey, startDate string) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from candidate_daily_stakes where pub_key = ? and created_at < ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(types.PubKeyString(pubKey), startDate)
	if err != nil {
		panic(err)
	}
}

func GetCandidateDailyStakeMax(pubKey types.PubKey, startDate string) string {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select max(amount) from candidate_daily_stakes where pub_key = ? and created_at >= ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var maxAmount string
	err = stmt.QueryRow(types.PubKeyString(pubKey), startDate).Scan(&maxAmount)

	if err != nil {
		panic(err)
	}

	return maxAmount
}
