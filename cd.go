package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

var workingDir string
var timeoutSec int32
var mainRepo string
var argReplaced []string
var deployerEnv []string

type cdman struct {
	ctx              context.Context
	repos            []string
	forks            map[string]string
	timeout          time.Duration
	lastCommitHashes map[string]string
	cmd              *exec.Cmd
	args             []string
	// Used in the beginning of iteration to execute `git reset --hard`
	reposMustBeCleaned bool
}

func verboseWriters() (out io.Writer, err io.Writer) {
	if gc.IsVerbose {
		return os.Stdout, os.Stderr
	}
	return ioutil.Discard, os.Stderr
}

func getLastCommitHash(repoDir string) string {
	stdout, _, err := new(gc.PipedExec).
		Command("git", "log", "-n", "1", `--pretty=format:%H`).
		WorkingDir(repoDir).
		RunToStrings()
	gc.PanicIfError(err)

	return strings.TrimSpace(stdout)
}

func (p *cdman) cycle(wg *sync.WaitGroup, d IDeployer) {
	defer wg.Done()

	gc.Info("Seeding started")
	gc.Info("repos", p.repos)
	gc.Info("forks", p.forks)
	gc.Info("timeout", p.timeout)

	// *************************************************

F:
	for {
		p.iteration(d)
		select {
		case <-time.After(p.timeout):
		case <-p.ctx.Done():
			d.Stop()
			gc.Verbose("cdman", "Done")
			break F
		}
	}
	gc.Info("Seeding ended")
}

func (p *cdman) iteration(d IDeployer) {
	defer func() {
		if r := recover(); r != nil {
			gc.Error("iteration: Recovered: ", r)
		}
	}()

	// *************************************************
	gc.Verbose("iteration:", "Checking if repos should be cloned")

	var changedReposPaths []string

	for _, curRepoURL := range p.repos {

		reposFolder := getReposFolder()
		repoPath, repoFolder := getAbsRepoFolders(curRepoURL)

		gc.Verbose("iteration:", "repoPath, repoFolder=", repoPath, repoFolder)

		os.MkdirAll(reposFolder, 0755)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			gc.Info("iteration:", "Repo folder does not exist, will be cloned", repoPath, curRepoURL)
			err := new(gc.PipedExec).
				Command("git", "clone", "--recurse-submodules", curRepoURL).
				WorkingDir(reposFolder).
				Run(os.Stdout, os.Stderr)
			gc.PanicIfError(err)
		} else {

			if p.reposMustBeCleaned {
				gc.Info("iteration:", "Cleaning "+repoPath)
				err = new(gc.PipedExec).
					Command("git", "reset", "--hard").
					WorkingDir(repoPath).
					Run(os.Stdout, os.Stderr)
				gc.PanicIfError(err)
			}

			gc.Verbose("iteration:", "Repo dir exists, will be pulled", repoPath, curRepoURL)
			stdouts, stderrs, err := new(gc.PipedExec).
				Command("git", "pull", curRepoURL).
				WorkingDir(repoPath).
				RunToStrings()
			if nil != err {
				gc.Info(stdouts, stderrs)
			}
			gc.PanicIfError(err)
		}

		newHash := getLastCommitHash(repoPath)
		oldHash := p.lastCommitHashes[repoPath]
		if oldHash == newHash {
			gc.Verbose("*** Nothing changed")
		} else {
			p.reposMustBeCleaned = true
			gc.Info("iteration:", "Commit hash changed", oldHash, newHash)
			if len(oldHash) > 0 {
				gitModulesPath := path.Join(repoPath, ".gitmodules")
				if _, err := os.Stat(gitModulesPath); err == nil {
					gc.Doing("iteration: updating modules")
					err = new(gc.PipedExec).
						Command("git", "submodule", "update").
						WorkingDir(repoPath).
						Run(os.Stdout, os.Stderr)
				}
			}

			d.Deploy(repoPath)

			p.lastCommitHashes[repoPath] = newHash
			changedReposPaths = append(changedReposPaths, repoPath)
		}
	} // for repors

	if len(changedReposPaths) > 0 {
		d.DeployAll(changedReposPaths)
	} else {
		p.reposMustBeCleaned = false
	}
}

func runCmdCD(cmd *cobra.Command, args []string) {

	retCode := 0
	defer os.Exit(retCode)

	// *************************************************
	gc.Doing("Calculating parameters")

	re := regexp.MustCompile(`([^=]*)=(.*)`)
	repos := []string{mainRepo}
	forks := make(map[string]string)
	for _, arg := range argReplaced {
		matches := re.FindStringSubmatch(arg)
		if matches == nil {
			retCode = 1
			gc.Error(`Wrong replaced repo specification, must be <repo>=<repo-to-replace>:`, arg)
			return
		}
		gc.Verbose("arg", arg)
		gc.Verbose("matches", matches)
		forks[matches[1]] = matches[2]
		repos = append(repos, matches[2])
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// *************************************************
	gc.Doing("Configuring deployer")

	var d IDeployer

	deployerPath := path.Join(workingDir, "deployer.sh")
	if _, err := os.Stat(deployerPath); err == nil {
		gc.Info("Custom deployer will be used: " + deployerPath)
		d = &deployer4sh{
			deployerPath: deployerPath,
			repos:        repos,
			forks:        forks,
			args:         args,
		}
	} else {
		gc.Info("Standart go deployer will be used")
		d = &deployer4go{
			repos: repos,
			forks: forks,
			args:  args,
		}
	}

	// *************************************************
	gc.Doing("Starting deployer")
	err := d.Start()
	defer d.Stop()
	if nil != err {
		gc.Error("Failed to initialize deployer: ", err)
		retCode = 1
		return
	}

	// *************************************************
	gc.Doing("Starting seeding")
	ctx, cancel := context.WithCancel(context.Background())
	cdman := &cdman{ctx: ctx,
		repos:            repos,
		forks:            forks,
		timeout:          time.Duration(timeoutSec) * time.Second,
		lastCommitHashes: map[string]string{},
		args:             args,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go cdman.cycle(&wg, d)

	go func() {
		signal := <-signals
		fmt.Println("Got signal:", signal)
		cancel()
	}()

	// *************************************************
	wg.Wait()

	gc.Info("Finished")
}
