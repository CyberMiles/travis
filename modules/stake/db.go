package stake

import (
	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
	"strings"
	"github.com/cosmos/cosmos-sdk"
	"encoding/hex"
)

func getDb() *sql.DB {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := path.Join(rootDir, "data", "stake.db")

	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func GetCandidate(pubKey string) *Candidate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select owner_address, shares, voting_power from candidates where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var ownerAddress string
	var shares, votingPower uint64
	err = stmt.QueryRow(strings.ToUpper(pubKey)).Scan(&ownerAddress, &shares, &votingPower)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
	bs, _ := hex.DecodeString(ownerAddress)
	return &Candidate{
		PubKey:      pk,
		Owner:       sdk.NewActor(stakingModuleName, bs),
		Shares:      shares,
		VotingPower: votingPower,
	}
}

func GetCandidates() (candidates Candidates) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select pub_key, owner_address, shares, voting_power from candidates")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, ownerAddress string
		var shares, votingPower uint64
		err = rows.Scan(&pubKey, &ownerAddress, &shares, &votingPower)
		if err != nil {
			panic(err)
		}

		pk, _ := GetPubKey(pubKey)
		bs, _ := hex.DecodeString(ownerAddress)
		candidate := &Candidate{
			PubKey:      pk,
			Owner:       sdk.NewActor(stakingModuleName, bs),
			Shares:      shares,
			VotingPower: votingPower,
		}
		candidates = append(candidates, candidate)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func saveCandidate(candidate *Candidate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into candidates(pub_key, owner_address, shares, voting_power) values(?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.PubKey.KeyString(), candidate.Owner.Address.String(), candidate.Shares, candidate.VotingPower)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func updateCandidate(candidate *Candidate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("update  candidates set shares = ?, voting_power = ? where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.Shares, candidate.VotingPower, candidate.PubKey.KeyString())
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func removeCandidate(pubKey string) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("delete from candidates where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(pubKey)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func saveSlot(slot *Slot) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into slots(id, validator_pub_key, total_amount, available_amount, proposed_roi) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(slot.Id), slot.ValidatorPubKey.KeyString(), slot.TotalAmount, slot.AvailableAmount, slot.ProposedRoi)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func updateSlot(slot *Slot) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("update slots set available_amount = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.AvailableAmount, strings.ToUpper(slot.Id))
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func GetSlot(slotId string) *Slot {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select validator_pub_key, total_amount, available_amount, proposed_roi from slots where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var validatorPubKey string
	var totalAmount, availableAmount, proposedRoi int64
	err = stmt.QueryRow(strings.ToUpper(slotId)).Scan(&validatorPubKey, &totalAmount, &availableAmount, &proposedRoi)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(validatorPubKey)
	return NewSlot(slotId, pk, totalAmount, availableAmount, proposedRoi)
}

func GetSlots() (slots []*Slot) {
	db := getDb()
	defer db.Close()
	rows, err := db.Query("select id, validator_pub_key, total_amount, available_amount, proposed_roi from slots")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, validatorPubKey string
		var totalAmount, availableAmount, proposedRoi int64
		err = rows.Scan(&slotId, &validatorPubKey, &totalAmount, &availableAmount, &proposedRoi)
		if err != nil {
			panic(err)
		}

		pk, _ := GetPubKey(validatorPubKey)
		slots = append(slots, NewSlot(slotId, pk, totalAmount, availableAmount, proposedRoi))
	}

	return
}

func GetSlotDelegate(delegatorAddress string, slotId string) *SlotDelegate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select Amount from slot_delegates where slot_id = ? and delegator_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var amount int64
	err = stmt.QueryRow(slotId, delegatorAddress).Scan(&amount)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return NewSlotDelegate(delegatorAddress, slotId, amount)
}

func saveSlotDelegate(slotDelegate SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into slot_delegates(delegator_address, slot_id, Amount) values(?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.DelegatorAddress, slotDelegate.SlotId, slotDelegate.Amount)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func removeSlotDelegate(slotDelegate SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("delete from slot_delegates where delegator_address = ? and slot_id =?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.DelegatorAddress, slotDelegate.SlotId)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func saveDelegateHistory(delegateHistory DelegateHistory) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into slot_delegates(delegator_address, slot_id, Amount, op_code) values(?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(delegateHistory.DelegatorAddress, delegateHistory.SlotId, delegateHistory.Amount, delegateHistory.OpCode)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}