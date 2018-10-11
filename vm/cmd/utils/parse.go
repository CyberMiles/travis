package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/ethereum/go-ethereum/core"

	gen "github.com/CyberMiles/travis/misc/genesis"
	"github.com/CyberMiles/travis/utils"
)

var blankGenesis = new(core.Genesis)

var errBlankGenesis = errors.New("could not parse a valid/non-blank Genesis")

// ParseGenesisOrDefault tries to read the content from provided
// genesisPath. If the path is empty or doesn't exist, it will
// use defaultGenesisBytes as the fallback genesis source. Otherwise,
// it will open that path and if it encounters an error that doesn't
// satisfy os.IsNotExist, it returns that error.
func ParseGenesisOrDefault(genesisPath string, chainID uint) (*core.Genesis, error) {
	var genesisBlob []byte
	var err error
	if chainID == utils.MainNet {
		genesisBlob, err = gen.DefaultGenesisBlock().MarshalJSON();
	} else {
		genesisBlob, err = gen.DevGenesisBlock().MarshalJSON()
	}

	if err != nil {
		return nil, err
	}
	if len(genesisPath) > 0 {
		blob, err := ioutil.ReadFile(genesisPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if len(blob) >= 2 { // Expecting atleast "{}"
			genesisBlob = blob
		}
	}

	genesis := new(core.Genesis)
	if err := json.Unmarshal(genesisBlob, genesis); err != nil {
		return nil, err
	}

	if reflect.DeepEqual(blankGenesis, genesis) {
		return nil, errBlankGenesis
	}

	return genesis, nil
}
