package types

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// TravisCmd ...
type TravisCmd struct {
	Root     string
	Path     string
	Name     string
	Args     []string
	NextName string
	NextArgs []string
	// Env  []string
	*sync.Mutex
	DownloadChan chan *CmdInfo   //
	UpgradeChan  chan *CmdInfo //
	KillChan     chan string   //
	started      bool          // cmd.Start called, no error
	stopped      bool          // Stop called
	downloaded   bool          // donwload successfully
	startTime    time.Time     // if started true
	cmd          *exec.Cmd
}

// NewTravisCmd create a new travis CMD
func NewTravisCmd(root string, name string, arg ...string) *TravisCmd {
	return &TravisCmd{
		Root:         root,
		Path:         filepath.Join(root, "bin"),
		Name:         name,
		Args:         arg,
		Mutex:        &sync.Mutex{},
		DownloadChan: make(chan *CmdInfo, 1),
		UpgradeChan:  make(chan *CmdInfo, 1),
		KillChan:     make(chan string, 1),
	}
}

// Start stop the sub travis process
func (c *TravisCmd) Start() error {
	var stdoutBuf, stderrBuf bytes.Buffer
	fullName := filepath.Join(c.Path, c.Name)
	cmd := exec.Command(fullName, c.Args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	c.started = true
	c.startTime = time.Now()
	c.cmd = cmd
	c.downloaded = false
	c.NextName = ""
	c.NextArgs = nil

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	go func() {
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("cmd.Run() failed with %s\n", err)
		}
		if errStdout != nil || errStderr != nil {
			fmt.Printf("failed to capture stdout or stderr\n")
		}
	}()
	return nil
}

// Stop the sub travis process
func (c *TravisCmd) Stop() error {
	pro, err := os.FindProcess(c.cmd.Process.Pid)
	if err != nil {
		fmt.Printf("can not find rpocess:%d\n", c.cmd.Process.Pid)
		return err
	}
	err = pro.Signal(syscall.SIGINT)
	if err != nil {
		fmt.Printf("failed to terminate sub-process: %s\n", c.cmd.Path)
		return err
	}
	fmt.Printf("terminate sub-process sucessfully: %s\n", c.cmd.Path)
	c.started = false
	c.startTime = time.Time{}
	c.cmd = nil
	c.stopped = true
	return nil
}

// Restart restart the travis
func (c *TravisCmd) Restart() error {
	c.Lock()
	defer c.Unlock()
	// stop the old
	if err := c.Stop(); err != nil {
		return err
	}
	c.Start()
	return nil
}

// Upgrade upgrade to new version travis
func (c *TravisCmd) Upgrade(cmdInfo *CmdInfo) error {
	c.Lock()
	defer c.Unlock()
	// TODO: need sleep a while to wait something finish ?
	time.Sleep(time.Second * 1)

	// stop the old
	if err := c.Stop(); err != nil {
		return err
	}

	if !c.downloaded || c.NextName == "" {
		return errors.New("no new version travis get ready")
	}

	// using the new version
	c.Name = c.NextName

	c.Start()
	return nil
}

// Download download the new version travis as specified
func (c *TravisCmd) Download(cmdInfo *CmdInfo) error {
	c.Lock()
	defer c.Unlock()

	if c.downloaded && c.NextName == cmdInfo.ReleaseName() {
		log.Println("same version already exist")
		return nil
	}

	// TODO: automatically download comming soon ...
	//if err := exec.Command("cp", "-f", filepath.Join(c.Path, "travis"), filepath.Join(c.Path, name)).Run(); err != nil {
	//	return err
	//}
	log.Println("download does not happen automatically, please copy it manually")

	// using the new version
	c.NextName = cmdInfo.ReleaseName()
	c.downloaded = true

	return nil
}

// Cmd ...
func (c *TravisCmd) Cmd() *exec.Cmd {
	return c.cmd
}

// Kill kill the travis command
func (c *TravisCmd) Kill() error {
	c.Lock()
	defer c.Unlock()
	// need sleep a while to wait something finish
	time.Sleep(time.Second * 1)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(syscall.SIGTERM)
}

