package app

import (
	"encoding/json"
	"strings"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

func (app BaseApp) checkHandler(ctx types.Context, store state.SimpleDB, tx *ethTypes.Transaction) abci.ResponseCheckTx {
	currentState, from, nonce, resp := app.EthApp.basicCheck(tx)
	if resp.Code != abci.CodeTypeOK {
		return resp
	}
	ctx.WithSigners(from)
	ctx.SetNonce(nonce)

	var travisTx sdk.Tx
	if err := json.Unmarshal(tx.Data(), &travisTx); err != nil {
		return errors.CheckResult(err)
	}

	name, err := lookupRoute(travisTx)
	if err != nil {
		return errors.CheckResult(err)
	}

	var res sdk.CheckResult
	if name == "stake" {
		res, err = stake.CheckTx(ctx, store, travisTx)
	} else if name == "governance" {
		res, err = governance.CheckTx(ctx, store, travisTx)
	}

	if err != nil {
		return errors.CheckResult(err)
	}

	utils.NonceCheckedTx[tx.Hash()] = true
	currentState.SetNonce(from, nonce+1)

	return res.ToABCI()
}

func (app BaseApp) deliverHandler(ctx types.Context, store state.SimpleDB, tx *ethTypes.Transaction) abci.ResponseDeliverTx {
	hash := tx.Hash().Bytes()

	var travisTx sdk.Tx
	if err := json.Unmarshal(tx.Data(), &travisTx); err != nil {
		return errors.DeliverResult(err)
	}

	signer := ethTypes.NewEIP155Signer(tx.ChainId())

	// Make sure the transaction is signed properly
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return errors.DeliverResult(err)
	}
	ctx.WithSigners(from)

	name, err := lookupRoute(travisTx)
	if err != nil {
		return errors.DeliverResult(err)
	}

	var res sdk.DeliverResult
	switch name {
	case "stake":
		res, err = stake.DeliverTx(ctx, store, travisTx, hash)
	case "governance":
		res, err = governance.DeliverTx(ctx, store, travisTx, hash)
	default:
		return errors.DeliverResult(errors.ErrUnknownTxType(travisTx.Unwrap()))
	}

	if err != nil {
		return errors.DeliverResult(err)
	}

	// no error, call ethereum app to add nonce
	app.EthApp.backend.AddNonce(from)
	// accumulate gasFee
	app.StoreApp.TotalUsedGasFee.Add(app.StoreApp.TotalUsedGasFee, res.GasFee)
	return res.ToABCI()
}

func lookupRoute(tx sdk.Tx) (string, error) {
	kind, err := tx.GetKind()
	if err != nil {
		return "", err
	}
	// grab everything before the /
	name := strings.SplitN(kind, "/", 2)[0]
	return name, nil
}
