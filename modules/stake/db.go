package stake

import (
	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
	"strings"
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

func GetCandidate(address string) *Candidate {
	address = strings.TrimPrefix(address, "0x")
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select pub_key, shares, voting_power, state, created_at, updated_at from candidates where address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var pubKey, state, createdAt, updatedAt string
	var shares, votingPower uint64
	err = stmt.QueryRow(strings.ToUpper(address)).Scan(&pubKey, &shares, &votingPower, &state, &createdAt, &updatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
	bs, _ := hex.DecodeString(address)
	return &Candidate{
		PubKey:      pk,
		Owner:       NewActor(bs),
		Shares:      shares,
		VotingPower: votingPower,
		State:       state,
		CreatedAt: 	 createdAt,
		UpdatedAt:   updatedAt,
	}
}

func GetCandidates() (candidates Candidates) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select pub_key, address, shares, voting_power, state, created_at, updated_at from candidates")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var pubKey, ownerAddress, state, createdAt, updatedAt string
		var shares, votingPower uint64
		err = rows.Scan(&pubKey, &ownerAddress, &shares, &votingPower, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := GetPubKey(pubKey)
		bs, _ := hex.DecodeString(ownerAddress)
		candidate := &Candidate{
			PubKey:      pk,
			Owner:       NewActor(bs),
			Shares:      shares,
			VotingPower: votingPower,
			State:       state,
			CreatedAt: 	 createdAt,
			UpdatedAt:   updatedAt,

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

	stmt, err := tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, state, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.PubKey.KeyString(), candidate.Owner.Address.String(), candidate.Shares, candidate.VotingPower, candidate.State, candidate.CreatedAt, candidate.UpdatedAt)
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

	stmt, err := tx.Prepare("update  candidates set shares = ?, voting_power = ?, state = ?, updated_at = ? where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.Shares, candidate.VotingPower, candidate.State, candidate.UpdatedAt, candidate.PubKey.KeyString())
	if err != nil {
		panic(err)
	}
}

//func removeCandidate(pubKey string) {
//	db := getDb()
//	defer db.Close()
//	tx, err := db.Begin()
//	if err != nil {
//		panic(err)
//	}
//
//	stmt, err := tx.Prepare("delete from candidates where pub_key = ?")
//	if err != nil {
//		panic(err)
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(pubKey)
//	if err != nil {
//		panic(err)
//	}
//	tx.Commit()
//}

func saveSlot(slot *Slot) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into slots(id, validator_address, total_amount, available_amount, proposed_roi, state, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(slot.Id), slot.ValidatorAddress, slot.TotalAmount, slot.AvailableAmount, slot.ProposedRoi, slot.State, slot.CreatedAt, slot.UpdatedAt)
	if err != nil {
		panic(err)
	}
}

func updateSlot(slot *Slot) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update slots set available_amount = ?, upudated_at = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.AvailableAmount, slot.UpdatedAt, strings.ToUpper(slot.Id))
	if err != nil {
		panic(err)
	}
}

func GetSlot(slotId string) *Slot {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select validator_address, total_amount, available_amount, proposed_roi, state, created_at, updated_at from slots where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var validatorAddress, state, createdAt, updatedAt string
	var totalAmount, availableAmount, proposedRoi int64
	err = stmt.QueryRow(strings.ToUpper(slotId)).Scan(&validatorAddress, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return NewSlot(slotId, validatorAddress, totalAmount, availableAmount, proposedRoi, state)
}

func GetSlots() (slots []*Slot) {
	db := getDb()
	defer db.Close()
	rows, err := db.Query("select id, validator_address, total_amount, available_amount, proposed_roi, state, created_at, updated_at from slots")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, validatorAddress, state, createdAt, updatedAt string
		var totalAmount, availableAmount, proposedRoi int64
		err = rows.Scan(&slotId, &validatorAddress, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		slot := &Slot{
			Id: 				slotId,
			ValidatorAddress: 	validatorAddress,
			TotalAmount: 		totalAmount,
			AvailableAmount: 	availableAmount,
			ProposedRoi: 		proposedRoi,
			State:       		state,
			CreatedAt: 			createdAt,
			UpdatedAt: 			updatedAt,
		}
		slots = append(slots, slot)
	}

	return
}

func GetSlotsByValidator(validatorAddress string) (slots []*Slot) {
	db := getDb()
	defer db.Close()
	rows, err := db.Query("select id, validator_address, total_amount, available_amount, proposed_roi, state, created_at, updated_at from slots where validator_address = ?", validatorAddress)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, validatorAddress, state, createdAt, updatedAt string
		var totalAmount, availableAmount, proposedRoi int64
		err = rows.Scan(&slotId, &validatorAddress, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		slot := &Slot{
			Id: 				slotId,
			ValidatorAddress: 	validatorAddress,
			TotalAmount: 		totalAmount,
			AvailableAmount: 	availableAmount,
			ProposedRoi: 		proposedRoi,
			State:       		state,
			CreatedAt: 			createdAt,
			UpdatedAt: 			updatedAt,
		}
		slots = append(slots, slot)
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

func GetSlotDelegatesByAddress(delegatorAddress string) (slotDelegates []*SlotDelegate) {
	db := getDb()
	defer db.Close()

	delegatorAddress = strings.TrimPrefix(delegatorAddress, "0x")
	rows, err := db.Query("select slot_id, amount from slot_delegates where delegator_address = ?", delegatorAddress)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId string
		var amount int64
		err = rows.Scan(&slotId, &amount)

		switch {
		case err == sql.ErrNoRows:
			return
		case err != nil:
			panic(err)
		}

		slotDelegates = append(slotDelegates, NewSlotDelegate(delegatorAddress, slotId, amount))
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func GetSlotDelegatesBySlot(slotId string) (slotDelegates []*SlotDelegate) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select slot_id, amount from slot_delegates where slot_id = ?", slotId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, delegatorAddress string
		var amount int64
		err = rows.Scan(&slotId, &amount)

		switch {
		case err == sql.ErrNoRows:
			return
		case err != nil:
			panic(err)
		}

		slotDelegates = append(slotDelegates, NewSlotDelegate(delegatorAddress, slotId, amount))
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func saveSlotDelegate(slotDelegate SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into slot_delegates(delegator_address, slot_id, amount) values(?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.DelegatorAddress, slotDelegate.SlotId, slotDelegate.Amount)
	if err != nil {
		panic(err)
	}
}

func removeSlotDelegate(slotDelegate SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from slot_delegates where delegator_address = ? and slot_id =?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.DelegatorAddress, slotDelegate.SlotId)
	if err != nil {
		panic(err)
	}
}

func saveDelegateHistory(delegateHistory DelegateHistory) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into slot_delegates(delegator_address, slot_id, Amount, op_code) values(?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(delegateHistory.DelegatorAddress, delegateHistory.SlotId, delegateHistory.Amount, delegateHistory.OpCode)
	if err != nil {
		panic(err)
	}
}