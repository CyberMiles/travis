package stake

import (
	"database/sql"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/cli"
	"math/big"
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

	var pubKey, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active string
	var votingPower, cut int64
	err = stmt.QueryRow(address.String()).Scan(&pubKey, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := utils.GetPubKey(pubKey)
	s := new(big.Int)
	s.SetString(shares, 10)
	ms := new(big.Int)
	ms.SetString(maxShares, 10)
	description := Description{
		Website:  website,
		Location: location,
		Details:  details,
	}
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: address,
		Shares:       s,
		VotingPower:  votingPower,
		MaxShares:    ms,
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

	var address, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active string
	var votingPower, cut int64
	err = stmt.QueryRow(pubKey).Scan(&address, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := utils.GetPubKey(pubKey)
	s := new(big.Int)
	s.SetString(shares, 10)
	ms := new(big.Int)
	ms.SetString(maxShares, 10)
	description := Description{
		Website:  website,
		Location: location,
		Details:  details,
	}
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: common.HexToAddress(address),
		Shares:       s,
		VotingPower:  votingPower,
		MaxShares:    ms,
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
		var pubKey, address, createdAt, updatedAt, shares, maxShares, website, location, details, verified, active string
		var votingPower, cut int64
		err = rows.Scan(&pubKey, &address, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &verified, &active, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := utils.GetPubKey(pubKey)
		s := new(big.Int)
		s.SetString(shares, 10)
		ms := new(big.Int)
		ms.SetString(maxShares, 10)
		description := Description{
			Website:  website,
			Location: location,
			Details:  details,
		}
		candidate := &Candidate{
			PubKey:       pk,
			OwnerAddress: common.HexToAddress(address),
			Shares:       s,
			VotingPower:  votingPower,
			MaxShares:    ms,
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
		utils.PubKeyString(candidate.PubKey),
		candidate.OwnerAddress.String(),
		candidate.Shares.String(),
		candidate.VotingPower,
		candidate.MaxShares.String(),
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
		candidate.Shares.String(),
		candidate.VotingPower,
		candidate.MaxShares.String(),
		candidate.Cut,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Details,
		candidate.Verified,
		candidate.Active,
		candidate.UpdatedAt,
		utils.PubKeyString(candidate.PubKey),
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

	_, err = stmt.Exec(d.DelegatorAddress.String(), utils.PubKeyString(d.PubKey), d.DelegateAmount.String(), d.AwardAmount.String(), d.WithdrawAmount.String(), d.SlashAmount.String(), d.CreatedAt, d.UpdatedAt)
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

	_, err = stmt.Exec(delegation.DelegatorAddress.String(), utils.PubKeyString(delegation.PubKey))
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

	_, err = stmt.Exec(d.DelegateAmount.String(), d.AwardAmount.String(), d.WithdrawAmount.String(), d.SlashAmount.String(), d.UpdatedAt, d.DelegatorAddress.String(), utils.PubKeyString(d.PubKey))
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

	_, err = stmt.Exec(d.DelegatorAddress.String(), d.UpdatedAt, originalAddress.String(), utils.PubKeyString(d.PubKey))
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
	err = stmt.QueryRow(delegatorAddress.String(), utils.PubKeyString(pubKey)).Scan(&delegateAmount, &awardAmount, &withdrawAmount, &slashAmount, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	d, _ := new(big.Int).SetString(delegateAmount, 10)
	a, _ := new(big.Int).SetString(awardAmount, 10)
	w, _ := new(big.Int).SetString(withdrawAmount, 10)
	s, _ := new(big.Int).SetString(slashAmount, 10)
	return &Delegation{
		DelegatorAddress: delegatorAddress,
		PubKey:           pubKey,
		DelegateAmount:   d,
		AwardAmount:      a,
		WithdrawAmount:   w,
		SlashAmount:      s,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

func GetDelegationsByPubKey(pubKey crypto.PubKey) (delegations []*Delegation) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select delegator_address, delegate_amount, award_amount, withdraw_amount, slash_amount, created_at, updated_at from delegations where pub_key = ?", utils.PubKeyString(pubKey))
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

		d, _ := new(big.Int).SetString(delegateAmount, 10)
		a, _ := new(big.Int).SetString(awardAmount, 10)
		w, _ := new(big.Int).SetString(withdrawAmount, 10)
		s, _ := new(big.Int).SetString(slashAmount, 10)
		delegation := &Delegation{
			DelegatorAddress: common.HexToAddress(delegatorAddress),
			PubKey:           pubKey,
			DelegateAmount:   d,
			AwardAmount:      a,
			WithdrawAmount:   w,
			SlashAmount:      s,
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

		d, _ := new(big.Int).SetString(delegateAmount, 10)
		a, _ := new(big.Int).SetString(awardAmount, 10)
		w, _ := new(big.Int).SetString(withdrawAmount, 10)
		s, _ := new(big.Int).SetString(slashAmount, 10)
		delegation := &Delegation{
			DelegatorAddress: delegatorAddress,
			PubKey:           pk,
			DelegateAmount:   d,
			AwardAmount:      a,
			WithdrawAmount:   w,
			SlashAmount:      s,
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
		utils.PubKeyString(delegateHistory.PubKey),
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
		utils.PubKeyString(punishHistory.PubKey),
		punishHistory.DeductionRatio,
		punishHistory.Deduction.String(),
		punishHistory.Reason,
		punishHistory.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}
