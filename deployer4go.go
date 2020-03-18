/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	gc "github.com/untillpro/gochips"
)

var binaryName string

type deployer4go struct {
	cmd  *exec.Cmd
	wd   string
	args []string
}

func (d *deployer4go) Stop() {
	d.stopCmd()
}

func (d *deployer4go) Deploy(repo string) {
}

func (d *deployer4go) DeployAll(repos []string) {
	// replace go.mod and build
	d.replaceGoMod()
	gc.Info("itdeployer4go.DeployAll:", "Main repo will be rebuilt")
	gc.Doing("go build")
	err := new(gc.PipedExec).
		Command("go", "build", "-o", binaryName).
		WorkingDir(d.wd).
		Run(gc.VerboseWriters())
	gc.PanicIfError(err)
	gc.Info("deployer4go.DeployAll:", "Build finished")

	// Stop and replace executable
	fileToExec, err := filepath.Abs(path.Join(workingDir, binaryName))
	gc.PanicIfError(err)
	d.stopCmd()
	gc.Doing("Moving " + binaryName + " to " + fileToExec)
	err = new(gc.PipedExec).
		Command("mv", binaryName, fileToExec).
		WorkingDir(d.wd).
		Run(os.Stdout, os.Stderr)
	gc.PanicIfError(err)

	// Run executable
	gc.Doing("deployer4go.DeployAll: Running " + fileToExec)
	pe := new(gc.PipedExec)
	err = pe.Command(fileToExec, d.args...).
		WorkingDir(d.wd).
		Start(os.Stdout, os.Stderr)
	gc.PanicIfError(err)
	d.cmd = pe.GetCmd(0)
	gc.Info("deployer4go.DeployAll:", "Process started!")
}

func withTimeout(f func()) bool {
	timeout := time.After(30 * time.Second)
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-timeout:
		return false
	case <-done:
		return true
	}
}

func (d *deployer4go) stopCmd() {
	defer func() { d.cmd = nil }()
	if nil != d.cmd {
		gc.Doing("deployer4go.stopCmd: sending SIGINT to the child process")
		if err := d.cmd.Process.Signal(os.Interrupt); err != nil {
			gc.Error("deployer4go.stopCmd: sending SIGINT error:", err)
		}
		gc.Doing("deployer4go.stopCmd: waiting the process to finish")
		var pc *os.ProcessState
		var err error
		if withTimeout(func() {
			pc, err = d.cmd.Process.Wait()
		}) {
			if err != nil {
				gc.Error("deployer4go.stopCmd: awaiting for the process to shutdown error:", err)
			} else {
				if pc.ExitCode() != 0 {
					gc.Error("deployer4go.stopCmd: process exit code:", pc.ExitCode())
				}
			}
		} else {
			gc.Doing("deployer4go.stopCmd: Timeout. Killing...")
			if err := d.cmd.Process.Kill(); err != nil {
				gc.Error("deployer4go.stopCmd: killing:", err)
			}
		}
		gc.Info("deployer4go.stopCmd: Done")
	}
}

func (d *deployer4go) replaceGoMod() {
	gc.Doing("deployer4go.replaceGoMod: Replacing go.mod")

	goModPath := path.Join(d.wd, "go.mod")
	goModPathContentBytes, err := ioutil.ReadFile(goModPath)
	gc.PanicIfError(err)
	goModPathContent := string(goModPathContentBytes)

	for repFrom, repTo := range replacements {
		_, toFolder := getAbsRepoFolders(repTo)

		u, err := url.Parse(repFrom)
		gc.PanicIfError(err)

		replace := "replace " + u.Hostname() + u.RequestURI() + " => " + path.Join("..", toFolder)
		gc.Info("deployer4go.replaceGoMod", replace)
		goModPathContent = goModPathContent + replace + "\n"
	}

	gc.PanicIfError(ioutil.WriteFile(goModPath, []byte(goModPathContent), 0644))
}
