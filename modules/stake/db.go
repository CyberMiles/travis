package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
	"strings"
	"math/big"
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
	stmt, err := db.Prepare("select pub_key, shares, voting_power, state, created_at, updated_at from candidates where address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var pubKey, state, createdAt, updatedAt, shares, votingPower string
	err = stmt.QueryRow(address.String()).Scan(&pubKey, &shares, &votingPower, &state, &createdAt, &updatedAt)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
	s := new(big.Int)
	s.SetString(shares, 10)
	v := new(big.Int)
	v.SetString(votingPower, 10)
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: address,
		Shares:       s,
		VotingPower:  v,
		State:        state,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func GetCandidateByPubKey(pubKey string) *Candidate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select address, shares, voting_power, state, created_at, updated_at from candidates where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var address, state, createdAt, updatedAt, shares, votingPower string
	err = stmt.QueryRow(pubKey).Scan(&address, &shares, &votingPower, &state, &createdAt, &updatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	pk, _ := GetPubKey(pubKey)
	s := new(big.Int)
	s.SetString(shares, 10)
	v := new(big.Int)
	v.SetString(votingPower, 10)
	return &Candidate{
		PubKey:       pk,
		OwnerAddress: common.HexToAddress(address),
		Shares:       s,
		VotingPower:  v,
		State:        state,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
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
		var pubKey, ownerAddress, state, createdAt, updatedAt, shares, votingPower string
		err = rows.Scan(&pubKey, &ownerAddress, &shares, &votingPower, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		pk, _ := GetPubKey(pubKey)
		s := new(big.Int)
		s.SetString(shares, 10)
		v := new(big.Int)
		v.SetString(votingPower, 10)
		candidate := &Candidate{
			PubKey: pk,
			//OwnerAddress:       NewActor(bs),
			OwnerAddress: common.HexToAddress(ownerAddress),
			Shares:       s,
			VotingPower:  v,
			State:        state,
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

	stmt, err := tx.Prepare("insert into candidates(pub_key, address, shares, voting_power, state, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.PubKey.KeyString(), candidate.OwnerAddress.String(), candidate.Shares.String(), candidate.VotingPower.String(), candidate.State, candidate.CreatedAt, candidate.UpdatedAt)
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

	stmt, err := tx.Prepare("update candidates set address = ?, shares = ?, voting_power = ?, state = ?, updated_at = ? where pub_key = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(candidate.OwnerAddress.String(), candidate.Shares.String(), candidate.VotingPower.String(), candidate.State, candidate.UpdatedAt, candidate.PubKey.KeyString())
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

	_, err = stmt.Exec(strings.ToUpper(slot.Id), slot.ValidatorAddress.String(), slot.TotalAmount.String(), slot.AvailableAmount.String(), slot.ProposedRoi, slot.State, slot.CreatedAt, slot.UpdatedAt)
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

	stmt, err := tx.Prepare("update slots set validator_address = ?, total_amount = ?, available_amount = ?, proposed_roi = ?, state = ?, updated_at = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.ValidatorAddress.String(), slot.TotalAmount.String(), slot.AvailableAmount.String(), slot.ProposedRoi, slot.State, slot.UpdatedAt, strings.ToUpper(slot.Id))
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

	var validatorAddress, state, createdAt, updatedAt, totalAmount, availableAmount string
	var proposedRoi int64
	err = stmt.QueryRow(strings.ToUpper(slotId)).Scan(&validatorAddress, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	t := new(big.Int)
	t.SetString(totalAmount, 10)
	a := new(big.Int)
	a.SetString(availableAmount, 10)
	return &Slot{
		Id: 				slotId,
		ValidatorAddress: 	common.HexToAddress(validatorAddress),
		TotalAmount: 		t,
		AvailableAmount: 	a,
		ProposedRoi: 		proposedRoi,
		State:       		state,
		CreatedAt: 			createdAt,
		UpdatedAt: 			updatedAt,
	}
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
		var slotId, validatorAddress, state, createdAt, updatedAt, totalAmount, availableAmount string
		var proposedRoi int64
		err = rows.Scan(&slotId, &validatorAddress, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		t := new(big.Int)
		t.SetString(totalAmount, 10)
		a := new(big.Int)
		a.SetString(availableAmount, 10)
		slot := &Slot{
			Id:               slotId,
			ValidatorAddress: common.HexToAddress(validatorAddress),
			TotalAmount:      t,
			AvailableAmount:  a,
			ProposedRoi:      proposedRoi,
			State:            state,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
		slots = append(slots, slot)
	}

	return
}

func GetSlotsByValidator(validatorAddress common.Address) (slots []*Slot) {
	db := getDb()
	defer db.Close()
	rows, err := db.Query("select id, total_amount, available_amount, proposed_roi, state, created_at, updated_at from slots where validator_address = ?", validatorAddress.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, state, createdAt, updatedAt, totalAmount, availableAmount string
		var proposedRoi int64
		err = rows.Scan(&slotId, &totalAmount, &availableAmount, &proposedRoi, &state, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}

		t := new(big.Int)
		t.SetString(totalAmount, 10)
		a := new(big.Int)
		a.SetString(availableAmount, 10)
		slot := &Slot{
			Id:               slotId,
			ValidatorAddress: validatorAddress,
			TotalAmount:      t,
			AvailableAmount:  a,
			ProposedRoi:      proposedRoi,
			State:            state,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
		slots = append(slots, slot)
	}

	return
}

func removeSlot(slot *Slot) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("delete from slots where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.Id)
	if err != nil {
		panic(err)
	}
}

func GetSlotDelegate(delegatorAddress common.Address, slotId string) *SlotDelegate {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select amount, created_at, updated_at from slot_delegates where slot_id = ? and delegator_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var createdAt, updatedAt, amount string
	err = stmt.QueryRow(slotId, delegatorAddress.String()).Scan(&amount, &createdAt, &updatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	a := new(big.Int)
	a.SetString(amount, 10)
	return &SlotDelegate{
		DelegatorAddress: delegatorAddress,
		SlotId:           slotId,
		Amount:           a,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

func GetSlotDelegatesByAddress(delegatorAddress common.Address) (slotDelegates []*SlotDelegate) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select delegator_address, slot_id, amount, created_at, updated_at from slot_delegates where delegator_address = ?", delegatorAddress.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, delegatorAddress, createdAt, updatedAt, amount string
		err = rows.Scan(&delegatorAddress, &slotId, &amount, &createdAt, &updatedAt)

		switch {
		case err == sql.ErrNoRows:
			return
		case err != nil:
			panic(err)
		}

		a := new(big.Int)
		a.SetString(amount, 10)
		slotDelegates = append(slotDelegates,
			&SlotDelegate{
				DelegatorAddress: common.HexToAddress(delegatorAddress),
				SlotId:           slotId,
				Amount:           a,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
			})
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

	rows, err := db.Query("select delegator_address, slot_id, amount, created_at, updated_at from slot_delegates where slot_id = ?", slotId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotId, delegatorAddress, createdAt, updatedAt, amount string
		err = rows.Scan(&delegatorAddress, &slotId, &amount, &createdAt, &updatedAt)

		switch {
		case err == sql.ErrNoRows:
			return
		case err != nil:
			panic(err)
		}

		a := new(big.Int)
		a.SetString(amount, 10)
		slotDelegates = append(slotDelegates,
			&SlotDelegate{
				DelegatorAddress: common.HexToAddress(delegatorAddress),
				SlotId:           slotId,
				Amount:           a,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
			})
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func updateSlotDelegate(slotDelegate *SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update slot_delegates set amount = ?, updated_at = ? where delegator_address = ? and slot_id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.Amount.String(), slotDelegate.UpdatedAt, slotDelegate.DelegatorAddress.String(), slotDelegate.SlotId)
	if err != nil {
		panic(err)
	}
}

func saveSlotDelegate(slotDelegate *SlotDelegate) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into slot_delegates(delegator_address, slot_id, amount, created_at, updated_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slotDelegate.DelegatorAddress.String(), slotDelegate.SlotId, slotDelegate.Amount.String(), slotDelegate.CreatedAt, slotDelegate.UpdatedAt)
	if err != nil {
		panic(err)
	}
}

func removeSlotDelegate(slotDelegate *SlotDelegate) {
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

	_, err = stmt.Exec(slotDelegate.DelegatorAddress.String(), slotDelegate.SlotId)
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

	stmt, err := tx.Prepare("insert into delegate_history(delegator_address, slot_id, amount, op_code, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(delegateHistory.DelegatorAddress.String(), delegateHistory.SlotId, delegateHistory.Amount.String(), delegateHistory.OpCode, delegateHistory.CreatedAt)
	if err != nil {
		panic(err)
	}
}
