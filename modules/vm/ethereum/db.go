package ethereum

import (
	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
	"github.com/ethereum/go-ethereum/common"
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

func HasSuicided(addr common.Address) bool {
	db := getDb()
	defer db.Close()
	stmt, err := db.Prepare("select 1 from suicided_contracts where contract_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(addr.Hex()).Scan()
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		panic(err)
	}
	return true
}

func Suicide(addr common.Address) {
	if HasSuicided(addr) {
		return
	}
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("insert into suicided_contracts(contract_address) values(?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(addr.Hex())
	if err != nil {
		panic(err)
	}
	tx.Commit()
}
