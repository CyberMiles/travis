package governance

import (
	"fmt"

	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
)

func getDb() *sql.DB {
	rootDir := viper.GetString(cli.HomeFlag)
	dbPath := path.Join(rootDir, "data", "travis.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func SaveProposal(pp *Proposal) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into governance_proposal(id, proposer, block_height, from_address, to_address, amount, reason, created_at) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(pp.Id, pp.Proposer.String(), pp.BlockHeight, pp.From.String(), pp.To.String(), pp.Amount.String(), pp.Reason, pp.CreatedAt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
