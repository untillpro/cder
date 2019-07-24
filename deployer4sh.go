package main

import (
	"os"

	gc "github.com/untillpro/gochips"
)

type deployer4sh struct {
	// Params
	deployerPath string
	repos        []string
	args         []string
	forks        map[string]string
}

func (d *deployer4sh) Start() (err error) {
	return d.execCommand("start", nil, false)
}

func (d *deployer4sh) Stop() {
	d.execCommand("stop", nil, false)
}

func (d *deployer4sh) DeployAll(repoPaths []string) {
	d.execCommand("deploy-all", repoPaths, true)
}

func (d *deployer4sh) Deploy(repoPath string) {
	d.execCommand("deploy", []string{repoPath}, true)

}

func (d *deployer4sh) execCommand(command string, commandArgs []string, panicOnError bool) (err error) {
	var args []string
	args = append(args, deployerEnv...)
	args = append(args, "./deployer.sh", command)
	args = append(args, commandArgs...)
	err = new(gc.PipedExec).
		Command("env", args...).
		WorkingDir(workingDir).
		Run(os.Stdout, os.Stderr)
	if panicOnError {
		gc.PanicIfError(err)
	}
	return err
}
