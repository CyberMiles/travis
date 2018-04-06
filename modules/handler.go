package modules

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/CyberMiles/travis/modules/auth"
	"fmt"

	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/types"
	"strings"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/modules/governance"
)

type Handler struct {
}

func (h Handler) CheckTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	// Verify signature
	res, tx, err = auth.VerifyTx(&ctx, tx)
	if err != nil {
		return res, fmt.Errorf("failed to verify signature: %v", err)
	}

	// Check nonce
	res, tx, err = nonce.ReplayCheck(ctx, store, tx)
	if err != nil {
		return res, fmt.Errorf("failed to check nonce: %v", err)
	}

	name, err := lookupRoute(tx)
	//fmt.Printf("Type of tx: %v\n", name)
	switch name {
	case "stake":
		return stake.CheckTx(ctx, store, tx)
	case "governance":
		return governance.CheckTx(ctx, store, tx)
	default:
		return res, errors.ErrUnknownTxType(tx.Unwrap())
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

	// Check nonce
	_, tx, err = nonce.ReplayCheck(ctx, store, tx)
	if err != nil {
		return res, fmt.Errorf("failed to check nonce: %v", err)
	}

	name, err := lookupRoute(tx)
	//fmt.Printf("Type of tx: %v\n", name)
	switch name {
	case "stake":
		return stake.DeliverTx(ctx, store, tx)
	case "governance":
		return governance.DeliverTx(ctx, store, tx)
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
