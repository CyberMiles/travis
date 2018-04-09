package api

import (
	"bytes"
	"time"

	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var (
	local = false
)

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
func (b *Backend) txBroadcastLoop() {
	b.txSub = b.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	for tries := 0; tries < 3; tries++ { // wait a moment for localClient initialized properly
		time.Sleep(time.Second)
		if b.localClient != nil {
			if _, err := b.localClient.Status(); err != nil {
				log.Info("Using local client for forwarding tx to tendermint!")
				local = true
				break
			}
		}
	}

	if !local {
		waitForServer(b.client)
	}

	for obj := range b.txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		result, err := b.BroadcastTx(event.Tx)
		if err != nil {
			log.Error("Broadcast error", "err", err)
		} else {
			if result.Code != uint32(0) {
				go removeTx(b, event.Tx)
			} else {
				// TODO: do something else?
			}
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
// #unstable
func (b *Backend) BroadcastTx(tx *ethTypes.Transaction) (*ctypes.ResultBroadcastTx, error) {
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return nil, err
	}

	if local {
		return b.localClient.BroadcastTxSync(buf.Bytes())
	} else {
		return b.client.BroadcastTxSync(buf.Bytes())
	}

}

//----------------------------------------------------------------------
// wait for Tendermint to open the socket and run http endpoint

func waitForServer(c *rpcClient.HTTP) {
	for {
		_, err := c.Status()
		if err == nil {
			break
		}

		log.Info("Waiting for tendermint endpoint to start", "err", err)
		time.Sleep(time.Second * 3)
	}
}

func removeTx(b *Backend, tx *ethTypes.Transaction) {
	b.Ethereum().TxPool().Remove(tx.Hash())
}
