package app

import (
	goerr "errors"
	"math/big"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	abci "github.com/tendermint/tendermint/abci/types"

	"bytes"
	"database/sql"
	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
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
}

const (
	BLOCK_AWARD_STR = "10000000000000000000000"
)

var (
	blockAward, _                  = big.NewInt(0).SetString(BLOCK_AWARD_STR, 10)
	_             abci.Application = &BaseApp{}
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
		StoreApp:         store,
		EthApp:           ethApp,
		checkedTx:        make(map[common.Hash]*types.Transaction),
		ethereum:         ethereum,
		AbsentValidators: stake.NewAbsentValidators(),
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
	app.BlockEnd = false
	app.blockTime = req.GetHeader().Time
	app.EthApp.BeginBlock(req)
	app.PresentValidators = app.PresentValidators[:0]

	// init deliver sql tx for statke
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		// TODO: wrapper error
		panic(err)
	}
	deliverSqlTx, err := db.Begin()
	if err != nil {
		// TODO: wrapper error
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

	app.logger.Info("BeginBlock", "absent_validators", app.AbsentValidators)
	app.ByzantineValidators = req.ByzantineValidators

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.EthApp.EndBlock(req)
	utils.BlockGasFee = big.NewInt(0).Add(utils.BlockGasFee, app.TotalUsedGasFee)

	// punish Byzantine validators
	if len(app.ByzantineValidators) > 0 {
		for _, bv := range app.ByzantineValidators {
			pk, err := ttypes.GetPubKey(string(bv.Validator.PubKey.Data))
			if err != nil {
				continue
			}

			stake.PunishByzantineValidator(pk)
		}
		app.ByzantineValidators = app.ByzantineValidators[:0]
	}

	// punish the absent validators
	for k, v := range app.AbsentValidators.Validators {
		stake.PunishAbsentValidator(k, v)
	}

	var backups stake.Validators
	for _, bv := range stake.GetBackupValidators() {
		// exclude the absent validators
		if !app.AbsentValidators.Contains(bv.PubKey) {
			backups = append(backups, bv.Validator())
		}
	}

	// block award
	if app.WorkingHeight()%utils.BlocksPerHour == 0 {
		// calculate the validator set difference
		diff, err := stake.UpdateValidatorSet()
		if err != nil {
			panic(err)
		}
		app.AddValChange(diff)

		// run once per hour
		if len(app.PresentValidators) > 0 {
			stake.NewAwardDistributor(app.WorkingHeight(), app.PresentValidators, backups, app.logger).Distribute()
		}
	}

	// handle the pending unstake requests
	stake.HandlePendingUnstakeRequests(app.WorkingHeight(), app.Append())

	// record candidates stakes daily
	if app.WorkingHeight()%utils.BlocksPerDay == 0 {
		// run once per day
		stake.RecordCandidateDailyStakes()
	}

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.checkedTx = make(map[common.Hash]*types.Transaction)
	ethAppCommit := app.EthApp.Commit()
	if len(ethAppCommit.Data) == 0 {
		// Rollback transaction
		if app.deliverSqlTx != nil {
			err := app.deliverSqlTx.Rollback()
			if err != nil {
				// TODO: wrapper error
				panic(err)
			}
			stake.ResetDeliverSqlTx()
			governance.ResetDeliverSqlTx()
		}
		return abci.ResponseCommit{}
	}
	if dirty := utils.CleanParams(); dirty {
		state := app.Append()
		state.Set(utils.ParamKey, utils.UnloadParams())
	}

	workingHeight := app.WorkingHeight()

	// reset store app
	app.TotalUsedGasFee = big.NewInt(0)

	res = app.StoreApp.Commit()
	if app.deliverSqlTx != nil {
		// Commit transaction
		err := app.deliverSqlTx.Commit()
		if err != nil {
			// TODO: wrapper error
			panic(err)
		}
		stake.ResetDeliverSqlTx()
		governance.ResetDeliverSqlTx()
	}
	dbHash := app.StoreApp.GetDbHash()
	res.Data = finalAppHash(ethAppCommit.Data, res.Data, dbHash, workingHeight, nil)

	app.BlockEnd = true

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

	//if store != nil {
	//	// TODO: save to DB
	//}
	return hash
}
