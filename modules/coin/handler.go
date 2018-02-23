package coin

import (
	"os"
	"path/filepath"
	"fmt"
	"time"
	"math/big"
	"bytes"

	//abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/abci/client"
)

const (
	//NameCoin - name space of the coin module
	NameCoin = "coin"
	// CostSend is GasAllocation per input/output
	CostSend = int64(10)
	// CostCredit is GasAllocation of a credit allocation
	CostCredit = int64(20)
	// Ethereum default keystore directory
	datadirDefaultKeyStore = "keystore"
	emHome = "/Users/dragon/.ethermint"
	ETHERMINT_ADDR = "localhost:8848"
)

// Handler includes an accountant
type Handler struct {
	stack.PassInitValidate
}

var (
	_ stack.Dispatchable = Handler{}
	client, err = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
)

// NewHandler - new accountant handler for the coin module
func NewHandler() Handler {
	client.Start()
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameCoin
}

// AssertDispatcher - to fulfill Dispatchable interface
func (Handler) AssertDispatcher() {}

// CheckTx checks if there is enough money in the account
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (res sdk.CheckResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case SendTx:
		// price based on inputs and outputs
		used := int64(len(t.Inputs) + len(t.Outputs))
		return sdk.NewCheck(used*CostSend, ""), h.checkSendTx(ctx, store, t)
	case CreditTx:
		// default price of 20, constant work
		return sdk.NewCheck(CostCredit, ""), h.creditTx(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// DeliverTx moves the money
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, cb sdk.Deliver) (res sdk.DeliverResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case SendTx:
		return h.sendTx(ctx, store, t, cb)
	case CreditTx:
		return res, h.creditTx(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// InitState - sets the genesis account balance
func (h Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	if module != NameCoin {
		return "", errors.ErrUnknownModule(module)
	}
	switch key {
	case "account":
		return setAccount(store, value)
	case "issuer":
		return setIssuer(store, value)
	}
	return "", errors.ErrUnknownKey(key)
}

func (h Handler) sendTx(ctx sdk.Context, store state.SimpleDB,
	send SendTx, cb sdk.Deliver) (res sdk.DeliverResult, err error) {

	err = checkTx(ctx, send)
	if err != nil {
		return res, err
	}

	// deduct from all input accounts
	//senders := sdk.Actors{}
	//for _, in := range send.Inputs {
	//	_, err = ChangeCoins(store, in.Address, in.Coins.Negative())
	//	if err != nil {
	//		return res, err
	//	}
	//	senders = append(senders, in.Address)
	//}
	//
	//// add to all output accounts
	//for _, out := range send.Outputs {
	//	// TODO: cleaner way, this makes sure we don't consider
	//	// incoming ibc packets with our chain to be remote packets
	//	if out.Address.ChainID == ctx.ChainID() {
	//		out.Address.ChainID = ""
	//	}
	//
	//	_, err = ChangeCoins(store, out.Address, out.Coins)
	//	if err != nil {
	//		return res, err
	//	}
	//	// now send ibc packet if needed...
	//	if out.Address.ChainID != "" {
	//		// FIXME: if there are many outputs, we need to adjust inputs
	//		// so the amounts in and out match.  how?
	//		inputs := make([]TxInput, len(send.Inputs))
	//		for i := range send.Inputs {
	//			inputs[i] = send.Inputs[i]
	//			inputs[i].Address = inputs[i].Address.WithChain(ctx.ChainID())
	//		}
	//
	//		outTx := NewSendTx(inputs, []TxOutput{out})
	//		packet := ibc.CreatePacketTx{
	//			DestChain:   out.Address.ChainID,
	//			Permissions: senders,
	//			Tx:          outTx,
	//		}
	//		ibcCtx := ctx.WithPermissions(ibc.AllowIBC(NameCoin))
	//		_, err := cb.DeliverTx(ibcCtx, store, packet.Wrap())
	//		if err != nil {
	//			return res, err
	//		}
	//	}
	//}
	//
	//// now we build the tags
	//tags := make([]*abci.KVPair, 0, 1+len(send.Inputs)+len(send.Outputs))
	//
	//tags = append(tags, abci.KVPairInt("height", int64(ctx.BlockHeight())))
	//
	//for _, in := range send.Inputs {
	//	addr := in.Address.String()
	//	tags = append(tags, abci.KVPairString("coin.sender", addr))
	//}
	//
	//for _, out := range send.Outputs {
	//	addr := out.Address.String()
	//	tags = append(tags, abci.KVPairString("coin.receiver", addr))
	//}
	//
	//// a-ok!
	//return sdk.DeliverResult{Tags: tags}, nil

	tx := types.NewTransaction(
		0,
		common.Address([20]byte{}),
		big.NewInt(0x2386f26fc10000),
		big.NewInt(0x15f90),
		big.NewInt(0x430e23400),
		nil,
	)

	am, _, _ := makeAccountManager()
	coinbase := common.Address{0x7e, 0xff, 0x12, 0x2b, 0x94, 0x89, 0x7e, 0xa5, 0xb0, 0xe2, 0xa9, 0xab, 0xf4, 0x7b, 0x86, 0x33, 0x7f, 0xaf, 0xeb, 0xdc}
	suc, err := UnlockAccount(am, coinbase, "1234", nil)
	fmt.Printf("unlock result: %v\n", suc)
	fmt.Printf("unlock error: %v\n", err)

	account := accounts.Account{Address: coinbase}
	wallet, err := am.Find(account)
	signed, err := wallet.SignTx(account, tx, big.NewInt(15))
	if err != nil {
		fmt.Errorf("error")
	}

	buf := new(bytes.Buffer)
	if err := signed.EncodeRLP(buf); err != nil {
		fmt.Errorf("error")
	}
	//params := map[string]interface{}{
	//	"tx": buf.Bytes(),
	//}

	resp, err := client.DeliverTxSync(buf.Bytes())
	fmt.Printf("ethermint DeliverTx response: %v\n", resp)

	//var result interface{}
	//res, err := client.Call("broadcast_tx_sync", params, &result)
	//fmt.Printf("%#v %#v", res, err)

	return sdk.DeliverResult{}, nil
}

func (h Handler) creditTx(ctx sdk.Context, store state.SimpleDB,
	credit CreditTx) error {

	// first check permissions!!
	info, err := loadHandlerInfo(store)
	if err != nil {
		return err
	}
	if info.Issuer.Empty() || !ctx.HasPermission(info.Issuer) {
		return errors.ErrUnauthorized()
	}

	// load up the account
	addr := ChainAddr(credit.Debitor)
	acct, err := GetAccount(store, addr)
	if err != nil {
		return err
	}

	// make and check changes
	acct.Coins = acct.Coins.Plus(credit.Credit)
	if !acct.Coins.IsNonnegative() {
		return ErrInsufficientFunds()
	}
	acct.Credit = acct.Credit.Plus(credit.Credit)
	if !acct.Credit.IsNonnegative() {
		return ErrInsufficientCredit()
	}

	err = storeAccount(store, addr.Bytes(), acct)
	return err
}

func checkTx(ctx sdk.Context, send SendTx) error {
	// check if all inputs have permission
	for _, in := range send.Inputs {
		if !ctx.HasPermission(in.Address) {
			return errors.ErrUnauthorized()
		}
	}
	return nil
}

func (Handler) checkSendTx(ctx sdk.Context, store state.SimpleDB, send SendTx) error {
	err := checkTx(ctx, send)
	if err != nil {
		return err
	}
	// now make sure there is money
	for _, in := range send.Inputs {
		_, err := CheckCoins(store, in.Address, in.Coins.Negative())
		if err != nil {
			return err
		}
	}
	return nil
}

func setAccount(store state.SimpleDB, value string) (log string, err error) {
	var acc GenesisAccount
	err = data.FromJSON([]byte(value), &acc)
	if err != nil {
		return "", err
	}
	acc.Balance.Sort()
	addr, err := acc.GetAddr()
	if err != nil {
		return "", ErrInvalidAddress()
	}
	// this sets the permission for a public key signature, use that app
	actor := auth.SigPerm(addr)
	err = storeAccount(store, actor.Bytes(), acc.ToAccount())
	if err != nil {
		return "", err
	}
	return "Success", nil
}

// setIssuer sets a permission for some super-powerful account to
// mint money
func setIssuer(store state.SimpleDB, value string) (log string, err error) {
	var issuer sdk.Actor
	err = data.FromJSON([]byte(value), &issuer)
	if err != nil {
		return "", err
	}
	err = storeIssuer(store, issuer)
	if err != nil {
		return "", err
	}
	return "Success", nil
}

func makeAccountManager() (*accounts.Manager, string, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	//if conf.UseLightweightKDF {
	//	scryptN = keystore.LightScryptN
	//	scryptP = keystore.LightScryptP
	//}

	keydir := filepath.Join(emHome, datadirDefaultKeyStore)
	fmt.Println(keydir)

	ephemeral := keydir
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}

	return accounts.NewManager(backends...), ephemeral, nil
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

func UnlockAccount(am *accounts.Manager, addr common.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, fmt.Errorf("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}