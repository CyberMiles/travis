package stake

import (
	"testing"
	"database/sql"
	"fmt"
	"time"
	"sync"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/sdk"
)

func TestLock(t *testing.T) {
	db, err := sql.Open("sqlite3", "/root/cybermiles.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var wg sync.WaitGroup

  /*
  	Run 1 routines within which a transaction(TXS) will save some data
  	and 9 routines within each a transaction(TXQ) will query some data.
  	Each query will fetch more than one row so it will last many seconds
  	because of the interval between each row scaning. #scan interval#

		There are two possible panic for 'database is locked':
			#panic (a)# TXS commit fails.
			#panic (b1, b2)# Querying data fails. Only occurs after the TXS began
				to commit and the TXQ just prepares to query for the first row.
		So it means if all 9 TXQs have fetched their first row, the only panic
		will be (a).
  */
	wg.Add(10)
	for i := int64(21); i <= 30; i++ {
		go func(t int64) {
			tx, _ := db.Begin()
			if t > 21 {
				queryDelegation(tx, t, &wg)
			} else {
				saveDelegation(tx, t, &wg)
			}
			if err := tx.Commit(); err != nil {
			  fmt.Println("commit err", t)
				panic(err) // #panic (a)#
			}
		}(i)
	}

	/*
	wg.Add(2)
	go func() {
		tx, _ := db.Begin()
		queryDelegation(tx, 26, &wg)
		if err := tx.Commit(); err != nil {
			panic(err)
		}
	}()
	go func() {
		tx, _ := db.Begin()
		saveDelegation(tx, 0, &wg)
		if err := tx.Commit(); err != nil {
			panic(err)
		}
	}()
	*/

	wg.Wait()
}

func queryDelegation(tx *sql.Tx, cid int64, wg *sync.WaitGroup) {
	defer wg.Done()

	rows, err := tx.Query(fmt.Sprintf("select id, delegator_address from delegations where candidate_id = %d", cid))
	if err != nil {
		fmt.Println("prepare query err", cid)
		panic(err) // #panic (b1)#
	}
	defer rows.Close()
	for rows.Next() {
	  var id int
		var delegator string
		err = rows.Scan(&id, &delegator)
		if err != nil {
			panic(err)
		}
		fmt.Println(cid, id, delegator)
		time.Sleep(1000 * time.Millisecond) // #scan interval#
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("fetch row err", cid)
		panic(err) // #panic (b2)#
	}
}

func saveDelegation(tx *sql.Tx, cid int64, wg *sync.WaitGroup) {
	defer wg.Done()

	d := &Delegation {
		Id:                    0,
		DelegatorAddress:      common.HexToAddress("0x1111111111111111111111111111111111111111"),
		CandidateId:           cid,
		DelegateAmount:        "0",
		AwardAmount:           "0",
		WithdrawAmount:        "0",
		PendingWithdrawAmount: "0",
		SlashAmount:           "0",
		CompRate:              sdk.NewRat(1, 5),
		VotingPower:           0,
		State:                 "N",
		BlockHeight:           0,
		AverageStakingDate:    0,
		CreatedAt:             0,
		Source:                "",
		CompletelyWithdraw:    "N",
	}

	stmt, err := tx.Prepare("insert into delegations(delegator_address, candidate_id, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, hash, voting_power, state, block_height, average_staking_date, created_at, source, completely_withdraw) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.DelegatorAddress.String(), d.CandidateId, d.DelegateAmount, d.AwardAmount, d.WithdrawAmount, d.PendingWithdrawAmount, d.SlashAmount, d.CompRate.String(), common.Bytes2Hex(d.Hash()), d.VotingPower, d.State, d.BlockHeight, d.AverageStakingDate, d.CreatedAt, d.Source, d.CompletelyWithdraw)
	if err != nil {
		panic(err)
	}
}
