package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
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
	stmt, err := db.Prepare("select pub_key, shares, voting_power, max_shares, cut, website, location, details, created_at, updated_at from candidates where address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var pubKey, createdAt, updatedAt, shares, maxShares, website, location, details string
	var votingPower int64
	var cut float64
	err = stmt.QueryRow(address.String()).Scan(&pubKey, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
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
	}
}

func GetCandidateByPubKey(pubKey string) *Candidate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select address, shares, voting_power, max_shares, cut, website, location, details, created_at, updated_at from candidates where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var address, createdAt, updatedAt, shares, maxShares, website, location, details string
	var votingPower int64
	var cut float64
	err = stmt.QueryRow(pubKey).Scan(&pubKey, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
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
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func GetCandidates() (candidates Candidates) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select pub_key, address, shares, voting_power, max_shares, cut, website, location, details, created_at, updated_at from candidates")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, address, createdAt, updatedAt, shares, maxShares, website, location, details string
		var votingPower int64
		var cut float64
		err = rows.Scan(&pubKey, &address, &shares, &votingPower, &maxShares, &cut, &website, &location, &details, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := GetPubKey(pubKey)
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
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
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

	stmt, err := tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, max_shares, cut, website, location, details, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.PubKey.KeyString(),
		candidate.OwnerAddress.String(),
		candidate.Shares.String(),
		candidate.VotingPower,
		candidate.MaxShares.String(),
		candidate.Cut,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Details,
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

	stmt, err := tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, max_shares = ?, cut = ?, website = ?, location = ?, details = ?, updated_at = ? where address = ?")
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
		candidate.UpdatedAt,
		candidate.OwnerAddress.String(),
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
	// todo
}

func RemoveDelegator(delegator *Delegator) {
	// todo
}

func UpdateDelegator(delegator *Delegator) {
	// todo
}

func GetDelegator(address string) *Delegator {
	// todo
	return nil
}

func SaveDelegation(delegation *Delegation) {
	// todo
}

func RemoveDelegation(delegation *Delegation) {
	// todo
}

func UpdateDelegation(delegation *Delegation) {
	// todo
}

func GetDelegation(address string) *Delegation {
	// todo
	return nil
}

func saveDelegateHistory(delegateHistory DelegateHistory) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into delegate_history(delegator_address, candidate_address, shares, op_code, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		delegateHistory.DelegatorAddress.String(),
		delegateHistory.CandidateAddress.String(),
		delegateHistory.Shares.String(),
		delegateHistory.OpCode,
		delegateHistory.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}
