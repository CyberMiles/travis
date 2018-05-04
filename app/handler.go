package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

func (app BaseApp) checkHandler(ctx types.Context, store state.SimpleDB, tx *ethTypes.Transaction) abci.ResponseCheckTx {
	currentState, from, nonce, resp := app.EthApp.basicCheck(tx)
	if resp.Code != abci.CodeTypeOK {
		return resp
	}
	ctx.WithSigners(from)

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

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

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

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func output(v interface{}) error {
	blob, err := data.ToJSON(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "%s\n", blob)
	return nil
}
