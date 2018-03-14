package stake

import (
	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
)

// loadCandidate - loads the candidate object for the provided pubkey
//func getCandidate(pubKey crypto.PubKey) *Candidate {
//	if pubKey.Empty() {
//		return nil
//	}
//
//	db, err := sql.Open("sqlite3", "./stake.db")
//	if err != nil {
//		panic(err)
//	}
//	defer db.Close()
//
//	stmt, err := db.Prepare("select * from validators where pub_key = ?")
//	if err != nil {
//		panic(err)
//	}
//	defer stmt.Close()
//	var name string
//	err = stmt.QueryRow("3").Scan(&name)
//	if err != nil {
//		panic(err)
//	}
//
//	return nil
//}

func getDb() *sql.DB {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := path.Join(rootDir, "data", "stake.db")

	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	return db
}

func saveSlot(slot *Slot) {
	db := getDb()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into slots(id, validator_pub_key, total_amount, available_amount, proposed_roi) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.Id, slot.ValidatorPubKey.KeyString(), slot.TotalAmount, slot.AvailableAmount, slot.ProposedRoi)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

func getSlot(slotId string) *Slot {
	db := getDb()
	stmt, err := db.Prepare("select * from slots where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var validatorPubKey string
	var totalAmount, availableAmount, proposedRoi int64
	err = stmt.QueryRow(slotId).Scan(&validatorPubKey, &totalAmount, &availableAmount, &proposedRoi)
	if err != nil {
		panic(err)
	}

	pk, _ := GetPubKey(validatorPubKey)
	return NewSlot(slotId, pk, totalAmount, availableAmount, proposedRoi)
}

func getSlotDelegates(delegatorAddress string, slotId string) *SlotDelegate {
	db := getDb()
	stmt, err := db.Prepare("select Amount from slot_delegates where slot_id = ? and delegator_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var amount int64
	err = stmt.QueryRow(slotId, delegatorAddress).Scan(&amount)
	if err != nil {
		//switch err.(type) {
		//case sql.ErrNoRows:
		//	return &SlotDelegate{}
		//default:
		//	panic(err)
		//}
		panic(err)
	}

	return NewSlotDelegate(delegatorAddress, slotId, amount)
}

func saveSlotDelegate(slotDelegate SlotDelegate) {
	db := getDb()

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