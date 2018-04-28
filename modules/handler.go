package modules

import (
	"fmt"

	"github.com/CyberMiles/travis/modules/auth"
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"

	"strings"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/go-wire/data"
	"os"
	"github.com/tendermint/tmlibs/merkle"
	"github.com/cosmos/cosmos-sdk/errors"
)

type Handler struct {
}

func (h Handler) CheckTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	// Verify signature
	res, tx, err = auth.VerifyTx(&ctx, tx)
	if err != nil {
		return res, fmt.Errorf("failed to verify signature: %v", err)
	}


	// make sure it is a the nonce Tx (Tx from this package)
	nonceTx, ok := tx.Unwrap().(nonce.Tx)
	if !ok {
		return res, nonce.ErrNoNonce()
	}

	name, err := lookupRoute(nonceTx.Tx)

	if name == "stake" {
		res, err = stake.CheckTx(ctx, store, nonceTx.Tx)
	} else if name == "governance" {
		res, err = governance.CheckTx(ctx, store, nonceTx.Tx)
	}

	if err != nil {
		return res, err
	}

	// Check nonce
	res, tx, err = nonce.ReplayCheck(ctx, store, tx)
	if err != nil {
		return res, fmt.Errorf("failed to check nonce: %v", err)
	}


	return
}

func (h Handler) DeliverTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
	//_, err = h.CheckTx(ctx, store, tx)
	//if err != nil {
	//	return
	//}

	// Verify signature
	_, tx, err = auth.VerifyTx(&ctx, tx)
	if err != nil {
		return res, fmt.Errorf("failed to verify signature: %v", err)
	}

	hash := merkle.SimpleHashFromBinary(tx)

	// Check nonce
	_, tx, err = nonce.ReplayCheck(ctx, store, tx)
	if err != nil {
		return res, fmt.Errorf("failed to check nonce: %v", err)
	}

	name, err := lookupRoute(tx)
	//fmt.Printf("Type of tx: %v\n", name)
	switch name {
	case "stake":
		return stake.DeliverTx(ctx, store, tx, hash)
	case "governance":
		return governance.DeliverTx(ctx, store, tx, hash)
	default:
		return res, errors.ErrUnknownTxType(tx.Unwrap())
	}

	return
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
