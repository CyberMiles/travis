package api

import (
	"encoding/json"

	"github.com/spf13/cast"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/CyberMiles/travis/sdk/client"
	"github.com/CyberMiles/travis/types"
)

func (s *CmtRPCService) getParsedFromJson(path string, key []byte, ptr interface{}, height uint64) (int64, error) {
	bs, h, err := s.get(path, key, cast.ToInt64(height))
	if err != nil {
		return 0, err
	}
	if len(bs) == 0 {
		return h, client.ErrNoData()
	}
	err = json.Unmarshal(bs, ptr)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func (s *CmtRPCService) getParsedFromCdc(path string, key []byte, ptr interface{}, height uint64) (int64, error) {
	bs, h, err := s.get(path, key, cast.ToInt64(height))
	if err != nil {
		return 0, err
	}
	if len(bs) == 0 {
		return h, client.ErrNoData()
	}
	err = types.Cdc.UnmarshalBinary(bs, ptr)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func (s *CmtRPCService) get(path string, key []byte, height int64) ([]byte, int64, error) {
	node := s.backend.GetLocalClient()
	resp, err := node.ABCIQueryWithOptions(path, key,
		rpcclient.ABCIQueryOptions{Trusted: true, Height: int64(height)})
	if resp == nil {
		return nil, height, err
	}
	return resp.Response.Value, resp.Response.Height, err
}
