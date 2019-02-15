package schedule_tx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	um "github.com/ethereum/go-ethereum/core/vm/umbrella"
	"golang.org/x/crypto/ripemd160"
)

func ScheduleTxHash(stx *um.ScheduleTx, parentHash string, status string) string {
	sj, err := json.Marshal(struct {
		ParentHash        string
		Sender            *common.Address
		Receiver          *common.Address
		TxData            []byte
		DueTime           uint64
		Status            string
	}{
		parentHash,
		&stx.Sender,
		&stx.Receiver,
		stx.TxData,
		stx.Unixtime,
		status,
	})
	if err != nil {
		panic(err)
	}
	hasher := ripemd160.New()
	hasher.Write(sj)
	return common.Bytes2Hex(hasher.Sum(nil))
}

var (
	deliverSqlTx *sql.Tx
)

func SetDeliverSqlTx(tx *sql.Tx) {
	deliverSqlTx = tx
}

func ResetDeliverSqlTx() {
	deliverSqlTx = nil
}

func getDb() *sql.DB {
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		log.Panic(err)
	}
	return db
}

type SqlTxWrapper struct {
	tx        *sql.Tx
	withBlock bool
}

func getSqlTxWrapper() *SqlTxWrapper {
	var wrapper = &SqlTxWrapper{
		tx:        deliverSqlTx,
		withBlock: true,
	}
	if wrapper.tx == nil {
		db := getDb()
		tx, err := db.Begin()
		if err != nil {
			log.Panic(err)
		}
		wrapper.tx = tx
		wrapper.withBlock = false
	}
	return wrapper
}

func (wrapper *SqlTxWrapper) Commit() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Commit(); err != nil {
			log.Panic(err)
		}
	}
}

func (wrapper *SqlTxWrapper) Rollback() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Rollback(); err != nil {
			log.Panic(err)
		}
	}
}

func SaveScheduleTx(parentHash string, stx *um.ScheduleTx) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into schedule_tx(parent_hash, from_address, to_address, data, time, hash) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(parentHash, stx.Sender.String(), stx.Receiver.String(), hexutil.Encode(stx.TxData), stx.Unixtime, ScheduleTxHash(stx, parentHash, ""))
	if err != nil {
		fmt.Println(err)
		log.Panic(err)
	}
}

func UpdateScheduleTxStatus(parentHash string, stx *um.ScheduleTx, status string) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update schedule_tx set status = ?, hash = ? where parent_hash = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, ScheduleTxHash(stx, parentHash, status), parentHash)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetScheduleTxByHash(parentHash string) (stx *um.ScheduleTx, status string) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select from_address, to_address, data, time, status from schedule_tx where parent_hash = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var sender, receiver, data string
	var time uint64
	err = stmt.QueryRow(parentHash).Scan(&sender, &receiver, &data, &time, &status)
	switch {
	case err == sql.ErrNoRows:
		return
	case err != nil:
		panic(err)
	}

	txData, _ := hexutil.Decode(data)
	stx = &um.ScheduleTx{
		common.HexToAddress(sender),
		common.HexToAddress(receiver),
		txData,
		time,
	}

	return
}

func GetUpcomingTsHash() (minTs int64, tsHash map[int64][]string) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select min(time) from schedule_tx where status = ''")
	if err != nil {
		log.Panic(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow().Scan(&minTs)
	switch {
	case err == sql.ErrNoRows:
		return
	case err != nil:
		log.Panic(err)
	}

	stmt1, err := txWrapper.tx.Prepare("select parent_hash, time from schedule_tx where status = '' and time <= ? order by time")
	if err != nil {
		log.Panic(err)
	}
	defer stmt1.Close()

	rows, err := stmt1.Query(minTs + utils.CommitSeconds)
	switch {
	case err == sql.ErrNoRows:
		return
	case err != nil:
		log.Panic(err)
	}

	var lastTs, ts int64
	var hash string
	for rows.Next() {
		err = rows.Scan(&hash, &ts)
		if err != nil {
			log.Panic(err)
		}
		if lastTs == ts {
			ha := tsHash[ts]
			ha = append(ha, hash)
		} else {
			tsHash[ts] = []string{hash}
			lastTs = ts
		}
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}
