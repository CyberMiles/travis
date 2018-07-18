package utils

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/CyberMiles/travis/types"
	"github.com/ethereum/go-ethereum/common"
)

type Params struct {
	HoldAccount               common.Address `json:"hold_account"`         // PubKey where all bonded coins are held
	MaxVals                   uint16         `json:"max_vals" type:"uint"` // maximum number of validators
	SelfStakingRatio          string         `json:"self_staking_ratio" type:"float"`
	InflationRate             int64          `json:"inflation_rate" type:"uint"`
	ValidatorSizeThreshold    string         `json:"validator_size_threshold" type:"float"`
	UnstakeWaitingPeriod      uint64         `json:"unstake_waiting_period" type:"uint"`
	ProposalExpirePeriod      uint64         `json:"proposal_expire_period" type:"uint"`
	DeclareCandidacy          uint64         `json:"declare_candidacy" type:"uint"`
	UpdateCandidacy           uint64         `json:"update_candidacy" type:"uint"`
	TransferFundProposal      uint64         `json:"transfer_fund_proposal" type:"uint"`
	ChangeParamsProposal      uint64         `json:"change_params_proposal" type:"uint"`
	GasPrice                  uint64         `json:"gas_price" type:"uint"`
	MinStakingAmount          int64          `json:"min_staking_amount" type:"int"`
	ValidatorsBlockAwardRatio int64          `json:"validators_block_award_ratio" type:"int"`
	MaxSlashingBlocks         int16          `json:"max_slashing_blocks" type:"int"`
	SlashingRatio             string         `json:"slashing_ratio" type:"float"`
	CubePubKeys               string         `json:"cube_pub_keys" type:"json"`
}

func defaultParams() *Params {
	return &Params{
		HoldAccount:               HoldAccount,
		MaxVals:                   100,
		SelfStakingRatio:          "0.1",
		InflationRate:             8,
		ValidatorSizeThreshold:    "0.12",
		UnstakeWaitingPeriod:      7 * 24 * 3600 / 10,
		ProposalExpirePeriod:      7 * 24 * 3600 / 10,
		DeclareCandidacy:          1e6,
		UpdateCandidacy:           1e6,
		TransferFundProposal:      2e6,
		ChangeParamsProposal:      2e6,
		GasPrice:                  2e9,
		MinStakingAmount:          1000,
		ValidatorsBlockAwardRatio: 80,
		MaxSlashingBlocks:         12,
		SlashingRatio:             "0.001",
		CubePubKeys:               "{}",
	}
}

var (
	// Keys for store prefixes
	ParamKey = []byte{0x01} // key for global parameters
	params   = defaultParams()
	dirty    = false
)

// load/save the global params
func LoadParams(b []byte) {
	types.Cdc.UnmarshalBinary(b, params)
}

func UnloadParams() (b []byte) {
	b, _ = types.Cdc.MarshalBinary(*params)
	return
}

func GetParams() *Params {
	return params
}

func CleanParams() (before bool) {
	before = dirty
	dirty = false
	return
}

func SetParam(name, value string) bool {
	pv := reflect.ValueOf(params).Elem()
	top := pv.Type()
	for i := 0; i < pv.NumField(); i++ {
		fv := pv.Field(i)
		if top.Field(i).Tag.Get("json") == name {
			switch fv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if iv, err := strconv.ParseInt(value, 10, 64); err == nil {
					fv.SetInt(iv)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if iv, err := strconv.ParseUint(value, 10, 64); err == nil {
					fv.SetUint(iv)
				}
			case reflect.String:
				fv.SetString(value)
			}
			dirty = true
			return true
		}
	}

	return false
}

func CheckParamType(name, value string) bool {
	pv := reflect.ValueOf(params).Elem()
	top := pv.Type()
	for i := 0; i < pv.NumField(); i++ {
		if top.Field(i).Tag.Get("json") == name {
			switch top.Field(i).Tag.Get("type") {
			case "int":
				if _, err := strconv.ParseInt(value, 10, 64); err == nil {
					return true
				}
			case "uint":
				if _, err := strconv.ParseUint(value, 10, 64); err == nil {
					return true
				}
			case "float":
				if iv, err := strconv.ParseFloat(value, 64); err == nil {
					if iv > 0 {
						return true
					}
				}
			case "json":
				var s map[string]interface{}
				if err := json.Unmarshal([]byte(value), &s); err == nil {
					return true
				}
			case "string":
				return true
			}
			return false
		}
	}

	return false
}
