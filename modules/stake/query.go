package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
)

func QueryCandidates() (candidates Candidates) {
	db := getDb()
	cond := make(map[string]interface{})
	cond["active"] = "Y"
	return queryCandidates(db, cond)
}

func QueryCandidateByAddress(address common.Address) *Candidate {
	db := getDb()
	cond := make(map[string]interface{})
	cond["address"] = address.String()
	candidates := queryCandidates(db, cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

// todo why do we need this function?
func QueryCandidateById(id int64) *Candidate {
	db := getDb()
	cond := make(map[string]interface{})
	cond["id"] = id
	candidates := queryCandidates(db, cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func QueryDelegationsByAddress(delegatorAddress common.Address) (delegations []*Delegation) {
	db := getDb()
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	return queryDelegations(db, cond)
}

func queryCandidates(db *sql.DB, cond map[string]interface{}) (candidates Candidates) {
	clause, params := buildQueryClause(cond)
	rows, err := db.Query("select id, pub_key, address, shares, voting_power, pending_voting_power,  max_shares, comp_rate, name, website, location, profile, email, verified, active, block_height, rank, state, num_of_delegators, created_at from candidates"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

func queryDelegations(db *sql.DB, cond map[string]interface{}) (delegations []*Delegation) {
	clause, params := buildQueryClause(cond)
	rows, err := db.Query("select id, delegator_address, candidate_id, delegate_amount, award_amount, withdraw_amount, pending_withdraw_amount, slash_amount, comp_rate, voting_power, state, block_height, average_staking_date, created_at from delegations"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	delegations = composeDelegationResults(rows)
	return
}
