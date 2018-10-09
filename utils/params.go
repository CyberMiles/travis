package utils

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/CyberMiles/travis/sdk"
)

type Params struct {
	MaxVals                                uint16  `json:"max_vals" type:"uint"`    // maximum number of validators
	BackupVals                             uint16  `json:"backup_vals" type:"uint"` // number of backup validators
	SelfStakingRatio                       sdk.Rat `json:"self_staking_ratio" type:"rat"`
	InflationRate                          sdk.Rat `json:"inflation_rate" type:"rat"`
	ValidatorSizeThreshold                 sdk.Rat `json:"validator_size_threshold" type:"rat"`
	UnstakeWaitingPeriod                   uint64  `json:"unstake_waiting_period" type:"uint"`
	ProposalExpirePeriod                   uint64  `json:"proposal_expire_period" type:"uint"`
	DeclareCandidacyGas                    uint64  `json:"declare_candidacy_gas" type:"uint"`
	UpdateCandidacyGas                     uint64  `json:"update_candidacy_gas" type:"uint"`
	SetCompRateGas                         uint64  `json:"set_comp_rate_gas" type:"uint"`
	UpdateCandidateAccountGas              uint64  `json:"update_candidate_account_gas" type:"uint"`
	AcceptCandidateAccountUpdateRequestGas uint64  `json:"accept_candidate_account_update_request_gas" type:"uint"`
	TransferFundProposalGas                uint64  `json:"transfer_fund_proposal_gas" type:"uint"`
	ChangeParamsProposalGas                uint64  `json:"change_params_proposal_gas" type:"uint"`
	DeployLibEniProposalGas                uint64  `json:"deploy_libeni_proposal_gas" type:"uint"`
	RetireProgramProposalGas               uint64  `json:"retire_program_proposal_gas" type:"uint"`
	UpgradeProgramProposalGas              uint64  `json:"upgrade_program_proposal_gas" type:"uint"`
	GasPrice                               uint64  `json:"gas_price" type:"uint"`
	MinStakingAmount                       int64   `json:"min_staking_amount" type:"uint"`
	ValidatorsBlockAwardRatio              sdk.Rat `json:"validators_block_award_ratio" type:"rat"`
	MaxSlashBlocks                         int16   `json:"max_slash_blocks" type:"uint"`
	SlashRatio                             sdk.Rat `json:"slash_ratio" type:"rat"`
	SlashEnabled                           bool    `json:"slash_enabled" type:"bool"`
	CubePubKeys                            string  `json:"cube_pub_keys" type:"json"`
	LowPriceTxGasLimit                     uint64  `json:"low_price_tx_gas_limit" type:"uint"`
	LowPriceTxSlotsCap                     int     `json:"low_price_tx_slots_cap" type:"int"`
	FoundationAddress                      string  `json:"foundation_address"`
	CalStakeInterval                       uint64  `json:"cal_stake_interval" type:"uint"`
	CalVPInterval                          uint64  `json:"cal_vp_interval" type:"uint"`
	CalAverageStakingDateInterval          uint64  `json:"cal_avg_staking_date_interval" type:"uint"`
}

func DefaultParams() *Params {
	return &Params{
		MaxVals:                                4,
		BackupVals:                             1,
		SelfStakingRatio:                       sdk.NewRat(10, 100),
		InflationRate:                          sdk.NewRat(8, 100),
		ValidatorSizeThreshold:                 sdk.NewRat(12, 100),
		UnstakeWaitingPeriod:                   7 * 24 * 3600 / CommitSeconds,
		ProposalExpirePeriod:                   7 * 24 * 3600 / CommitSeconds,
		DeclareCandidacyGas:                    1e6,   // gas setting for declareCandidacy
		UpdateCandidacyGas:                     1e6,   // gas setting for updateCandidacy
		SetCompRateGas:                         21000, // gas setting for setCompRate
		UpdateCandidateAccountGas:              1e6,   // gas setting for UpdateCandidateAccountGas
		AcceptCandidateAccountUpdateRequestGas: 1e6,   // gas setting for AcceptCandidateAccountUpdateRequestGas
		TransferFundProposalGas:                2e6,
		ChangeParamsProposalGas:                2e6,
		RetireProgramProposalGas:               2e6,
		UpgradeProgramProposalGas:              2e6,
		DeployLibEniProposalGas:                2e6,
		GasPrice:                               2e9,
		MinStakingAmount:                       1000,
		ValidatorsBlockAwardRatio:              sdk.NewRat(90, 100),
		MaxSlashBlocks:                         12,
		SlashRatio:                             sdk.NewRat(1, 1000),
		SlashEnabled:                           false,
		CubePubKeys:                            `[{"cube_batch":"01","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCiWpvDnwYFTqgSWPlA3VO8u+Yv\n9r8QGlRaYZFszUZEXUQxquGlFexMSVyFeqYjIokfPOEHHx2voqWgi3FKKlp6dkxw\nApP3T22y7Epqvtr+EfNybRta15snccZy47dY4UcmYxbGWFTaL66tz22pCAbjFrxY\n3IxaPPIjDX+FiXdJWwIDAQAB\n-----END PUBLIC KEY-----"},{"cube_batch":"02","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDQ8FL6/9zul+X7bFSRiWAzFiAE\n9vHYbClEHwlC7zUZ/JWzU7UT5S2qnYsseYF2WFjJtrGwHRAlTUyPtCpxV8f1uJsI\nl+/N9l6torUHwkhhib1catUSd/T72ltjvVyyg5LQjtRsskFnv3wM/yxYotrgnOs+\ndRpU6WI5XPCIyZqsGwIDAQAB\n-----END PUBLIC KEY-----"},{"cube_batch":"05","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCZ7Fw+1ddvy5OPFftbea0MxewW\nKUTb/E7B4/MHvLz2h7f7snyveFwxxj7QwxaCoVxobEq6AigIlUFUXLM8Y598/jts\nTaN+jh4xdoQN7qKwrbz1MWGf58Aa78Vnoj54B7V0LSajVbLJSZNUEI/24HLcG2iN\nTD3dSvH0ARvRJJ9hZQIDAQAB\n-----END PUBLIC KEY-----"},{"cube_batch":"06","pub_key":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCbNyWzuQ8Vrgrf9no9woaqgifc\njxfpvpuREoGNBOzOMl9BpyTa45t2ZeigE+xLaTZJc7dVTMQus8ik1b2qQcmrdViR\nbFx2P7tPg5z0DlDVXjq2G8Q3mP0WBEhGzyfycUmaT+yXoLu/UzGfFhr5nVztkUVD\noOHnTtsKCKQekuY3YwIDAQAB\n-----END PUBLIC KEY-----"}]`,
		LowPriceTxGasLimit:                     500000, // Maximum gas limit for low-price transaction
		LowPriceTxSlotsCap:                     100,    // Maximum number of low-price transaction slots per block
		FoundationAddress:                      "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
		CalStakeInterval:                       1, // calculate stake interval, default per block
		CalVPInterval:                          1, // calculate voting power interval, default per block
		CalAverageStakingDateInterval:          24 * 3600 / 10,
	}
}

var (
	// Keys for store prefixes
	ParamKey            = []byte{0x01} // key for global parameters
	AwardInfosKey       = []byte{0x02} // key for award infos
	AbsentValidatorsKey = []byte{0x03} // key for award infos
	dirty               = false
	params              = new(Params)
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
			case reflect.Bool:
				if iv, err := strconv.ParseBool(value); err == nil {
					fv.SetBool(iv)
				}
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
			case "bool":
				if _, err := strconv.ParseBool(value); err == nil {
					return true
				}
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
