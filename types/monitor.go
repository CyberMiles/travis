package types

import (
	"errors"
)

// CmdInfo ...
type CmdInfo struct {
	Name string
	Version  string
	DownloadURLs []string
	MD5 string
}

// Monitor ...
type Monitor struct {
	cmd *TravisCmd
}

// NewMonitor ...
func NewMonitor(cmd *TravisCmd) *Monitor {
	return &Monitor{cmd: cmd}
}

// MonitorResponse ...
type MonitorResponse struct {
	Code uint
	Msg  []byte
}

// Download ...
func (r *Monitor) Download(info *CmdInfo, reply *MonitorResponse) error {
	if info == nil || info.Name == "" {
		return errors.New("CmdInfo can't be nil")
	}
	reply.Code = 0
	reply.Msg = []byte("Received download info successfully")
	r.cmd.DownloadChan <- info
	return nil
}

// Upgrade ...
func (r *Monitor) Upgrade(info *CmdInfo, reply *MonitorResponse) error {
	if info == nil || info.Name == "" {
		return errors.New("CmdInfo can't be nil")
	}
	reply.Code = 0
	reply.Msg = []byte("Received upgrade info successfully")
	r.cmd.UpgradeChan <- info
	return nil
}

// Kill ...
func (r *Monitor) Kill(info *CmdInfo, reply *MonitorResponse) error {
	reply.Code = 0
	reply.Msg = []byte("Received kill info successfully")
	r.cmd.KillChan <- ""
	return nil
}

// ReleaseName get the travis release name
func (c *CmdInfo)ReleaseName() string {
	return c.Name + "_" + c.Version
}
