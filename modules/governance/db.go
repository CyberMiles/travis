package governance

import (
	"fmt"
	"strings"

	"database/sql"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/ethereum/go-ethereum/common"
)

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
		panic(err)
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
			panic(err)
		}
		wrapper.tx = tx
		wrapper.withBlock = false
	}
	return wrapper
}

func (wrapper *SqlTxWrapper) Commit() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Commit(); err != nil {
			panic(err)
		}
	}
}

func (wrapper *SqlTxWrapper) Rollback() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Rollback(); err != nil {
			panic(err)
		}
	}
}

func SaveProposal(pp *Proposal) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into governance_proposal(id, type, proposer, block_height, expire_timestamp, expire_block_height, hash) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(pp.Id, pp.Type, pp.Proposer.String(), pp.BlockHeight, pp.ExpireTimestamp, pp.ExpireBlockHeight, common.Bytes2Hex(pp.Hash()))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	switch pp.Type {
	case TRANSFER_FUND_PROPOSAL:
		stmt1, err := txWrapper.tx.Prepare("insert into governance_transfer_fund_detail(proposal_id, from_address, to_address, amount, reason) values(?, ?, ?, ?, ?)")
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
		stmt1, err := txWrapper.tx.Prepare("insert into governance_change_param_detail(proposal_id, param_name, param_value,  reason) values(?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["name"], pp.Detail["value"], pp.Detail["reason"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case DEPLOY_LIBENI_PROPOSAL:
		stmt1, err := txWrapper.tx.Prepare("insert into governance_deploy_libeni_detail(proposal_id, name, version, fileurl, md5, reason, status) values(?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["name"], pp.Detail["version"], pp.Detail["fileurl"], pp.Detail["md5"], pp.Detail["reason"], pp.Detail["status"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case RETIRE_PROGRAM_PROPOSAL:
		stmt1, err := txWrapper.tx.Prepare("insert into governance_retire_program_detail(proposal_id, retired_version, preserved_validators, reason) values(?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["retired_version"], pp.Detail["preserved_validators"], pp.Detail["reason"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	case UPGRADE_PROGRAM_PROPOSAL:
		stmt1, err := txWrapper.tx.Prepare("insert into governance_upgrade_program_detail(proposal_id, retired_version, name, version, fileurl, md5, reason) values(?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()

		_, err = stmt1.Exec(pp.Id, pp.Detail["retired_version"], pp.Detail["name"], pp.Detail["version"], pp.Detail["fileurl"], pp.Detail["md5"], pp.Detail["reason"])
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}

func GetProposalById(pid string) *Proposal {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select type, proposer, block_height, expire_timestamp, expire_block_height, hash, result, result_msg, result_block_height from governance_proposal where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var ptype, proposer, result, resultMsg, hash string
	var blockHeight, expireTimestamp, expireBlockHeight, resultBlockHeight int64
	err = stmt.QueryRow(pid).Scan(&ptype, &proposer, &blockHeight, &expireTimestamp, &expireBlockHeight, &hash, &result, &resultMsg, &resultBlockHeight)
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
		stmt1, err := txWrapper.tx.Prepare("select from_address, to_address, amount, reason from governance_transfer_fund_detail where proposal_id = ?")
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
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			map[string]interface{}{
				"from":   &fr,
				"to":     &to,
				"amount": amount,
				"reason": reason,
			},
		}
	case CHANGE_PARAM_PROPOSAL:
		var name, value, reason string
		stmt1, err := txWrapper.tx.Prepare("select param_name, param_value, reason from governance_change_param_detail where proposal_id = ?")
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
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			map[string]interface{}{
				"name":   name,
				"value":  value,
				"reason": reason,
			},
		}
	case DEPLOY_LIBENI_PROPOSAL:
		var name, version, fileurl, md5, reason, status string
		stmt1, err := txWrapper.tx.Prepare("select name, version, fileurl, md5, reason, status from governance_deploy_libeni_detail where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()
		err = stmt1.QueryRow(pid).Scan(&name, &version, &fileurl, &md5, &reason, &status)
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
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			map[string]interface{}{
				"name":    name,
				"version": version,
				"fileurl": fileurl,
				"md5":     md5,
				"reason":  reason,
				"status":  status,
			},
		}
	case RETIRE_PROGRAM_PROPOSAL:
		var retiredVersion, preservedValidators, reason string
		stmt1, err := txWrapper.tx.Prepare("select retired_version, preserved_validators, reason from governance_retire_program_detail where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()
		err = stmt1.QueryRow(pid).Scan(&retiredVersion, &preservedValidators, &reason)
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
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			map[string]interface{}{
				"retired_version": retiredVersion,
				"preserved_validators": preservedValidators,
				"reason": reason,
			},
		}
	case UPGRADE_PROGRAM_PROPOSAL:
		var retiredVersion, name, version, fileurl, md5, reason string
		stmt1, err := txWrapper.tx.Prepare("select retired_version, name, version, fileurl, md5, reason from governance_upgrade_program_detail where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt1.Close()
		err = stmt1.QueryRow(pid).Scan(&retiredVersion, &name, &version, &fileurl, &md5, &reason)
		switch {
		case err == sql.ErrNoRows:
			return nil
		case err != nil:
			panic(err)
		}

		return &Proposal {
			pid,
			ptype,
			&prp,
			blockHeight,
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			map[string]interface{}{
				"retired_version": retiredVersion,
				"name": name,
				"version": version,
				"fileurl": fileurl,
				"md5": md5,
				"reason": reason,
			},
		}
	}

	return nil
}

func UpdateProposalResult(pid, result, msg string, blockHeight int64) {
	p := GetProposalById(pid)
	if p == nil {
		return
	}

	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update governance_proposal set result = ?, result_msg = ?, result_block_height = ?, hash = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	p.Result = result
	p.ResultMsg = msg
	p.ResultBlockHeight = blockHeight

	_, err = stmt.Exec(result, msg, blockHeight, common.Bytes2Hex(p.Hash()), pid)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func UpdateDeployLibEniStatus(pid, status string) {
	go func() {
		db := getDb()
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		defer tx.Commit()

		stmt, err := tx.Prepare("update governance_deploy_libeni_detail set status = ? where proposal_id = ?")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(status, pid)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}()
}

func QueryProposals() (proposals []*Proposal) {
	tx, err := getDb().Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	proposals = getProposals(tx)
	return
}

func getProposals(tx *sql.Tx) (proposals []*Proposal) {
	rows, err := tx.Query(`select p.id, p.type, p.proposer, p.block_height, p.expire_timestamp, p.expire_block_height, p.hash, p.result, p.result_msg, p.result_block_height,
		case
		when p.type = 'transfer_fund'
		then (select printf('%s-+-%s-+-%s-+-%s', from_address, to_address, amount, reason) from governance_transfer_fund_detail where proposal_id = p.id) 
		when p.type = 'change_param'
		then (select printf('%s-+-%s-+-%s', param_name, param_value, reason) from governance_change_param_detail where proposal_id = p.id)
		when p.type = 'deploy_libeni'
		then (select printf('%s-+-%s-+-%s-+-%s-+-%s-+-%s', name, version, fileurl, md5, reason, status) from governance_deploy_libeni_detail where proposal_id = p.id)
		when p.type = 'retire_program'
		then (select printf('%s-+-%s-+-%s', retired_version, preserved_validators, reason) from governance_retire_program_detail where proposal_id = p.id)
		when p.type = 'upgrade_program'
		then (select printf('%s-+-%s-+-%s-+-%s-+-%s-+-%s', retired_version, name, version, fileurl, md5, reason) from governance_upgrade_program_detail where proposal_id = p.id)
		end as detail
		from governance_proposal p`)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, ptype, proposer, result, resultMsg, hash, detail string
		var blockHeight, expireTimestamp, expireBlockHeight, resultBlockHeight int64

		err = rows.Scan(&id, &ptype, &proposer, &blockHeight, &expireTimestamp, &expireBlockHeight, &hash, &result, &resultMsg, &resultBlockHeight, &detail)
		if err != nil {
			panic(err)
		}

		prp := common.HexToAddress(proposer)

		pp := &Proposal{
			id,
			ptype,
			&prp,
			blockHeight,
			expireTimestamp,
			expireBlockHeight,
			result,
			resultMsg,
			resultBlockHeight,
			nil,
		}

		d := strings.Split(detail, "-+-")

		switch ptype {
		case TRANSFER_FUND_PROPOSAL:
			if len(d) != 4 {
				continue
			}
			fromAddr := d[0]
			toAddr := d[1]

			fr := common.HexToAddress(fromAddr)
			to := common.HexToAddress(toAddr)

			pp.Detail = map[string]interface{}{
				"from":   &fr,
				"to":     &to,
				"amount": d[2],
				"reason": d[3],
			}
		case CHANGE_PARAM_PROPOSAL:
			if len(d) != 3 {
				continue
			}
			pp.Detail = map[string]interface{}{
				"name":   d[0],
				"value":  d[1],
				"reason": d[2],
			}
		case DEPLOY_LIBENI_PROPOSAL:
			if len(d) != 6 {
				continue
			}
			pp.Detail = map[string]interface{}{
				"name":    d[0],
				"version": d[1],
				"fileurl": d[2],
				"md5":     d[3],
				"reason":  d[4],
				"status":  d[5],
			}
		case RETIRE_PROGRAM_PROPOSAL:
			if len(d) != 3 {
				continue
			}
			pp.Detail = map[string]interface{}{
				"retired_version":      d[0],
				"preserved_validators": d[1],
				"reason":               d[2],
			}
		case UPGRADE_PROGRAM_PROPOSAL:
			if len(d) != 6 {
				continue
			}
			pp.Detail = map[string]interface{} {
				"retired_version": d[0],
				"name":            d[1],
				"version":         d[2],
				"fileurl":         d[3],
				"md5":             d[4],
				"reason":          d[5],
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

func HasUndeployedProposal(name string) bool {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select p.id from governance_proposal p, governance_deploy_libeni_detail d where p.id = d.proposal_id and p.type='deploy_libeni' and (p.result = 'Approved' or p.result = '') and (d.status != 'deployed' and d.status != 'failed' and d.status != 'collapsed')  and d.name = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(name)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		return true
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	return false
}

func GetPendingProposals() (proposals []*Proposal) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	rows, err := txWrapper.tx.Query("select id, type, expire_timestamp, expire_block_height from governance_proposal p where result = '' or (result = 'Approved' and type = 'deploy_libeni' and exists (select * from governance_deploy_libeni_detail d where d.proposal_id=p.id and (d.status != 'deployed' and d.status != 'failed' and d.status != 'collapsed')))")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, ptype string
		var expireTimestamp int64
		var expireBlockHeight int64

		err = rows.Scan(&id, &ptype, &expireTimestamp, &expireBlockHeight)
		if err != nil {
			panic(err)
		}

		pp := &Proposal{
			Id:                id,
			Type:              ptype,
			ExpireTimestamp:   expireTimestamp,
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

func GetRetiringProposal(version string) *Proposal {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select id, type, result, expire_timestamp, expire_block_height from governance_proposal p where (result = '' or result = 'Approved') and type = 'retire_program' and exists (select * from governance_retire_program_detail d where d.proposal_id=p.id and d.retired_version = ?)")
	if err != nil {
		panic(err)
	}

	rows, err := stmt.Query(version)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id, ptype, result string
		var expireTimestamp int64
		var expireBlockHeight int64

		err = rows.Scan(&id, &ptype, &result, &expireTimestamp, &expireBlockHeight)
		if err != nil {
			panic(err)
		}

		pp := &Proposal{
			Id: id,
			Type: ptype,
			Result: result,
			ExpireTimestamp: expireTimestamp,
			ExpireBlockHeight: expireBlockHeight,
		}

		return pp
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	return nil
}

func GetUpgradingProposal(version string) *Proposal {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select id, type, expire_timestamp, expire_block_height from governance_proposal p where result = 'Approved' and type = 'upgrade_program' and exists (select * from governance_upgrade_program_detail d where d.proposal_id=p.id and d.retired_version = ?)")
	if err != nil {
		panic(err)
	}

	rows, err := stmt.Query(version)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id, ptype string
		var expireTimestamp int64
		var expireBlockHeight int64

		err = rows.Scan(&id, &ptype, &expireTimestamp, &expireBlockHeight)
		if err != nil {
			panic(err)
		}

		pp := &Proposal{
			Id:                id,
			Type:              ptype,
			ExpireTimestamp:   expireTimestamp,
			ExpireBlockHeight: expireBlockHeight,
		}

		return pp
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	return nil
}

func SaveVote(vote *Vote) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into governance_vote(proposal_id, voter, block_height, answer, hash) values(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(vote.ProposalId, vote.Voter.String(), vote.BlockHeight, vote.Answer, common.Bytes2Hex(vote.Hash()))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func UpdateVote(vote *Vote) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update governance_vote set answer = ?, hash = ? where proposal_id = ? and voter = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(vote.Answer, common.Bytes2Hex(vote.Hash()), vote.ProposalId, vote.Voter.String())
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetVoteByPidAndVoter(pid string, voter string) *Vote {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select answer, block_height, hash from governance_vote where proposal_id = ? and voter = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var answer, hash string
	var blockHeight int64
	err = stmt.QueryRow(pid, voter).Scan(&answer, &blockHeight, &hash)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	}

	return &Vote{
		pid,
		common.HexToAddress(voter),
		blockHeight,
		answer,
	}
}

func GetVotesByPid(pid string) (votes []*Vote) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("select voter, answer, block_height, hash from governance_vote where proposal_id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(pid)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var voter, answer, hash string
		var blockHeight int64
		err = rows.Scan(&voter, &answer, &blockHeight, &hash)
		if err != nil {
			panic(err)
		}

		vote := &Vote{
			pid,
			common.HexToAddress(voter),
			blockHeight,
			answer,
		}

		votes = append(votes, vote)
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	return
}
