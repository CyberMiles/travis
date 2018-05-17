package types

// GenesisValidator is an initial validator.
type GenesisValidator struct {
	PubKey    PubKey        `json:"pub_key"`
	Power     int64         `json:"power"`
	Name      string        `json:"name"`
	Address   string        `json:"address"`
	Cut       string        `json:"cut"`
	MaxAmount int64         `json:"max_amount"`
}

