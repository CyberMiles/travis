package api

import (
	"bytes"
	"time"

	//"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/ethereum/go-ethereum/core"
)

var (
	local = false
)

//----------------------------------------------------------------------
// Transactions sent via the go-ethereum rpc need to be routed to tendermint

// listen for txs and forward to tendermint
func (b *Backend) txBroadcastLoop() {
	//b.txSub = b.ethereum.EventMux().Subscribe(core.TxPreEvent{})

	b.txsCh = make(chan core.NewTxsEvent, 10)
	b.txsSub = b.ethereum.TxPool().SubscribeNewTxsEvent(b.txsCh)

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

	for {
		select {
		// Handle NewTxsEvent
		case ev := <- b.txsCh:
			for _, tx := range ev.Txs {
				result, err := b.BroadcastTxSync(tx)
				if err != nil {
					log.Error("Broadcast error", "err", err)
				} else {
					if result.Code != uint32(0) {
						go removeTx(b, tx)
					} else {
						// TODO: do something else?
					}
				}
			}
			// System stopped
		case <-b.txsSub.Err():
			return
		}
	}
}

// BroadcastTx broadcasts a transaction to tendermint core
// #unstable
func (b *Backend) BroadcastTxSync(tx *ethTypes.Transaction) (*ctypes.ResultBroadcastTx, error) {
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

func (b *Backend) BroadcastTxCommit(tx *ethTypes.Transaction) (*ctypes.ResultBroadcastTxCommit, error) {
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return nil, err
	}

	if local {
		return b.localClient.BroadcastTxCommit(buf.Bytes())
	} else {
		return b.client.BroadcastTxCommit(buf.Bytes())
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
	// TODO: add Remove in txPool ???
	//b.Ethereum().TxPool().Remove(tx.Hash())
}
