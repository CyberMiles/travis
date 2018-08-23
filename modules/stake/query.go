package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
)

func QueryCandidates() (candidates Candidates) {
	db := getDb()
	cond := make(map[string]interface{})
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

func QueryDelegationsByDelegator(delegatorAddress common.Address) (delegations []*Delegation) {
	db := getDb()
	cond := make(map[string]interface{})
	cond["delegator_address"] = delegatorAddress.String()
	return queryDelegations(db, cond)
}

func queryCandidates(db *sql.DB, cond map[string]interface{}) (candidates Candidates) {
	clause := composeQueryClause(cond)
	rows, err := db.Query("select pub_key, address, shares, voting_power, max_shares, comp_rate, name, website, location, profile, email, verified, active, block_height, rank, state, created_at, updated_at from candidates" + clause)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

func queryDelegations(db *sql.DB, cond map[string]interface{}) (delegations []*Delegation) {
	clause := composeQueryClause(cond)
	rows, err := db.Query("select delegator_address, pub_key, delegate_amount, award_amount, withdraw_amount, slash_amount, comp_rate, created_at, updated_at from delegations" + clause)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	delegations = composeDelegationResults(rows)
	return
}
