/* Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	gc "github.com/untillpro/gochips"
)

type watcherURL struct {
	artifactURLStored string
	deployerURLStored string
}

func (w *watcherURL) Watch(repos []string) (changedRepos []string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	repo := repos[0]
	bodyBytes := readFromURL(client, repo)
	if bodyBytes == nil {
		return
	}
	contentStr := string(bodyBytes)
	content := strings.Split(contentStr, "\n")
	artifactURLNew := content[0]
	deployerURLNew := content[1]
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")                                            // assuming errors are impossible at runtime
	_, artifactFileName := parseArtifactURL(artifactURLNew)                              // artifact1
	artifactHomePath := path.Join(getArtifactsFolder(), reg.ReplaceAllString(repo, "_")) // artifacts/<url>
	artifactDeployerFile := path.Join(artifactHomePath, "deploy.sh")                     // artifacts/<url>/deploy.sh
	artifactZipFile := path.Join(artifactHomePath, artifactFileName)                     // artifacts/<url>/artifact1.zip
	artifactWD := path.Join(artifactHomePath, "work-dir")                                // artifacts/<url>/work-dir/

	isChanged := false
	deployer = &deployer4sh{
		wd: artifactWD,
	}

	if artifactURLNew != w.artifactURLStored {
		gc.Info("watcherURL:", "zip url changed", w.artifactURLStored, artifactURLNew)
		gc.Info("watcherURL:", "cleaning artifact home dir")
		files, err := filepath.Glob(path.Join(artifactHomePath, "*.zip"))
		gc.PanicIfError(err)
		for _, f := range files {
			gc.PanicIfError(os.Remove(f))
		}
		gc.PanicIfError(os.RemoveAll(artifactWD))
		gc.PanicIfError(os.MkdirAll(artifactWD, 0755))
		gc.Info("watcherURL:", "downloading zip...")
		artifactZipBytes := readFromURL(client, artifactURLNew)
		if artifactZipBytes == nil {
			return
		}
		gc.Info("watcherURL:", "saving zip...")
		gc.PanicIfError(ioutil.WriteFile(artifactZipFile, artifactZipBytes, 0755))
		unzipAll(artifactZipFile, artifactWD)
		isChanged = true
		w.artifactURLStored = artifactURLNew
		w.deployerURLStored = ""
	}

	if deployerURLNew != w.deployerURLStored {
		gc.Info("watcherURL:", "deployer url changed", w.deployerURLStored, deployerURLNew)
		gc.Info("watcherURL:", "downloading deployer...")
		artifactDeployerBytes := readFromURL(client, deployerURLNew)
		if artifactDeployerBytes == nil {
			return
		}
		gc.Info("watcherURL:", "saving deployer...")
		os.MkdirAll(artifactHomePath, 0755)
		if !isChanged {
			unzipAll(artifactZipFile, artifactWD) // will clean work-dir
		}
		gc.PanicIfError(ioutil.WriteFile(artifactDeployerFile, artifactDeployerBytes, 0755))
		gc.PanicIfError(ioutil.WriteFile(path.Join(artifactWD, "deploy.sh"), artifactDeployerBytes, 0755))
		isChanged = true
		w.deployerURLStored = deployerURLNew
	}

	if isChanged {
		changedRepos = append(changedRepos, artifactWD)
	}

	return
}

func readFromURL(client *http.Client, url string) []byte {
	resp, err := client.Get(url)
	gc.PanicIfError(err)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		gc.Info("watcherURL:", fmt.Sprintf("response: %d", resp.StatusCode))
		return nil
	}
	res, err := ioutil.ReadAll(resp.Body)
	gc.PanicIfError(err)
	return res
}

func unzipAll(zipFile string, dir string) {
	gc.Info("watcherURL:", "unzipping...")
	gc.PanicIfError(os.RemoveAll(dir))
	gc.PanicIfError(os.MkdirAll(dir, 0755))

	r, err := zip.OpenReader(zipFile)
	gc.PanicIfError(err)
	defer r.Close()

	extractAndWriteFile := func(f *zip.File) {
		rc, err := f.Open()
		gc.PanicIfError(err)
		defer rc.Close()

		path := filepath.Join(dir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			gc.PanicIfError(err)
			defer f.Close()

			_, err = io.Copy(f, rc)
			gc.PanicIfError(err)
		}
	}

	for _, f := range r.File {
		extractAndWriteFile(f)
	}
}
