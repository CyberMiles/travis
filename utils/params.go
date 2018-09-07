package utils

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/CyberMiles/travis/sdk"
)

type Params struct {
	MaxVals                   uint16         `json:"max_vals" type:"uint"`    // maximum number of validators
	BackupVals                uint16         `json:"backup_vals" type:"uint"` // number of backup validators
	SelfStakingRatio          sdk.Rat        `json:"self_staking_ratio" type:"rat"`
	InflationRate             sdk.Rat        `json:"inflation_rate" type:"rat"`
	ValidatorSizeThreshold    sdk.Rat        `json:"validator_size_threshold" type:"rat"`
	UnstakeWaitingPeriod      uint64         `json:"unstake_waiting_period" type:"uint"`
	ProposalExpirePeriod      uint64         `json:"proposal_expire_period" type:"uint"`
	DeclareCandidacy          uint64         `json:"declare_candidacy" type:"uint"`
	UpdateCandidacy           uint64         `json:"update_candidacy" type:"uint"`
	TransferFundProposal      uint64         `json:"transfer_fund_proposal" type:"uint"`
	ChangeParamsProposal      uint64         `json:"change_params_proposal" type:"uint"`
	DeployLibEniProposal      uint64         `json:"deploy_libeni_proposal" type:"uint"`
	GasPrice                  uint64         `json:"gas_price" type:"uint"`
	MinStakingAmount          int64          `json:"min_staking_amount" type:"uint"`
	ValidatorsBlockAwardRatio sdk.Rat        `json:"validators_block_award_ratio" type:"rat"`
	MaxSlashingBlocks         int16          `json:"max_slashing_blocks" type:"uint"`
	SlashingRatio             sdk.Rat        `json:"slashing_ratio" type:"rat"`
	CubePubKeys               string         `json:"cube_pub_keys" type:"json"`
	LowPriceTxGasLimit        uint64         `json:"low_price_tx_gas_limit" type:"uint"`
	LowPriceTxSlotsCap        int            `json:"low_price_tx_slots_cap" type:"int"`
	SetCompRate               uint64         `json:"set_comp_rate" type:"uint"`
	FoundationAddress         string         `json:"foundation_address"`
}

func DefaultParams() *Params {
	return &Params{
		MaxVals:                   4,
		BackupVals:                1,
		SelfStakingRatio:          sdk.NewRat(10, 100),
		InflationRate:             sdk.NewRat(8, 100),
		ValidatorSizeThreshold:    sdk.NewRat(12, 100),
		UnstakeWaitingPeriod:      7 * 24 * 3600 / 10,
		ProposalExpirePeriod:      7 * 24 * 3600,
		DeclareCandidacy:          1e6, // gas setting for declareCandidacy
		UpdateCandidacy:           1e6, // gas setting for updateCandidacy
		TransferFundProposal:      2e6,
		ChangeParamsProposal:      2e6,
		DeployLibEniProposal:      2e6,
		GasPrice:                  2e9,
		MinStakingAmount:          1000,
		ValidatorsBlockAwardRatio: sdk.NewRat(80, 100),
		MaxSlashingBlocks:         12,
		SlashingRatio:             sdk.NewRat(1, 1000),
		CubePubKeys:               `[{"cube_batch":"01","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCiWpvDnwYFTqgSWPlA3VO8u+Yv\n9r8QGlRaYZFszUZEXUQxquGlFexMSVyFeqYjIokfPOEHHx2voqWgi3FKKlp6dkxw\nApP3T22y7Epqvtr+EfNybRta15snccZy47dY4UcmYxbGWFTaL66tz22pCAbjFrxY\n3IxaPPIjDX+FiXdJWwIDAQAB\n-----END PUBLIC KEY-----"},{"cube_batch":"02","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDQ8FL6/9zul+X7bFSRiWAzFiAE\n9vHYbClEHwlC7zUZ/JWzU7UT5S2qnYsseYF2WFjJtrGwHRAlTUyPtCpxV8f1uJsI\nl+/N9l6torUHwkhhib1catUSd/T72ltjvVyyg5LQjtRsskFnv3wM/yxYotrgnOs+\ndRpU6WI5XPCIyZqsGwIDAQAB\n-----END PUBLIC KEY-----"},{"cube_batch":"05","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCZ7Fw+1ddvy5OPFftbea0MxewW\nKUTb/E7B4/MHvLz2h7f7snyveFwxxj7QwxaCoVxobEq6AigIlUFUXLM8Y598/jts\nTaN+jh4xdoQN7qKwrbz1MWGf58Aa78Vnoj54B7V0LSajVbLJSZNUEI/24HLcG2iN\nTD3dSvH0ARvRJJ9hZQIDAQAB\n-----END PUBLIC KEY-----"}]`,
		LowPriceTxGasLimit:        500000, // Maximum gas limit for low-price transaction
		LowPriceTxSlotsCap:        100,    // Maximum number of low-price transaction slots per block
		SetCompRate:               21000,  // gas setting for setCompRate
		FoundationAddress:         "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
	}
}

var (
	// Keys for store prefixes
	ParamKey = []byte{0x01} // key for global parameters
	dirty    = false
	params *Params
)

// load/save the global params
func LoadParams(b []byte) {
	json.Unmarshal(b, params)
}

func UnloadParams() (b []byte) {
	b, _ = json.Marshal(*params)
	return
}

func GetParams() *Params {
	return params
}

func SetParams(p *Params) {
	params = p
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
			case reflect.Struct:
				switch reflect.TypeOf(fv.Interface()).Name() {
				case "Rat":
					v := sdk.NewRat(0, 1)
					if err := json.Unmarshal([]byte("\""+value+"\""), &v); err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				}
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
				var b []interface{}
				if err := json.Unmarshal([]byte(value), &b); err == nil {
					return true
				}
			case "string":
				return true
			case "rat":
				v := sdk.NewRat(0, 1)
				if err := json.Unmarshal([]byte("\""+value+"\""), &v); err == nil {
					return true
				}
			}
			return false
		}
	}

	return false
}
