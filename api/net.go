package api

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// We must implement our own net service since we don't have access to `internal/ethapi`

// NetRPCService mirrors the implementation of `internal/ethapi`
// #unstable
type NetRPCService struct {
	backend        *Backend
	networkVersion uint64
}

// NewNetRPCService creates a new net API instance.
// #unstable
func NewNetRPCService(b *Backend) *NetRPCService {
	return &NetRPCService{
		backend:        b,
		networkVersion: b.ethConfig.NetworkId,
	}
}

// Listening returns an indication if the node is listening for network connections.
// #unstable
func (s *NetRPCService) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *NetRPCService) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.backend.PeerCount())
}

// Version returns the current ethereum protocol version.
// #unstable
func (s *NetRPCService) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}
