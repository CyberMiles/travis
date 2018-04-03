package modules

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/CyberMiles/travis/modules/auth"
	"fmt"

	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/types"
)

type Handler struct {
}

func (h Handler) CheckTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	// Verify signature
	res, tx, ctx, err = auth.VerifyTx(ctx, tx)
	if err != nil {
		return res, fmt.Errorf("failed to verify signature")
	}

	// Check nonce
	res, tx, err = nonce.ReplayCheck(ctx, store, tx)
	if err != nil {
		return res, fmt.Errorf("failed to check nonce")
	}

	fmt.Printf("Type of inner tx: %v", tx.Unwrap())

	//switch txInner := tx.Unwrap().(type) {
	//case TxDeclareCandidacy:
	//	return sdk.NewCheck(params.GasDeclareCandidacy, ""),
	//		checker.declareCandidacy(txInner)
	//case TxEditCandidacy:
	//	return sdk.NewCheck(params.GasEditCandidacy, ""),
	//		checker.editCandidacy(txInner)
	//case TxWithdraw:
	//	return sdk.NewCheck(params.GasWithdraw, ""),
	//		checker.withdraw(txInner)
	//case TxProposeSlot:
	//	_, err := checker.proposeSlot(txInner)
	//	return sdk.NewCheck(params.GasProposeSlot, ""), err
	//case TxAcceptSlot:
	//	return sdk.NewCheck(params.GasAcceptSlot, ""),
	//		checker.acceptSlot(txInner)
	//case TxWithdrawSlot:
	//	return sdk.NewCheck(params.GasWithdrawSlot, ""),
	//		checker.withdrawSlot(txInner)
	//case TxCancelSlot:
	//	return sdk.NewCheck(params.GasCancelSlot, ""),
	//		checker.cancelSlot(txInner)
	//}

	return

}

func (h Handler) DeliverTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
	_, err = h.CheckTx(ctx, store, tx)
	if err != nil {
		return
	}

	return
}

func (Handler) InitState(key, value string, store state.SimpleDB) error {
	//params := loadParams(store)
	//switch key {
	//case "allowed_bond_denom":
	//	params.AllowedBondDenom = value
	//case "max_vals",
	//	"gas_bond",
	//	"gas_unbond":
	//
	//	// TODO: enforce non-negative integers in input
	//	i, err := strconv.Atoi(value)
	//	if err != nil {
	//		return fmt.Errorf("input must be integer, Error: %v", err.Error())
	//	}
	//
	//	switch key {
	//	case "max_vals":
	//		params.MaxVals = uint16(i)
	//	}
	//case "validator":
	//	setValidator(value)
	//default:
	//	return errors.ErrUnknownKey(key)
	//}
	//
	//saveParams(store, params)
	return nil
}

//func setValidator(value string) error {
//	var val genesisValidator
//	err := data.FromJSON([]byte(value), &val)
//	if err != nil {
//		return fmt.Errorf("error reading validators")
//	}
//
//	// create and save the empty candidate
//	bond := GetCandidateByAddress(val.Address)
//	if bond != nil {
//		return ErrCandidateExistsAddr()
//	}
//
//	candidate := NewCandidate(val.PubKey, val.Address, val.Power, val.Power, "Y")
//	SaveCandidate(candidate)
//
//	return nil
//}

// load/save the global staking params
//func loadParams(store state.SimpleDB) (params Params) {
//	b := store.Get(ParamKey)
//	if b == nil {
//		return defaultParams()
//	}
//
//	err := wire.ReadBinaryBytes(b, &params)
//	if err != nil {
//		panic(err) // This error should never occure big problem if does
//	}
//
//	return
//}
//func saveParams(store state.SimpleDB, params Params) {
//	b := wire.BinaryBytes(params)
//	store.Set(ParamKey, b)
//}