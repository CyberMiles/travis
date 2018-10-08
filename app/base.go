package app

import (
	"bytes"
	"database/sql"
	goerr "errors"
	"math/big"
	"strings"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/server"
	ttypes "github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/CyberMiles/travis/version"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/tendermint/tendermint/crypto"
	"golang.org/x/crypto/ripemd160"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	EthApp              *EthermintApplication
	checkedTx           map[common.Hash]*types.Transaction
	ethereum            *eth.Ethereum
	AbsentValidators    *stake.AbsentValidators
	ByzantineValidators []abci.Evidence
	PresentValidators   stake.Validators
	blockTime           int64
	deliverSqlTx        *sql.Tx
	proposer            abci.Validator
}

var (
	_ abci.Application = &BaseApp{}
)

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, ethApp *EthermintApplication, ethereum *eth.Ethereum) (*BaseApp, error) {
	// init pending proposals
	pendingProposals := governance.GetPendingProposals()
	if len(pendingProposals) > 0 {
		proposalsTS := make(map[string]int64)
		proposalsBH := make(map[string]int64)
		for _, pp := range pendingProposals {
			if pp.ExpireTimestamp > 0 {
				proposalsTS[pp.Id] = pp.ExpireTimestamp
			} else {
				proposalsBH[pp.Id] = pp.ExpireBlockHeight
			}

			if pp.Type == governance.DEPLOY_LIBENI_PROPOSAL {
				dp := governance.GetProposalById(pp.Id)
				if dp.Detail["status"] != "ready" {
					governance.DownloadLibEni(dp)
				}
			}
		}
		utils.PendingProposal.BatchAddTS(proposalsTS)
		utils.PendingProposal.BatchAddBH(proposalsBH)
	}

	b := store.Append().Get(utils.ParamKey)
	if b != nil {
		utils.LoadParams(b)
	}

	app := &BaseApp{
		StoreApp:  store,
		EthApp:    ethApp,
		checkedTx: make(map[common.Hash]*types.Transaction),
		ethereum:  ethereum,
	}
	return app, nil
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	return
}

// Info implements abci.Application. It returns the height and hash,
// as well as the abci name and version.
//
// The height is the block that holds the transactions, not the apphash itself.
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	ethInfoRes := app.EthApp.Info(req)

	if big.NewInt(ethInfoRes.LastBlockHeight).Cmp(bigZero) == 0 {
		return ethInfoRes
	}

	rp := governance.GetRetiringProposal(version.Version)
	if rp != nil && rp.Result == "Approved" {
		if rp.ExpireBlockHeight <= ethInfoRes.LastBlockHeight {
			server.StopFlag <- true
		} else if rp.ExpireBlockHeight == ethInfoRes.LastBlockHeight+1 {
			utils.RetiringProposalId = rp.Id
		} else {
			// check ahead one block
			utils.PendingProposal.Add(rp.Id, 0, rp.ExpireBlockHeight-1)
		}
	}

	travisInfoRes := app.StoreApp.Info(req)

	travisInfoRes.LastBlockAppHash = finalAppHash(ethInfoRes.LastBlockAppHash, travisInfoRes.LastBlockAppHash, app.StoreApp.GetDbHash(), travisInfoRes.LastBlockHeight, nil)
	return travisInfoRes
}

// DeliverTx - ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("DeliverTx: Received invalid transaction", "err", err)
		return errors.DeliverResult(err)
	}

	if utils.IsEthTx(tx) {
		if checkedTx, ok := app.checkedTx[tx.Hash()]; ok {
			tx = checkedTx
		} else {
			// force cache from of tx
			networkId := big.NewInt(int64(app.ethereum.NetVersion()))
			signer := types.NewEIP155Signer(networkId)

			if _, err := types.Sender(signer, tx); err != nil {
				app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)
				return errors.DeliverResult(err)
			}
		}
		resp := app.EthApp.DeliverTx(tx)
		app.logger.Debug("EthApp DeliverTx response", "resp", resp)
		return resp
	}

	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.blockTime, app.EthApp.DeliverTxState())
	return app.deliverHandler(ctx, app.Append(), tx)
}

