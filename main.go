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
	"regexp"

	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

var (
	workingDir   string
	timeoutSec   int32
	mainRepo     string
	deployerEnv  []string
	watcher      IWatcher
	deployer     IDeployer
	extraRepos   []string
	argURL       string
	replacements map[string]string = map[string]string{}
	cmdRoot                        = &cobra.Command{
		Use: "cder watches over provided git repo or artifact and deploys it if changed",
	}
	cmdCDURL = &cobra.Command{
		Use:     "cdurl --url <url to watch over>",
		Short:   "Track lines (1st - artifact zip url, 2nd - deploy.sh url) of content from <url>. Something changed -> download all, unzip and run deploy.sh at unzipped dir",
		PreRunE: preRunCmdURL,
		RunE:    runCmdRoot,
	}
	cmdCDGit = &cobra.Command{
		Use:     "cd --repo <main-repo> [--extraRepo (<repo1-to-track>|<repo1-from=repo1-to>)[, (<repo2-to-track>|<repo2-from=repo2-to>)]...] [args]",
		Short:   "Pull and build sources from given git repo",
		Long:    "If <main-repo> or <repo-to-track> or <repo-to> is changed then <main-repo> will be build using appropriate deployer (deploy.sh if exists, `go build` otherwise). If main-repo is changed or have changed repo-to-track then main-repo will be build (deploy.sh if exists, golang builder otherwise).",
		PreRunE: preRunCDGit,
		RunE:    runCmdRoot,
	}
	initCmds []string
)

func main() {
	if err := execute(); err != nil {
		gc.Error(err)
		os.Exit(1)
	}
}

func execute() error {
	cmdRoot.PersistentFlags().BoolVarP(&gc.IsVerbose, "verbose", "v", false, "Verbose output")
	cmdRoot.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
	cmdRoot.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Seconds between pulls")
	cmdRoot.PersistentFlags().StringSliceVar(&deployerEnv, "deployer-env", []string{}, "Deployer environment variable")
	cmdRoot.PersistentFlags().StringSliceVar(&initCmds, "init", []string{}, "Any commands to be executed before start")
	cmdRoot.MarkPersistentFlagRequired("repo")
	cmdRoot.AddCommand(cmdCDGit)
	cmdRoot.AddCommand(cmdCDURL)

	cmdCDGit.PersistentFlags().StringSliceVar(&extraRepos, "extraRepo", []string{}, "Dependencies of main repository to track for changes")
	cmdCDGit.PersistentFlags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
	cmdCDGit.PersistentFlags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
	cmdCDGit.MarkPersistentFlagRequired("output")
	cmdCDGit.MarkPersistentFlagRequired("repo")

	cmdCDURL.PersistentFlags().StringVarP(&argURL, "url", "u", "", "URL to download artifact state from")

	return cmdRoot.Execute()
}

func preRunCDGit(cmd *cobra.Command, args []string) error {
	watcher = &watcherGit{
		lastCommitHashes: map[string]string{},
	}
	repos = []string{mainRepo}

	// *************************************************
	gc.Doing("Calculating parameters")
	re := regexp.MustCompile(`([^=]*)(=(.*))*`)
	for _, extraRep := range extraRepos {
		matches := re.FindStringSubmatch(extraRep)
		if matches == nil {
			continue
		}
		if len(matches[2]) == 0 {
			replacements[matches[1]] = matches[1]
			repos = append(repos, matches[1])
		} else {
			replacements[matches[1]] = matches[3]
			repos = append(repos, matches[3])
		}
	}

	// *************************************************
	gc.Doing("Configuring deployer")
	deployerPath := path.Join(workingDir, "deploy.sh")
	if _, err := os.Stat(deployerPath); err == nil {
		gc.Info("Custom deployer will be used: " + deployerPath)
		deployer = &deployer4sh{
			wd: workingDir,
		}
	} else {
		gc.Info("Standart go deployer will be used")
		repoPath, _ := getAbsRepoFolders(mainRepo)
		deployer = &deployer4go{
			wd:   repoPath,
			args: args,
		}
	}
	return nil
}

func preRunCmdURL(cmd *cobra.Command, args []string) error {
	watcher = &watcherURL{}
	repos = []string{argURL}
	return nil
}
