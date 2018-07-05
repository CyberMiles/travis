package governance

import (
	"fmt"
	"strings"

	"database/sql"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
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

	stmt, err := tx.Prepare("insert into governance_proposal(id, type, proposer, block_height, expire_block_height, hash, created_at) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(pp.Id, pp.Type, pp.Proposer.String(), pp.BlockHeight, pp.ExpireBlockHeight, common.Bytes2Hex(pp.Hash()), pp.CreatedAt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	switch pp.Type {
	case TRANSFER_FUND_PROPOSAL:
		stmt1, err := tx.Prepare("insert into governance_transfer_fund_detail(proposal_id, from_address, to_address, amount, reason) values(?, ?, ?, ?, ?)") 
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["from"].(*common.Address).String(), pp.Detail["to"].(*common.Address).String(), pp.Detail["amount"], pp.Detail["reason"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case CHANGE_PARAM_PROPOSAL:
		stmt1, err := tx.Prepare("insert into governance_change_param_detail(proposal_id, param_name, param_value,  reason) values(?, ?, ?, ?)") 
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["name"], pp.Detail["value"], pp.Detail["reason"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}

func GetProposalById(pid string) *Proposal {
	db := getDb()
	defer db.Close()

	stmt, err := db.Prepare("select type, proposer, block_height, expire_block_height, hash, created_at, result, result_msg, result_block_height, result_at from governance_proposal where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var ptype, proposer, createdAt, result, resultMsg, resultAt, hash string
	var blockHeight, expireBlockHeight, resultBlockHeight uint64
	err = stmt.QueryRow(pid).Scan(&ptype, &proposer, &blockHeight, &expireBlockHeight, &hash, &createdAt, &result, &resultMsg, &resultBlockHeight, &resultAt)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	prp := common.HexToAddress(proposer)

	switch ptype {
	case TRANSFER_FUND_PROPOSAL:
		var fromAddr, toAddr, amount, reason string 
		stmt1, err := db.Prepare("select from_address, to_address, amount, reason from governance_transfer_fund_detail where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()
		err = stmt1.QueryRow(pid).Scan(&fromAddr, &toAddr, &amount, &reason)
		switch {
		case err == sql.ErrNoRows:
			return nil
		case err != nil:
			panic(err)
		}

		fr := common.HexToAddress(fromAddr)
		to := common.HexToAddress(toAddr)

		return &Proposal{
			pid,
			ptype,
			&prp,
			blockHeight,
			expireBlockHeight,
			createdAt,
			result,
			resultMsg,
			resultBlockHeight,
			resultAt,
			map[string]interface{}{
				"from": &fr,
				"to": &to,
				"amount": amount,
				"reason": reason,
			},
		}
	case CHANGE_PARAM_PROPOSAL:
		var name, value, reason string
		stmt1, err := db.Prepare("select param_name, param_value, reason from governance_change_param_detail where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()
		err = stmt1.QueryRow(pid).Scan(&name, &value, &reason)
		switch {
		case err == sql.ErrNoRows:
			return nil
		case err != nil:
			panic(err)
		}

		return &Proposal{
			pid,
			ptype,
			&prp,
			blockHeight,
			expireBlockHeight,
			createdAt,
			result,
			resultMsg,
			resultBlockHeight,
			resultAt,
			map[string]interface{}{
				"name": name,
				"value": value,
				"reason": reason,
			},
		}
	}

	return nil
}

func UpdateProposalResult(pid, result, msg string, blockHeight uint64, resultAt string) {
	p := GetProposalById(pid)
	if p == nil {
		return
	}

	db := getDb()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("update governance_proposal set result = ?, result_msg = ?, result_block_height = ?, result_at = ?, hash = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	p.Result = result
	p.ResultMsg = msg
	p.ResultBlockHeight = blockHeight
	p.ResultAt = resultAt

	_, err = stmt.Exec(result, msg, blockHeight, resultAt, common.Bytes2Hex(p.Hash()), pid)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetProposals() (proposals []*Proposal) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query(`select p.id, p.type, p.proposer, p.block_height, p.expire_block_height, p.hash, p.created_at, p.result, p.result_msg, p.result_block_height, p.result_at,
		case
		when p.type = 'transfer_fund'
		then (select printf('%s-+-%s-+-%s-+-%s', from_address, to_address, amount, reason) from governance_transfer_fund_detail where proposal_id = p.id) 
		when p.type = 'change_param'
		then (select printf('%s-+-%s-+-%s', param_name, param_value, reason) from governance_change_param_detail where proposal_id = p.id)
		end as detail
		from governance_proposal p`)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, ptype, proposer, createdAt, result, resultMsg, resultAt, hash, detail string
		var blockHeight, expireBlockHeight, resultBlockHeight uint64

		err = rows.Scan(&id, &ptype, &proposer, &blockHeight, &expireBlockHeight, &hash, &createdAt, &result, &resultMsg, &resultBlockHeight, &resultAt, &detail)
		if err != nil {
			panic(err)
		}

		prp := common.HexToAddress(proposer)

		pp := &Proposal{
			id,
			ptype,
			&prp,
			blockHeight,
			expireBlockHeight,
			createdAt,
			result,
			resultMsg,
			resultBlockHeight,
			resultAt,
			nil,
		}

		d := strings.Split(detail, "-+-")

		switch ptype {
		case TRANSFER_FUND_PROPOSAL:
			fromAddr := d[0]
			toAddr := d[1]

			fr := common.HexToAddress(fromAddr)
			to := common.HexToAddress(toAddr)

			pp.Detail = map[string]interface{} {
				"from": &fr,
				"to": &to,
				"amount": d[2],
				"reason": d[3],
			}
		case CHANGE_PARAM_PROPOSAL:
			pp.Detail = map[string]interface{} {
				"name": d[0],
				"value": d[1],
				"reason": d[2],
			}
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

	stmt, err := tx.Prepare("insert into governance_vote(proposal_id, voter, block_height, answer, hash, created_at) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(vote.ProposalId, vote.Voter.String(), vote.BlockHeight, vote.Answer, common.Bytes2Hex(vote.Hash()), vote.CreatedAt)
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
