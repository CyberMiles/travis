package stake

import (
	"testing"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"github.com/stretchr/testify/assert"
)

func TestSqlite(t *testing.T) {
	assert := assert.New(t)

	os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		assert.Fail(err.Error())
	}

	tx, err := db.Begin()
	if err != nil {
		assert.Fail(err.Error())
	}
	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("Hello world %03d", i))
		if err != nil {
			assert.Fail(err.Error())
		}
	}
	tx.Commit()

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			assert.Fail(err.Error())
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		assert.Fail(err.Error())
	}

	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		assert.Fail(err.Error())
	}
	fmt.Println(name)

	_, err = db.Exec("delete from foo")
	if err != nil {
		assert.Fail(err.Error())
	}

	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		assert.Fail(err.Error())
	}

	rows, err = db.Query("select id, name from foo")
	if err != nil {
		assert.Fail(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			assert.Fail(err.Error())
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		assert.Fail(err.Error())
	}
}