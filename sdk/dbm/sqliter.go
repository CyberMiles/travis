package dbm


import (
	"os"
	"github.com/CyberMiles/travis/sdk/errors"

	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"strings"
	"github.com/ethereum/go-ethereum/log"
)

// SQLite helper functions
type sqliter struct{
	dbPath string
	db *sql.DB
	immuDb *sql.DB
}

var Sqliter = &sqliter{}

func InitSqliter(dbPath string) error {
	if strings.Compare(dbPath, "") != 0 {
		errors.ErrInternal("The sqlite database has been initialized.")
	}
	if err := initSqliteDB(dbPath);  err != nil {
		return err
	}
	Sqliter.dbPath = dbPath
	return nil
}

func (s *sqliter) GetDB() (*sql.DB, error) {
	if strings.Compare(s.dbPath, "")  == 0 {
		return nil, errors.ErrInternal("Sqlite database path is not set.")
	}
	if s.db != nil {
		if s.db.Stats().OpenConnections > 0 {
			return s.db, nil
		}
	}
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return nil, errors.ErrInternal("Open database: " + err.Error())
	}
	if db.Ping(); err != nil {
		return nil, errors.ErrInternal("Open database: " + err.Error())
	}
	s.db = db
	return db, nil
}

func (s *sqliter) GetImmuDB() (*sql.DB, error) {
	if strings.Compare(s.dbPath, "")  == 0 {
		return nil, errors.ErrInternal("Sqlite database path is not set.")
	}
	if s.immuDb != nil {
		if s.immuDb.Stats().OpenConnections > 0 {
			return s.immuDb, nil
		}
	}
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return nil, errors.ErrInternal("Open database: " + err.Error())
	}
	if db.Ping(); err != nil {
		return nil, errors.ErrInternal("Open database: " + err.Error())
	}
	db.Exec("PRAGMA journal_mode=WAL;")
	s.immuDb = db
	return db, nil
}

func (s *sqliter) CloseDB() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Warn("Failed to close sqlite db connection: " + s.dbPath)
		}
		s.db = nil
	}
	if s.immuDb != nil {
		if err := s.immuDb.Close(); err != nil {
			log.Warn("Failed to close sqlite immutable connection: " + s.dbPath)
		}
		s.immuDb = nil
	}
	log.Info("Sqlite db closed successfully！")
}


func initSqliteDB(dbPath string) error {
	// dbPath does not exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if _, err := os.Create(dbPath); err != nil {
			return err
		}
	}
	return nil
}

