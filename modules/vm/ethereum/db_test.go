package ethereum

import (
	"testing"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/common"
)

var (
	account = "0xb6b29ef90120bec597939e0eda6b8a9164f75deb"
)


func TestSuicide(t *testing.T)  {
	defer func() {
		err := recover()
		assert.Nil(t, err, "expecting no panics")
	}()

	assert := assert.New(t)

	addr := common.StringToAddress(account)

	os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer db.Close()

	sqlStmt := `create table suicided_contracts(contract_address tx primary key, created_at text, updated_at text);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		assert.Fail(err.Error())
	}


	stmt, err := db.Prepare("select 1 from suicided_contracts where contract_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	stmt.Exec()
	rows, err := stmt.Query(addr.Hex())
	if err != nil {
		panic(err)
	}
	assert.False(rows.Next(), "Expect no result!")

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt2, err := tx.Prepare("insert into suicided_contracts(contract_address) values(?)")

	if err != nil {
		panic(err)
	}
	defer stmt2.Close()

	_, err = stmt2.Exec(addr.Hex())
	if err != nil {
		panic(err)
	}
	tx.Commit()


	stmt3, err := db.Prepare("select 1 from suicided_contracts where contract_address = ?")
	if err != nil {
		panic(err)
	}
	defer stmt3.Close()

	rows, err = stmt3.Query(addr.Hex())
	if err != nil {
		panic(err)
	}
	assert.True(rows.Next(), "Expect one result at least!")
}
