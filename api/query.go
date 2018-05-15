package api

import (
	"github.com/spf13/cast"

	"github.com/CyberMiles/travis/sdk/client"
	"github.com/tendermint/go-wire/data"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"encoding/json"
)

func (s *CmtRPCService) getParsed(path string, key []byte, data interface{}, height uint64) (int64, error) {
	bs, h, err := s.get(path, key, cast.ToInt64(height))
	if err != nil {
		return 0, err
	}
	if len(bs) == 0 {
		return h, client.ErrNoData()
	}
	err = json.Unmarshal(bs, data)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func (s *CmtRPCService) get(path string, key []byte, height int64) (data.Bytes, int64, error) {
	node := s.backend.localClient
	resp, err := node.ABCIQueryWithOptions(path, key,
		rpcclient.ABCIQueryOptions{Trusted: true, Height: int64(height)})
	if resp == nil {
		return nil, height, err
	}
	return data.Bytes(resp.Response.Value), resp.Response.Height, err
}
