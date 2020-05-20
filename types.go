/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

// IDeployer s.e.
type IDeployer interface {
	Deploy(repo string)
	DeployAll(repos []string)
	Stop()
}

// IWatcher s.e.
type IWatcher interface {
	Watch(repos []string) (changedRepoPaths []string) // [0] must be main
}

// IGitTracker s.e.
type IGitTracker interface {
	// retrieves last commit from repo defined by `repoURL`. 
	// len(repoPath) > 0 -> the repo must be cloned already to `repoPath`. 
	// !ok -> no commits or no notifications about commits
	GetLastCommit(repoURL string, repoPath string) (lastCommit string, ok bool)
}
