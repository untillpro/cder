package main

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	gc "github.com/untillpro/gochips"
)

var binaryName string

type deployer4go struct {
	// Params
	repos []string
	args  []string
	forks map[string]string

	// Internal
	cmd *exec.Cmd
}

func (d *deployer4go) Start() (err error) {
	if len(binaryName) == 0 {
		return errors.New("Output binary name is a must, use `-o` flag")
	}
	return nil
}

func (d *deployer4go) Stop() {
	d.stopCmd()
}

func (d *deployer4go) DeployAll(repoPaths []string) {
	var err error

	d.replaceGoMod()

	gc.Info("itdeployer4go.DeployAll:", "Main repo will be rebuilt")
	repoPath, _ := getAbsRepoFolders(d.repos[0])

	gc.Doing("go build")
	err = new(gc.PipedExec).
		Command("go", "build", "-o", binaryName).
		WorkingDir(repoPath).
		Run(verboseWriters())
	gc.PanicIfError(err)
	gc.Info("deployer4go.DeployAll:", "Build finished")

	// Stop and replace executable

	fileToExec, err := filepath.Abs(path.Join(workingDir, binaryName))
	gc.PanicIfError(err)

	d.stopCmd()

	gc.Doing("Moving " + binaryName + " to " + fileToExec)
	err = new(gc.PipedExec).
		Command("mv", binaryName, fileToExec).
		WorkingDir(repoPath).
		Run(os.Stdout, os.Stderr)
	gc.PanicIfError(err)

	// Run executable

	gc.Doing("deployer4go.DeployAll: Running " + fileToExec)

	pe := new(gc.PipedExec)
	err = pe.Command(fileToExec, d.args...).
		WorkingDir(repoPath).
		Start(os.Stdout, os.Stderr)
	gc.PanicIfError(err)

	d.cmd = pe.GetCmd(0)
	gc.Info("deployer4go.DeployAll:", "Process started!")
}

func (d *deployer4go) Deploy(repoPath string) {

}

func (d *deployer4go) stopCmd() {
	defer func() { d.cmd = nil }()
	if nil != d.cmd {
		gc.Doing("deployer4go.stopCmd: Terminating  previous process")
		err := d.cmd.Process.Kill()
		if nil != err {
			gc.Error("deployer4go.stopCmd: Error occured:", err)
		}
		gc.Info("deployer4go.stopCmd: Done")
	}
}

func (d *deployer4go) replaceGoMod() {

	mainRepoPath, _ := getAbsRepoFolders(d.repos[0])
	goModPath := path.Join(mainRepoPath, "go.mod")

	goModPathContentBytes, err := ioutil.ReadFile(goModPath)
	goModPathContent := string(goModPathContentBytes)
	gc.PanicIfError(err)

	for forkedfrom, forkedTo := range d.forks {
		_, toFolder := getAbsRepoFolders(forkedTo)

		u, err := url.Parse(forkedfrom)
		gc.PanicIfError(err)

		replace := "replace " + u.Hostname() + u.RequestURI() + " => " + path.Join("..", toFolder)
		gc.Info("deployer4go.replaceGoMod", "replace", replace)
		goModPathContent = goModPathContent + replace + "\n"
	}

	gc.Doing("deployer4go.replaceGoMod: Replacing go.mod")
	err = ioutil.WriteFile(goModPath, []byte(goModPathContent), 0644)
	gc.PanicIfError(err)

}
