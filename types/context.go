package types

import (
	"github.com/ethereum/go-ethereum/common"
	"sort"
	"bytes"
	"math/rand"
)

type nonce int64

type Context struct {
	id     nonce
	chain  string
	height int64
	signers []common.Address
}

func NewContext(chain string, height int64) Context {
	return Context{
		id: 	nonce(rand.Int63()),
		chain: 	chain,
		height:	height,
	}
}

func (c Context) ChainID() string {
	return c.chain
}

func (c Context) BlockHeight() int64 {
	return c.height
}

func (c Context) WithSigners(signers ...common.Address) Context {
	return Context{
		id: c.id,
		chain: c.chain,
		height: c.height,
		signers: append(c.signers, signers...),
	}
}

func (c Context) HasSigner(signer common.Address) bool {
	for _, s := range c.signers {
		if s == signer {
			return true
		}
	}
	return false
}

func (c Context) GetSigners() []common.Address {
	return c.signers
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c Context) Reset() Context {
	return Context{
		id:     c.id,
		chain:  c.chain,
		height: c.height,
	}
}

//////////////////////////////// Sort Interface
// USAGE sort.Sort(ByAll(<common.Address>))

// ByAll implements sort.Interface for []common.Address.
// It sorts be the Address
type ByAll []common.Address

// Verify the sort interface at compile time
var _ sort.Interface = ByAll{}

func (a ByAll) Len() int      { return len(a) }
func (a ByAll) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAll) Less(i, j int) bool {
	return bytes.Compare(a[i].Bytes(), a[j].Bytes()) == -1
}