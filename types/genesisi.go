package types

import (
	"encoding/json"
	"github.com/CyberMiles/travis/sdk"
	"github.com/pkg/errors"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/types"
	"io/ioutil"
	"strconv"
	"time"
)

//------------------------------------------------------------
// core types for a genesis definition

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time              `json:"genesis_time"`
	ChainID         string                 `json:"chain_id"`
	ConsensusParams *types.ConsensusParams `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator     `json:"validators"`
	AppHash         []byte                 `json:"app_hash"`
}

// GenesisValidator is an initial validator.
type GenesisValidator struct {
	PubKey    PubKey  `json:"pub_key"`
	Power     string  `json:"power"`
	Shares    int64   `json:"shares"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	CompRate  sdk.Rat `json:"comp_rate"`
	MaxAmount int64   `json:"max_amount"`
	Website   string  `json:"website"`
	Location  string  `json:"location"`
	Email     string  `json:"email"`
	Profile   string  `json:profile`
}

// SaveAs is a utility method for saving GenensisDoc as a JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes, err := json.Marshal(genDoc)
	if err != nil {
		return err
	}
	return cmn.WriteFile(file, genDocBytes, 0644)
}

// ValidatorHash returns the hash of the validator set contained in the GenesisDoc
func (genDoc *GenesisDoc) ValidatorHash() []byte {
	vals := make([]*types.Validator, len(genDoc.Validators))
	for i, v := range genDoc.Validators {
		power, _ := strconv.ParseInt(v.Power, 10, 64)
		vals[i] = types.NewValidator(v.PubKey, power)
	}
	vset := types.NewValidatorSet(vals)
	return vset.Hash()
}

// ValidateAndComplete checks that all necessary fields are present
// and fills in defaults for optional fields left empty
func (genDoc *GenesisDoc) ValidateAndComplete() error {

	if genDoc.ChainID == "" {
		return errors.Errorf("Genesis doc must include non-empty chain_id")
	}

	if genDoc.ConsensusParams == nil {
		genDoc.ConsensusParams = types.DefaultConsensusParams()
	} else {
		if err := genDoc.ConsensusParams.Validate(); err != nil {
			return err
		}
	}

	if len(genDoc.Validators) == 0 {
		return errors.Errorf("The genesis file must have at least one validator")
	}

	for _, v := range genDoc.Validators {
		if v.Power == "0" {
			return errors.Errorf("The genesis file cannot contain validators with no voting power: %v", v)
		}
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = time.Now()
	}

	return nil
}

//------------------------------------------------------------
// Make genesis state from file

// GenesisDocFromJSON unmarshalls JSON data into a GenesisDoc.
func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	genDoc := GenesisDoc{}
	err := json.Unmarshal(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return &genDoc, err
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't read GenesisDoc file")
	}
	genDoc, err := GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, errors.Wrap(err, cmn.Fmt("Error reading GenesisDoc at %v", genDocFile))
	}
	return genDoc, nil
}
