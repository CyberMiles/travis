package governance

import (
	"fmt"

	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"path"
	"github.com/ethereum/go-ethereum/common"
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

	stmt, err := tx.Prepare("insert into governance_proposal(id, proposer, block_height, from_address, to_address, amount, reason, expire_block_height, created_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(pp.Id, pp.Proposer.String(), pp.BlockHeight, pp.From.String(), pp.To.String(), pp.Amount, pp.Reason, pp.ExpireBlockHeight, pp.CreatedAt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetProposalById(pid string) *Proposal {
	db := getDb()
	defer db.Close()

	stmt, err := db.Prepare("select proposer, block_height, from_address, to_address, amount, reason, expire_block_height, hash, created_at, result, result_msg, result_block_height, result_at from governance_proposal where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var proposer, fromAddr, toAddr, amount, reason, createdAt, result, resultMsg, resultAt, hash string
	var blockHeight, expireBlockHeight, resultBlockHeight uint64
	err = stmt.QueryRow(pid).Scan(&proposer, &blockHeight, &fromAddr, &toAddr, &amount, &reason, &expireBlockHeight, &hash, &createdAt, &result, &resultMsg, &resultBlockHeight, &resultAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	prp := common.HexToAddress(proposer)
	fr := common.HexToAddress(fromAddr)
	to := common.HexToAddress(toAddr)

	return &Proposal{
		pid,
		&prp,
		blockHeight,
		&fr,
		&to,
		amount,
		reason,
		expireBlockHeight,
		hash,
		createdAt,
		result,
		resultMsg,
		resultBlockHeight,
		resultAt,
	}
}

func UpdateProposalResult(pid, result, msg string, blockHeight uint64, resultAt string) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update governance_proposal set result = ?, result_msg = ?, result_block_height = ?, result_at = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(result, msg, blockHeight, resultAt, pid)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetProposals() (proposals []*Proposal) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select id, proposer, block_height, from_address, to_address, amount, reason, expire_block_height, hash, created_at, result, result_msg, result_block_height, result_at from governance_proposal")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, proposer, fromAddr, toAddr, amount, reason, createdAt, result, resultMsg, resultAt, hash string
		var blockHeight, expireBlockHeight, resultBlockHeight uint64

		err = rows.Scan(&id, &proposer, &blockHeight, &fromAddr, &toAddr, &amount, &reason, &expireBlockHeight, &hash, &createdAt, &result, &resultMsg, &resultBlockHeight, &resultAt)
		if err != nil {
			panic(err)
		}

		prp := common.HexToAddress(proposer)
		fr := common.HexToAddress(fromAddr)
		to := common.HexToAddress(toAddr)

		pp := &Proposal{
			id,
			&prp,
			blockHeight,
			&fr,
			&to,
			amount,
			reason,
			expireBlockHeight,
			hash,
			createdAt,
			result,
			resultMsg,
			resultBlockHeight,
			resultAt,
		}

		proposals = append(proposals, pp)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}

func GetPendingProposals() (proposals []*Proposal) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query("select id, expire_block_height from governance_proposal where result = ''")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var expireBlockHeight uint64

		err = rows.Scan(&id, &expireBlockHeight)
		if err != nil {
			panic(err)
		}

		pp := &Proposal{
			Id: id,
			ExpireBlockHeight: expireBlockHeight,
		}

		proposals = append(proposals, pp)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}


func SaveVote(vote *Vote) {
	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into governance_vote(proposal_id, voter, block_height, answer, created_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(vote.ProposalId, vote.Voter.String(), vote.BlockHeight, vote.Answer, vote.CreatedAt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetVoteByPidAndVoter(pid string, voter string) *Vote {
	db := getDb()
	defer db.Close()

	stmt, err := db.Prepare("select answer, block_height, hash, created_at from governance_vote where proposal_id = ? and voter = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var answer, createdAt, hash string
	var blockHeight uint64
	err = stmt.QueryRow(pid, voter).Scan(&answer, &blockHeight, &hash, &createdAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return &Vote {
		pid,
		common.HexToAddress(voter),
		blockHeight,
		answer,
		hash,
		createdAt,
	}
}

func GetVotesByPid(pid string) (votes []*Vote) {
	db := getDb()
	defer db.Close()

	stmt, err := db.Prepare("select voter, answer, block_height, hash, created_at from governance_vote where proposal_id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(pid)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var voter, answer, createdAt, hash string
		var blockHeight uint64
		err = rows.Scan(&voter, &answer, &blockHeight, &hash, &createdAt)
		if err != nil {
			panic(err)
		}

		vote := &Vote {
			pid,
			common.HexToAddress(voter),
			blockHeight,
			answer,
			hash,
			createdAt,
		}

		votes = append(votes, vote)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return
}
