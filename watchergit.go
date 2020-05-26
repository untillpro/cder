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

type watcherGit struct {
	commitsTracker   IGitTracker
	lastCommitHashes map[string]string
	// Used in the beginning of iteration to execute `git reset --hard`
	reposMustBeCleaned bool
}

func (w *watcherGit) Watch(repoURLs []string) (changedRepos []string) {
	defer func() {
		if r := recover(); r != nil {
			gc.Error("watcherGit: Recovered: ", r)
		}
	}()

	// *************************************************
	reposFolder := getReposFolder()

	for _, repoURL := range repoURLs {
		repoPath, repoFolder := getAbsRepoFolders(repoURL)

		gc.Verbose("watcherGit", "repoPath, repoFolder=", repoPath, repoFolder)

		gc.ExitIfError(os.MkdirAll(reposFolder, 0755))

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			gc.Info("watcherGit", "Repo folder does not exist, will be cloned", repoPath, repoURL)
			err := new(gc.PipedExec).
				Command("git", "clone", "--recurse-submodules", repoURL).
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
		}

		newHash, ok := w.commitsTracker.GetLastCommit(repoURL, repoPath)
		if ok {
			oldHash := w.lastCommitHashes[repoPath]
			if oldHash == newHash {
				continue
			}
			gc.Info("watcherGit", "Commit hash changed", repoURL, oldHash, newHash)
		} else if _, ok := w.lastCommitHashes[repoPath]; ok {
			// built once already -> skip
			continue
		}
		w.reposMustBeCleaned = true
		gitModulesPath := path.Join(repoPath, ".gitmodules")
		if _, err := os.Stat(gitModulesPath); err == nil {
			gc.Doing("watcherGit: updating modules")
			err = new(gc.PipedExec).
				Command("git", "submodule", "update", "--init", "--recursive").
				WorkingDir(repoPath).
				Run(os.Stdout, os.Stderr)
		}
		w.lastCommitHashes[repoPath] = newHash
		changedRepos = append(changedRepos, repoPath)
	}

	return
}
