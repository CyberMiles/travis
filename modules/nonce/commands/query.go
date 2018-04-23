package commands

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/cosmos/cosmos-sdk/client"
	"github.com/CyberMiles/travis/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/ethereum/go-ethereum/common"
	"fmt"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	"github.com/cosmos/cosmos-sdk/client"
)

// NonceQueryCmd - command to query an nonce account
var NonceQueryCmd = &cobra.Command{
	Use:   "nonce [address]",
	Short: "Get details of a nonce sequence number, with proof",
	RunE:  commands.RequireInit(nonceQueryCmd),
}

func nonceQueryCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Missing required argument [address]")
	}
	addr := strings.Join(args, ",")

	signers, err := commands.ParseActors(addr)
	if err != nil {
		return err
	}

	seq, height, err := doNonceQuery(signers)
	if err != nil {
		return err
	}

	return query.OutputProof(seq, height)
}

func doNonceQuery(signers []common.Address) (sequence uint32, height int64, err error) {
	//fmt.Printf("doNonceQuery, before prefixed: %s\n", hex.EncodeToString(nonce.GetSeqKey(signers)))
	//key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(signers))
	key := nonce.GetSeqKey(signers)
	//fmt.Printf("doNonceQuery, after prefixed: %s\n", hex.EncodeToString(key))
	prove := !viper.GetBool(commands.FlagTrustNode)
	prove = false
	height, err = getParsed(key, &sequence, query.GetHeight(), prove)
	//if client.IsNoDataErr(err) {
	// TODO: 这里直接判断err是否为nil不准确
	if err != nil {
		// no data, return sequence 0
		return 0, 0, nil
	}
	return
}


// GetParsed does most of the work of the query commands, but is quite
// opinionated, so if you want more control about parsing, call Get
// directly.
//
// It will try to get the proof for the given key.  If it is successful,
// it will return the height and also unserialize proof.Data into the data
// argument (so pass in a pointer to the appropriate struct)
func getParsed(key []byte, data interface{}, height int64, prove bool) (int64, error) {
	bs, h, err := get(key, height, prove)
	if err != nil {
		return 0, err
	}
	err = wire.ReadBinaryBytes(bs, data)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// Get queries the given key and returns the value stored there and the
// height we checked at.
//
// If prove is true (and why shouldn't it be?),
// the data is fully verified before returning.  If prove is false,
// we just repeat whatever any (potentially malicious) node gives us.
// Only use that if you are running the full node yourself,
// and it is localhost or you have a secure connection (not HTTP)
func get(key []byte, height int64, prove bool) (data.Bytes, int64, error) {
	if height < 0 {
		return nil, 0, fmt.Errorf("Height cannot be negative")
	}

	if !prove {
		node := commands.GetNode()
		resp, err := node.ABCIQueryWithOptions("/key", key,
			rpcclient.ABCIQueryOptions{Trusted: true, Height: int64(height)})
		if resp == nil {
			return nil, height, err
		}
		if len(resp.Response.Value) > 0 {
			return data.Bytes(resp.Response.Value), resp.Response.Height, err
		} else {
			return data.Bytes{}, 0, client.ErrNoData()
		}
	}
	val, h, _, err := query.GetWithProof(key, height)
	return val, h, err
}