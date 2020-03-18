/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	gc "github.com/untillpro/gochips"
)

// GetAbsRepoFolders  ...
// <reposFolder>/<repoFolder>
// <repoPath                >
//  repoPath = reposFolder + '/' + repoFolder
func getAbsRepoFolders(repoURL string) (repoPath string, repoFolder string) {
	u, err := url.Parse(repoURL)
	gc.PanicIfError(err)
	urlParts := strings.Split(u.Path, "/")
	repoFolder = urlParts[len(urlParts)-1]
	repoPath, _ = filepath.Abs(path.Join(getReposFolder(), repoFolder))
	return
}

func parseArtifactURL(repoURL string) (aPath, aFN string) {
	u, err := url.Parse(repoURL)
	gc.PanicIfError(err)
	urlParts := strings.Split(u.Path, "/")
	aFN = urlParts[len(urlParts)-1]
	lastDotPos := strings.LastIndex(aFN, ".")
	artifactName := aFN[:lastDotPos]
	artifactPath, _ := filepath.Abs(path.Join(getArtifactsFolder(), artifactName))
	return artifactPath, aFN
}

func getReposFolder() string {
	return path.Join(workingDir, "repos")
}

func getArtifactsFolder() string {
	return path.Join(workingDir, "artifacts")
}
