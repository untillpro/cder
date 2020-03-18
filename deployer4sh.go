/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"os"
	"path"

	gc "github.com/untillpro/gochips"
)

type deployer4sh struct {
	wd string
}

func (d *deployer4sh) Deploy(repo string) {
	d.execCommand("deploy", nil, true)
}

func (d *deployer4sh) DeployAll(repos []string) {
	d.execCommand("deploy-all", nil, true)
}

func (d *deployer4sh) Stop() {
	d.execCommand("stop", nil, false)
}

func (d *deployer4sh) execCommand(command string, commandArgs []string, panicOnError bool) (err error) {
	var args []string
	args = append(args, deployerEnv...)
	args = append(args, path.Join(d.wd, "deploy.sh"), command)
	args = append(args, commandArgs...)
	err = new(gc.PipedExec).
		Command("env", args...).
		WorkingDir(d.wd).
		Run(os.Stdout, os.Stderr)
	if panicOnError {
		gc.PanicIfError(err)
	}
	return err
}
