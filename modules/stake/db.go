package stake

import (
	"database/sql"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/cli"
	"path"
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

func GetCandidateByAddress(address common.Address) *Candidate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select pub_key, shares, voting_power, max_shares, cut, website, location, details, verified, active, created_at, updated_at from candidates where address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var pubKey, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active, cut string
	var votingPower int64
	err = stmt.QueryRow(address.String()).Scan(&pubKey, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := utils.GetPubKey(pubKey)
	description := Description{
		Website:  website,
		Location: location,
		Details:  details,
	}
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: address,
		Shares:       shares,
		VotingPower:  votingPower,
		MaxShares:    maxShares,
		Cut:          cut,
		Description:  description,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		Verified:     verified,
		Active:       active,
	}
}

func GetCandidateByPubKey(pubKey string) *Candidate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select address, shares, voting_power, max_shares, cut, website, location, details, verified, active, created_at, updated_at from candidates where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var address, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active, cut string
	var votingPower int64
	err = stmt.QueryRow(pubKey).Scan(&address, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := utils.GetPubKey(pubKey)
	description := Description{
		Website:  website,
		Location: location,
		Details:  details,
	}
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: common.HexToAddress(address),
		Shares:       shares,
		VotingPower:  votingPower,
		MaxShares:    maxShares,
		Cut:          cut,
		Description:  description,
		Verified:     verified,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		Active:       active,
	}
}

func GetCandidates() (candidates Candidates) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select pub_key, address, shares, voting_power, max_shares, cut, website, location, details, verified, active, created_at, updated_at from candidates where active=?", "Y")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, address, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active, cut string
		var votingPower int64
		err = rows.Scan(&pubKey, &address, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := utils.GetPubKey(pubKey)
		description := Description{
			Website:  website,
			Location: location,
			Details:  details,
		}
		candidate := &Candidate{
			PubKey:       pk,
			OwnerAddress: common.HexToAddress(address),
			Shares:       shares,
			VotingPower:  votingPower,
			MaxShares:    maxShares,
			Cut:          cut,
			Description:  description,
			Verified:     verified,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			Active:       active,
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

	stmt, err := tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, max_shares, cut, website, location, details, verified, active, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.PubKey.KeyString(),
		candidate.OwnerAddress.String(),
		candidate.Shares,
		candidate.VotingPower,
		candidate.MaxShares,
		candidate.Cut,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Details,
		candidate.Verified,
		candidate.Active,
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

	stmt, err := tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, max_shares = ?, cut = ?, website = ?, location = ?, details = ?, verified = ?, active = ?, updated_at = ? where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.OwnerAddress.String(),
		candidate.Shares,
		candidate.VotingPower,
		candidate.MaxShares,
		candidate.Cut,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Details,
		candidate.Verified,
		candidate.Active,
		candidate.UpdatedAt,
		candidate.PubKey.KeyString(),
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

	_, err = stmt.Exec(candidate.OwnerAddress.String())
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

	stmt, err := tx.Prepare("insert into delegations(delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), d.PubKey.KeyString(), d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, d.CreatedAt, d.UpdatedAt)
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

	_, err = stmt.Exec(delegation.DelegatorAddress.String(), delegation.PubKey.KeyString())
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

	stmt, err := tx.Prepare("update delegations set delegate_amount = ?, award_amount =?, withdraw_amount = ?, slash_amount = ?, updated_at = ? where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.SlashAmount, d.UpdatedAt, d.DelegatorAddress.String(), d.PubKey.KeyString())
	if err != nil {
		panic(err)
	}
}

func UpdateDelegatorAddress(d *Delegation, originalAddress common.Address) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update delegations set delegator_address = ?, updated_at = ? where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), d.UpdatedAt, originalAddress.String(), d.PubKey.KeyString())
	if err != nil {
		panic(err)
	}
}

func GetDelegation(delegatorAddress common.Address, pubKey crypto.PubKey) *Delegation {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select delegate_amount, award_amount, withdraw_amount, slash_amount, created_at, updated_at from delegations where delegator_address = ? and pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var delegateAmount, awardAmount, withdrawAmount, slashAmount, createdAt, updatedAt string
	err = stmt.QueryRow(delegatorAddress.String(), pubKey.KeyString()).Scan(&delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return &Delegation{
		DelegatorAddress: delegatorAddress,
		PubKey:           pubKey,
		DelegateAmount:   delegateAmount,
		AwardAmount:      awardAmount,
		WithdrawAmount:   withdrawAmount,
		SlashAmount:      slashAmount,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

func GetDelegationsByPubKey(pubKey crypto.PubKey) (delegations []*Delegation) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select delegator_address, delegate_amount, award_amount, withdraw_amount, slash_amount, created_at, updated_at from delegations where pub_key = ?", pubKey.KeyString())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var delegatorAddress, delegateAmount, awardAmount, withdrawAmount, slashAmount, createdAt, updatedAt string
		err = rows.Scan(&delegatorAddress, &delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		delegation := &Delegation{
			DelegatorAddress: common.HexToAddress(delegatorAddress),
			PubKey:           pubKey,
			DelegateAmount:   delegateAmount,
			AwardAmount:      awardAmount,
			WithdrawAmount:   withdrawAmount,
			SlashAmount:      slashAmount,
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

func GetDelegationsByDelegator(delegatorAddress common.Address) (delegations []*Delegation) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, created_at, updated_at from delegations where delegator_address = ?", delegatorAddress.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, delegateAmount, awardAmount, withdrawAmount, slashAmount, createdAt, updatedAt string
		err = rows.Scan(&pubKey, &delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, err := utils.GetPubKey(pubKey)
		if err != nil {
			return
		}

		delegation := &Delegation{
			DelegatorAddress: delegatorAddress,
			PubKey:           pk,
			DelegateAmount:   delegateAmount,
			AwardAmount:      awardAmount,
			WithdrawAmount:   withdrawAmount,
			SlashAmount:      slashAmount,
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
		delegateHistory.PubKey.KeyString(),
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

	stmt, err := tx.Prepare("insert into punish_history(pub_key, deduction_ratio, deduction, reason, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		punishHistory.PubKey.KeyString(),
		punishHistory.DeductionRatio,
		punishHistory.Deduction.String(),
		punishHistory.Reason,
		punishHistory.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}