// CheckTx - ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("CheckTx: Received invalid transaction", "err", err)
		return errors.CheckResult(err)
	}

	if utils.IsEthTx(tx) {
		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("EthApp CheckTx response", "resp", resp)
		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.String()))
		}
		app.checkedTx[tx.Hash()] = tx
		return sdk.NewCheck(0, "").ToABCI()
	}

	app.logger.Info("CheckTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.blockTime, app.EthApp.checkTxState)
	return app.checkHandler(ctx, app.Check(), tx)
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.blockTime = req.GetHeader().Time
	app.EthApp.BeginBlock(req)
	app.PresentValidators = app.PresentValidators[:0]
	app.AbsentValidators = stake.LoadAbsentValidators(app.Append())

	// init deliver sql tx for statke
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		panic(err)
	}
	deliverSqlTx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	app.deliverSqlTx = deliverSqlTx
	stake.SetDeliverSqlTx(deliverSqlTx)
	governance.SetDeliverSqlTx(deliverSqlTx)
	// init end

	// handle the absent validators
	for _, sv := range req.Validators {
		var pk crypto.PubKeyEd25519
		copy(pk[:], sv.Validator.PubKey.Data)

		pubKey := ttypes.PubKey{pk}
		if !sv.SignedLastBlock {
			app.AbsentValidators.Add(pubKey, app.WorkingHeight())
		} else {
			v := stake.GetCandidateByPubKey(pubKey)
			if v != nil {
				app.PresentValidators = append(app.PresentValidators, v.Validator())
			}
		}
	}

	app.AbsentValidators.Clear(app.WorkingHeight())
	stake.SaveAbsentValidators(app.Append(), app.AbsentValidators)

	app.logger.Info("BeginBlock", "absent_validators", app.AbsentValidators)
	app.ByzantineValidators = req.ByzantineValidators
	app.proposer = req.Header.Proposer

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.EthApp.EndBlock(req)
	utils.BlockGasFee = big.NewInt(0).Add(utils.BlockGasFee, app.TotalUsedGasFee)

	// slash Byzantine validators
	if len(app.ByzantineValidators) > 0 {
		for _, bv := range app.ByzantineValidators {
			pk, err := ttypes.GetPubKey(string(bv.Validator.PubKey.Data))
			if err != nil {
				continue
			}

			stake.SlashByzantineValidator(pk, app.blockTime, app.WorkingHeight())
		}
		app.ByzantineValidators = app.ByzantineValidators[:0]
	}

	// slash the absent validators
	for k, v := range app.AbsentValidators.Validators {
		pk, err := ttypes.GetPubKey(k)
		if err != nil {
			continue
		}

		stake.SlashAbsentValidator(pk, v, app.blockTime, app.WorkingHeight())
	}

	var backups stake.Validators
	for _, bv := range stake.GetBackupValidators() {
		// exclude the absent validators
		if !app.AbsentValidators.Contains(bv.PubKey) {
			backups = append(backups, bv.Validator())
		}
	}

	// Deactivate validators that not in the list of preserved validators
	if utils.RetiringProposalId != "" {
		if proposal := governance.GetProposalById(utils.RetiringProposalId); proposal != nil {
			pks := strings.Split(proposal.Detail["preserved_validators"].(string), ",")
			vs := stake.GetCandidates().Validators()
			inaVs := make(stake.Validators, 0)
			abciVs := make([]abci.Validator, 0)
			for _, v := range vs {
				i := 0
				for ; i < len(pks); i++ {
					if pks[i] == ttypes.PubKeyString(v.PubKey) {
						abciVs = append(abciVs, v.ABCIValidator())
						break
					}
				}
				if i == len(pks) {
					inaVs = append(inaVs, v)
					pk := v.PubKey.PubKey.(crypto.PubKeyEd25519)
					abciVs = append(abciVs, abci.Ed25519Validator(pk[:], 0))
				}
			}
			inaVs.Deactivate()
			app.AddValChange(abciVs)
		} else {
			app.logger.Error("Getting invalid RetiringProposalId")
		}
	} else { // should not update validator set twice if the node is to be shutdown
		// calculate the validator set difference
		if calVPCheck(app.WorkingHeight()) {
			diff, err := stake.UpdateValidatorSet(app.WorkingHeight())
			if err != nil {
				panic(err)
			}
			app.AddValChange(diff)
		}
	}

	// block award
	// run once per hour
	if len(app.PresentValidators) > 0 {
		stake.NewAwardDistributor(app.Append(), app.WorkingHeight(), app.PresentValidators, backups, app.logger).Distribute()
	}
	// block award end

	// handle the pending unstake requests
	stake.HandlePendingUnstakeRequests(app.WorkingHeight())

	// record candidates stakes daily
	if calStakeCheck(app.WorkingHeight()) {
		// run once a day
		stake.RecordCandidateDailyStakes(app.WorkingHeight())
	}

	// Accumulates the average staking date of all delegations
	if calAvgStakingDateCheck(app.WorkingHeight()) {
		// run once a day
		stake.AccumulateDelegationsAverageStakingDate()
	}

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	if utils.RetiringProposalId != "" {
		server.StopFlag <- true
	}

	app.checkedTx = make(map[common.Hash]*types.Transaction)
	ethAppCommit, err := app.EthApp.Commit()
	if err != nil {
		// Rollback transaction
		if app.deliverSqlTx != nil {
			err := app.deliverSqlTx.Rollback()
			if err != nil {
				panic(err)
			}
			stake.ResetDeliverSqlTx()
			governance.ResetDeliverSqlTx()
		}

		// slash block proposer
		var pk crypto.PubKeyEd25519
		copy(pk[:], app.proposer.PubKey.Data)
		pubKey := ttypes.PubKey{pk}
		stake.SlashBadProposer(pubKey, app.blockTime, app.WorkingHeight())
	} else {
		if app.deliverSqlTx != nil {
			// Commit transaction
			err := app.deliverSqlTx.Commit()
			if err != nil {
				panic(err)
			}
			stake.ResetDeliverSqlTx()
			governance.ResetDeliverSqlTx()
		}
	}

	workingHeight := app.WorkingHeight()

	if dirty := utils.CleanParams(); workingHeight == 1 || dirty {
		state := app.Append()
		state.Set(utils.ParamKey, utils.UnloadParams())
	}

	// reset store app
	app.TotalUsedGasFee = big.NewInt(0)

	res = app.StoreApp.Commit()
	dbHash := app.StoreApp.GetDbHash()
	res.Data = finalAppHash(ethAppCommit.Data, res.Data, dbHash, workingHeight, nil)

	return
}

func finalAppHash(ethCommitHash []byte, travisCommitHash []byte, dbHash []byte, workingHeight int64, store *state.SimpleDB) []byte {

	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	buf.Write(ethCommitHash)
	buf.Write(travisCommitHash)
	buf.Write(dbHash)
	hasher.Write(buf.Bytes())
	hash := hasher.Sum(nil)
	return hash
}

func calStakeCheck(height int64) bool {
	return height%int64(utils.GetParams().CalStakeInterval) == 0
}

func calVPCheck(height int64) bool {
	return height == 1 || height%int64(utils.GetParams().CalVPInterval) == 0
}

func calAvgStakingDateCheck(height int64) bool {
	return height%int64(utils.GetParams().CalAverageStakingDateInterval) == 0
}
