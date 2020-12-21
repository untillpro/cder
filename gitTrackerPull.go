package main

import (
	"strings"

	gc "github.com/untillpro/gochips"
)

type gitTrackerPull struct {
}

func (t *gitTrackerPull) GetLastCommit(repoURL string, repoPath string) (lastCommit string, ok bool) {
	gc.Verbose("watcherGit", "Repo dir exists, will be pulled", repoPath, repoURL)
	stdouts, stderrs, err := new(gc.PipedExec).
		Command("git", "pull", repoURL).
		WorkingDir(repoPath).
		RunToStrings()
	if nil != err {
		gc.Info(stdouts, stderrs)
	}
	gc.PanicIfError(err)

	stdout, _, err := new(gc.PipedExec).
		Command("git", "log", "-n", "1", `--pretty=format:%H`).
		WorkingDir(repoPath).
		RunToStrings()
	gc.PanicIfError(err)

	return strings.TrimSpace(stdout), true
}
