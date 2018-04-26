package api

import (
	"errors"
	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/CyberMiles/travis/modules/auth"
	"github.com/CyberMiles/travis/modules/nonce"
	ttypes "github.com/CyberMiles/travis/types"
)

func (s *CmtRPCService) wrapAndSignTx(tx sdk.Tx, address string, sequence uint64) (sdk.Tx, error) {
	// wrap
	// only add the actual signer to the nonce
	signers := []common.Address{getSigner(address)}
	if sequence <= 0 {
		// calculate default sequence
		err := s.getSequence(signers, &sequence)
		if err != nil {
			return sdk.Tx{}, err
		}
		sequence = sequence + 1
	}
	tx = nonce.NewTx(sequence, signers, tx)
	tx = auth.NewSig(tx).Wrap()

	// sign
	err := s.signTx(tx, address)
	if err != nil {
		return sdk.Tx{}, err
	}
	return tx, err
}

// sign the transaction with private key
func (s *CmtRPCService) signTx(tx sdk.Tx, address string) error {
	// validate tx client-side
	err := tx.ValidateBasic()
	if err != nil {
		return err
	}

	if sign, ok := tx.Unwrap().(ttypes.Signable); ok {
		if address == "" {
			return errors.New("address is required to sign tx")
		}
		err := s.sign(sign, address)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *CmtRPCService) sign(data ttypes.Signable, address string) error {
	ethTx := types.NewTransaction(
		0,
		common.Address([20]byte{}),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		data.SignBytes(),
	)

	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	wallet, err := s.am.Find(account)
	if err != nil {
		return err
	}

	ethChainId := int64(s.backend.ethConfig.NetworkId)
	signed, err := wallet.SignTx(account, ethTx, big.NewInt(ethChainId))
	if err != nil {
		return err
	}

	return data.Sign(signed)
}

func getSigner(address string) (res common.Address) {
	// this could be much cooler with multisig...
	res = common.HexToAddress(address)
	return res
}
