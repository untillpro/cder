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
	"strings"

	gc "github.com/untillpro/gochips"
)

type watcherGit struct {
	lastCommitHashes map[string]string
	// Used in the beginning of iteration to execute `git reset --hard`
	reposMustBeCleaned bool
}

func (w *watcherGit) Watch(repos []string) (changedRepos []string) {
	defer func() {
		if r := recover(); r != nil {
			gc.Error("watcherGit: Recovered: ", r)
		}
	}()

	// *************************************************
	reposFolder := getReposFolder()

	for _, repo := range repos {
		repoPath, repoFolder := getAbsRepoFolders(repo)
	
		gc.Verbose("watcherGit", "repoPath, repoFolder=", repoPath, repoFolder)
	
		gc.ExitIfError(os.MkdirAll(reposFolder, 0755))
	
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			gc.Info("watcherGit", "Repo folder does not exist, will be cloned", repoPath, repo)
			err := new(gc.PipedExec).
				Command("git", "clone", "--recurse-submodules", repo).
				WorkingDir(reposFolder).
				Run(os.Stdout, os.Stderr)
			gc.PanicIfError(err)
		} else {
			if w.reposMustBeCleaned {
				gc.Info("watcherGit", "Cleaning "+repoPath)
				err = new(gc.PipedExec).
					Command("git", "reset", "--hard").
					WorkingDir(repoPath).
					Run(os.Stdout, os.Stderr)
				gc.PanicIfError(err)
			}
	
			gc.Verbose("watcherGit", "Repo dir exists, will be pulled", repoPath, repo)
			stdouts, stderrs, err := new(gc.PipedExec).
				Command("git", "pull", repo).
				WorkingDir(repoPath).
				RunToStrings()
			if nil != err {
				gc.Info(stdouts, stderrs)
			}
			gc.PanicIfError(err)
		}
	
		newHash := getLastCommitHash(repoPath)
		oldHash := w.lastCommitHashes[repoPath]
		if oldHash == newHash {
			continue
		}
		w.reposMustBeCleaned = true
		gc.Info("watcherGit", "Commit hash changed", oldHash, newHash)
		if len(oldHash) > 0 {
			gitModulesPath := path.Join(repoPath, ".gitmodules")
			if _, err := os.Stat(gitModulesPath); err == nil {
				gc.Doing("watcherGit: updating modules")
				err = new(gc.PipedExec).
				Command("git", "submodule", "update").
				WorkingDir(repoPath).
				Run(os.Stdout, os.Stderr)
			}
		}
		w.lastCommitHashes[repoPath] = newHash
		changedRepos = append(changedRepos, repoPath)
	}

	return 
}

func getLastCommitHash(repoDir string) string {
	stdout, _, err := new(gc.PipedExec).
		Command("git", "log", "-n", "1", `--pretty=format:%H`).
		WorkingDir(repoDir).
		RunToStrings()
	gc.PanicIfError(err)

	return strings.TrimSpace(stdout)
}
