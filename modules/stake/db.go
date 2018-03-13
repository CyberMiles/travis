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

func saveSlot(slot *Slot) error {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := path.Join(rootDir, "data", "stake.db")

	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into slots(id, validator_pub_key, total_amount, available_amount, proposed_roi) values(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(slot.Id, slot.ValidatorPubKey.KeyString(), slot.TotalAmount, slot.AvailableAmount, slot.ProposedRoi)
	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}