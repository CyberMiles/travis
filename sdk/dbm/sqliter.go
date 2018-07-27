package dbm


import (
	"os"
	"github.com/CyberMiles/travis/sdk/errors"

	_ "github.com/mattn/go-sqlite3"
	"database/sql"
)

// SQLite helper functions
type Sqliter struct{
	dbPath string
}

func NewSqliter(dbPath string) *Sqliter {
	if err := InitSqliteDB(dbPath);  err != nil {
		panic(err)
	}
	return &Sqliter{
		dbPath: dbPath,
	}
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (sqliter *Sqliter) Exec(query string, args ...interface{}) (sql.Result, error) {
	db, err := sql.Open("sqlite3", sqliter.dbPath)
	if err != nil {
		return nil, errors.ErrInternal("Initializing stake database: " + err.Error())
	}
	defer db.Close()
	return db.Exec(query, args...)
}

func (sqliter *Sqliter) GetDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", sqliter.dbPath)
	if err != nil {
		return nil, errors.ErrInternal("Initializing stake database: " + err.Error())
	}
	return db, nil
}

func (sqliter *Sqliter) CloseDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}


func InitSqliteDB(dbPath string)  error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// dbPath does not exist
		if db, err := sql.Open("sqlite3", dbPath); err != nil {
			return err
		} else {
			db.Close()
		}

	}
	return nil
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (sqliter *Sqliter) ExecBatch(query string, args ...interface{}) (error) {
	return nil
}