package handler

import (
	"github.com/tendermint/go-wire/data"
	"github.com/cosmos/cosmos-sdk"
	"github.com/pkg/errors"
)

type nonce int64

type Actor struct {
	ChainID string     `json:"chain"` // this is empty unless it comes from a different chain
	App     string     `json:"app"`   // the app that the actor belongs to
	Address data.Bytes `json:"addr"`  // arbitrary app-specific unique id
}

type Context struct {
	app string
	ibc bool
	id     nonce
	chain  string
	height int64
	perms  []sdk.Actor
}

func (c Context) ChainID() string {
	return c.chain
}

func (c Context) BlockHeight() int64 {
	return c.height
}

// WithPermissions will panic if they try to set permission without the proper app
func (c Context) WithPermissions(perms ...sdk.Actor) Context {
	// the guard makes sure you only set permissions for the app you are inside
	for _, p := range perms {
		if !c.validPermission(p) {
			err := errors.Errorf("Cannot set permission for %s/%s on (app=%s, ibc=%b)",
				p.ChainID, p.App, c.app, c.ibc)
			panic(err)
		}
	}

	return Context{
		app:    c.app,
		ibc:    c.ibc,
		id:     c.id,
		chain:  c.chain,
		height: c.height,
		perms:  append(c.perms, perms...),
	}
}

func (c Context) validPermission(p sdk.Actor) bool {
	// if app is set, then it must match
	if c.app != "" && c.app != p.App {
		return false
	}
	// if ibc, chain must be set, otherwise it must not
	return c.ibc == (p.ChainID != "")
}

// Reset should clear out all permissions,
// but carry on knowledge that this is a child
func (c Context) Reset() Context {
	return Context{
		app:    c.app,
		ibc:    c.ibc,
		id:     c.id,
		chain:  c.chain,
		height: c.height,
	}
}