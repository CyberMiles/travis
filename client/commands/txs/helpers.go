package txs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/client/commands"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/CyberMiles/travis/commons"
)

// Validatable represents anything that can be Validated
type Validatable interface {
	ValidateBasic() error
}

func GetSigner() common.Address {
	address := viper.GetString(FlagAddress)

	if address == "" {
		fmt.Errorf("--address is required to sign tx")
		return common.Address{}
	}

	return common.HexToAddress(address)
}

// DoTx is a helper function for the lazy :)
//
// It uses only public functions and goes through the standard sequence of
// wrapping the tx with middleware layers, signing it, either preparing it,
// or posting it and displaying the result.
//
// If you want a non-standard flow, just call the various functions directly.
// eg. if you already set the middleware layers in your code, or want to
// output in another format.
func DoTx(tx sdk.Tx) (err error) {
	address := viper.GetString(FlagAddress)
	from := common.HexToAddress(address)

	prompt := fmt.Sprintf("Please enter passphrase for %s: ", address)
	passphrase, err := getPassword(prompt)
	if err != nil {
		return err
	}

	txBytes, err := wrapAndSign(tx, from, passphrase)
	if err != nil {
		return err
	}
	commit := viper.GetString(FlagType)
	if commit == "commit" {
		bres, err := broadcastTxCommit(txBytes)
		if err != nil {
			return err
		}
		if bres == nil {
			return nil // successful prep, nothing left to do
		}
		return OutputTx(bres) // print response of the post

	} else {
		bres, err := broadcastTxSync(txBytes)
		if err != nil {
			return err
		}
		if bres == nil {
			return nil // successful prep, nothing left to do
		}
		return OutputTxSync(bres) // print response of the post
	}
}

func wrapAndSign(tx sdk.Tx, from common.Address, passphrase string) (hexutil.Bytes, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	ethTx := types.NewContractCreation(
		getNonce(from),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		data,
	)

	am, _, _ := commons.MakeAccountManager()
	_, err = commons.UnlockAccount(am, from, passphrase, nil)
	if err != nil {
		return nil, err
	}

	account := accounts.Account{Address: from}
	wallet, err := am.Find(account)
	signed, err := wallet.SignTx(account, ethTx, big.NewInt(111))
	if err != nil {
		return nil, err
	}

	encodedTx, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}

	return encodedTx, nil
}

func getNonce(addr common.Address) uint64 {
	//add the nonce tx layer to the tx
	input := viper.GetInt(FlagNonce)

	if input >= 0 {
		return uint64(input)
	}

	var nonce uint64
	//get nonce
	client, err := ethclient.Dial(getEthUrl())
	if err != nil {
		fmt.Errorf(err.Error())
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		nonce, err = client.NonceAt(ctx, addr, nil)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
	return nonce
}

func getEthUrl() string {
	node := viper.GetString(commands.NodeFlag)
	u, _ := url.Parse(node)
	return fmt.Sprintf("http://%s:%d", u.Hostname(), 8545)
}

func broadcastTxSync(packet []byte) (*ctypes.ResultBroadcastTx, error) {
	// post the bytes
	node := commands.GetNode()
	return node.BroadcastTxSync(packet)
}

func broadcastTxCommit(packet []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	// post the bytes
	node := commands.GetNode()
	return node.BroadcastTxCommit(packet)
}

// PrepareOrPostTx checks the flags to decide to prepare the tx for future
// multisig, or to post it to the node. Returns error on any failure.
// If no error and the result is nil, it means it already wrote to file,
// no post, no need to do more.
func PrepareOrPostTx(tx sdk.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	wrote, err := PrepareTx(tx)
	// error in prep
	if err != nil {
		return nil, err
	}
	// successfully wrote the tx!
	if wrote {
		return nil, nil
	}
	// or try to post it
	return PostTx(tx)
}

func PrepareOrPostTxSync(tx sdk.Tx) (*ctypes.ResultBroadcastTx, error) {
	wrote, err := PrepareTx(tx)
	// error in prep
	if err != nil {
		return nil, err
	}
	// successfully wrote the tx!
	if wrote {
		return nil, nil
	}
	// or try to post it
	return PostTxSync(tx)
}

// PostTx does all work once we construct a proper struct
// it validates the data, signs if needed, transforms to bytes,
// and posts to the node.
func PostTxSync(tx sdk.Tx) (*ctypes.ResultBroadcastTx, error) {
	packet := wire.BinaryBytes(tx)
	// post the bytes
	node := commands.GetNode()
	return node.BroadcastTxSync(packet)
}

// PrepareTx checks for FlagPrepare and if set, write the tx as json
// to the specified location for later multi-sig.  Returns true if it
// handled the tx (no futher work required), false if it did nothing
// (and we should post the tx)
func PrepareTx(tx sdk.Tx) (bool, error) {
	prep := viper.GetString(FlagPrepare)
	if prep == "" {
		return false, nil
	}

	js, err := data.ToJSON(tx)
	if err != nil {
		return false, err
	}
	err = writeOutput(prep, js)
	if err != nil {
		return false, err
	}
	return true, nil
}

// PostTx does all work once we construct a proper struct
// it validates the data, signs if needed, transforms to bytes,
// and posts to the node.
func PostTx(tx sdk.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	packet := wire.BinaryBytes(tx)
	// post the bytes
	node := commands.GetNode()
	return node.BroadcastTxCommit(packet)
}

// OutputTx validates if success and prints the tx result to stdout
func OutputTx(res *ctypes.ResultBroadcastTxCommit) error {
	if res.CheckTx.IsErr() {
		return errors.Errorf("CheckTx: (%d): %s", res.CheckTx.Code, res.CheckTx.Log)
	}
	if res.DeliverTx.IsErr() {
		return errors.Errorf("DeliverTx: (%d): %s", res.DeliverTx.Code, res.DeliverTx.Log)
	}
	js, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(js))
	return nil
}

func OutputTxSync(res *ctypes.ResultBroadcastTx) error {
	js, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(js))
	return nil
}

// if we read from non-tty, we just need to init the buffer reader once,
// in case we try to read multiple passwords
var buf *bufio.Reader

func inputIsTty() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func stdinPassword() (string, error) {
	if buf == nil {
		buf = bufio.NewReader(os.Stdin)
	}
	pass, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

func getPassword(prompt string) (pass string, err error) {
	if inputIsTty() {
		pass, err = speakeasy.Ask(prompt)
	} else {
		pass, err = stdinPassword()
	}
	return
}

func writeOutput(file string, d []byte) error {
	var writer io.Writer
	if file == "-" {
		writer = os.Stdout
	} else {
		f, err := os.Create(file)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		writer = f
	}

	_, err := writer.Write(d)
	// this returns nil if err == nil
	return errors.WithStack(err)
}

func readInput(file string) ([]byte, error) {
	var reader io.Reader
	// get the input stream
	if file == "" || file == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer f.Close()
		reader = f
	}

	// and read it all!
	data, err := ioutil.ReadAll(reader)
	return data, errors.WithStack(err)
}
