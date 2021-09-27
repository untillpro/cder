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
		Use:     "cd --repo <main-repo> [--extraRepo (<repo1-to-track>|<repo1-from=repo1-to>)[, (<repo2-to-track>|<repo2-from=repo2-to>)]...] [--gitDirect] [args]",
		Short:   "Build sources from given git repo. Pulls repositories each `--timeout` seconds to know if something changed",
		Long:    "If <main-repo> or <repo-to-track> or <repo-to> is changed then <main-repo> will be build using appropriate deployer (deploy.sh if exists, `go build` otherwise). If main-repo is changed or have changed repo-to-track then main-repo will be build (deploy.sh if exists, golang builder otherwise). `--gitDirect` specified -> git repo pull will be used instead of `gotify`",
		PreRunE: preRunCDGit,
		RunE:    runCmdRoot,
	}
	cmdCDGotify = &cobra.Command{
		Use:     "cdGotify --repo <main-repo> [--extraRepo (<repo1-to-track>|<repo1-from=repo1-to>)[, (<repo2-to-track>|<repo2-from=repo2-to>)]...] --url <gotify url> --token <gotify token> --app <gotify app> [args]",
		Short:   "Build sources from given git repo. Queries Gotify server each `--timeout` seconds to know if something changed",
		Long:    "Last commit hashes are received as notifications from Gotify server. <main-repo> will be build using appropriate deployer (deploy.sh if exists, `go build` otherwise). If main-repo is changed or have changed repo-to-track then main-repo will be build (deploy.sh if exists, golang builder otherwise). `--gitDirect` specified -> git repo pull will be used instead of `gotify`",
		PreRunE: preRunCDGotify,
		RunE:    runCmdRoot,
	}
	initCmds []string
)

func main() {
	gc.ExitIfError(execute())
}

func execute() error {
	cmdRoot.PersistentFlags().BoolVarP(&gc.IsVerbose, "verbose", "v", false, "Verbose output")
	cmdRoot.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
	cmdRoot.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Seconds between pulls")
	cmdRoot.PersistentFlags().StringSliceVar(&deployerEnv, "deployer-env", []string{}, "Deployer environment variable")
	cmdRoot.PersistentFlags().StringSliceVar(&initCmds, "init", []string{}, "Any commands to be executed before start. Could be separated with `;`")
	cmdRoot.AddCommand(cmdCDGit)
	cmdRoot.AddCommand(cmdCDURL)
	cmdRoot.AddCommand(cmdCDGotify)

	cmdCDGit.Flags().StringSliceVar(&extraRepos, "extraRepo", []string{}, "Dependencies of main repository to track for changes")
	cmdCDGit.Flags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
	cmdCDGit.Flags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
	cmdCDGit.Flags().StringVarP(&buildPath, "build", "b", "", "Path to build at")
	cmdCDGit.MarkFlagRequired("output")
	cmdCDGit.MarkFlagRequired("repo")


	cmdCDGotify.Flags().StringSliceVar(&extraRepos, "extraRepo", []string{}, "Dependencies of main repository to track for changes")
	cmdCDGotify.Flags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
	cmdCDGotify.Flags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
	cmdCDGotify.Flags().StringVarP(&buildPath, "build", "b", "", "Path to build at")
	cmdCDGotify.Flags().StringVarP(&gToken, "token", "", "", "Gotify token")
	cmdCDGotify.Flags().StringVarP(&gURL, "url", "u", "", "Gotify server url")
	cmdCDGotify.MarkFlagRequired("output")
	cmdCDGotify.MarkFlagRequired("repo")
	cmdCDGotify.MarkFlagRequired("app")
	cmdCDGotify.MarkFlagRequired("token")
	cmdCDGotify.MarkFlagRequired("url")

	cmdCDURL.Flags().StringVarP(&argURL, "url", "u", "", "URL to download artifact state from")
	cmdCDURL.MarkFlagRequired("url")

	return cmdRoot.Execute()
}

func prepareGitRepos(args []string) {
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
			repoURLs = append(repoURLs, matches[1])
		} else {
			replacements[matches[1]] = matches[3]
			repoURLs = append(repoURLs, matches[3])
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
}

func preRunCDGit(cmd *cobra.Command, args []string) error {
	watcher = &watcherGit{
		lastCommitHashes: map[string]string{},
		commitsTracker: &gitTrackerPull{},
	}
	repoURLs = []string{mainRepo}
	prepareGitRepos(args)
	return nil
}

func preRunCmdURL(cmd *cobra.Command, args []string) error {
	watcher = &watcherURL{}
	repoURLs = []string{argURL}
	return nil
}

func preRunCDGotify(cmd *cobra.Command, args []string) error {
	watcher = &watcherGit{
		lastCommitHashes: map[string]string{},
		commitsTracker: &gitTrackerGotify{},
	}
	repoURLs = []string{mainRepo}
	prepareGitRepos(args)
	return nil
}
