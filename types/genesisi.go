package types

// GenesisValidator is an initial validator.
type GenesisValidator struct {
	PubKey    PubKey `json:"pub_key"`
	Power     string `json:"power"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	CompRate  string `json:"comp_rate"`
	MaxAmount int64  `json:"max_amount"`
	Website   string `json:"website"`
	Location  string `json:"location"`
	Email     string `json:"email"`
	Profile   string `json:profile`
}

type GenesisCubePubKey struct {
	CubeBatch string `json:"cube_batch"`
	PubKey    string `json:"pub_key"`
}
