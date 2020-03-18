/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/untillpro/gochips"
)

var (
	tempDir string
	script  = `
#!/bin/bash
case $1 in
	"deploy")
		rm -r ../../../deploy     
		mkdir ../../../deploy
		cp test%d.txt ../../../deploy/testDeployed%d.txt
		;;
	"stop")
		echo "deployer.stop"
		;;
esac
`
)

func TestCderURLBasic(t *testing.T) {
	setUp()
	defer tearDown()

	artifactoryDir := path.Join(tempDir, "artifactory")
	require.Nil(t, os.MkdirAll(artifactoryDir, 0755))

	// artifact1
	srcArtifact1File := path.Join(artifactoryDir, "test1.txt")
	srcArtifact1Zip := path.Join(artifactoryDir, "artifact1.zip")
	require.Nil(t, ioutil.WriteFile(srcArtifact1File, []byte("hello, world!"), 0755))
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("test1.txt")
	require.Nil(t, err)
	artBytes, err := ioutil.ReadFile(srcArtifact1File)
	require.Nil(t, err)
	_, err = f.Write(artBytes)
	require.Nil(t, err)
	require.Nil(t, w.Close())
	require.Nil(t, ioutil.WriteFile(srcArtifact1Zip, buf.Bytes(), 0755))

	// artifact2
	srcArtifact2File := path.Join(artifactoryDir, "test2.txt")
	srcArtifact2Zip := path.Join(artifactoryDir, "artifact2.zip")
	require.Nil(t, ioutil.WriteFile(srcArtifact2File, []byte("hello, world! 2"), 0755))
	buf = new(bytes.Buffer)
	w = zip.NewWriter(buf)
	f, err = w.Create("test2.txt")
	require.Nil(t, err)
	artBytes, err = ioutil.ReadFile(srcArtifact2File)
	require.Nil(t, err)
	_, err = f.Write(artBytes)
	require.Nil(t, err)
	require.Nil(t, w.Close())
	require.Nil(t, ioutil.WriteFile(srcArtifact2Zip, buf.Bytes(), 0755))

	counter := 1
	tsArt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		artZip := srcArtifact1Zip
		if counter >= 2 {
			artZip = srcArtifact2Zip
		}
		bytes, err := ioutil.ReadFile(artZip)
		require.Nil(t, err)
		w.Write(bytes)
	}))
	defer tsArt.Close()

	tsDeployer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch counter {
		case 1:
			fmt.Fprintf(w, script, 1, 1)
		case 2:
			// deployer logic unchanged on 2nd artifact version
			fmt.Fprintf(w, script, 2, 1)
		default:
			// new deployer logic for 2nd artifact version
			fmt.Fprintf(w, script, 2, 2)
		}
	}))
	defer tsDeployer.Close()
	tsMain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch counter {
		case 1:
			fmt.Fprintln(w, tsArt.URL+"/"+filepath.Base(srcArtifact1Zip))
			fmt.Fprintln(w, tsDeployer.URL+"/"+"1")
		case 2:
			fmt.Fprintln(w, tsArt.URL+"/"+filepath.Base(srcArtifact2Zip))
			fmt.Fprintln(w, tsDeployer.URL+"/"+"1")
		default:
			fmt.Fprintln(w, tsArt.URL+"/"+filepath.Base(srcArtifact2Zip))
			fmt.Fprintln(w, tsDeployer.URL+"/"+"2")
		}
	}))
	defer tsMain.Close()

	cmdRoot.SetArgs([]string{"cdurl", "--url", tsMain.URL, "--verbose", "--working-dir", tempDir, "--init", "bash -c 'mkdir init'", "--timeout", "0"})
	ctx, cancel = context.WithCancel(context.Background())
	afterIteration = func() {
		fmt.Printf("******************* iter %d finished\n", counter)
		var actualBytes []byte
		var expectedBytes []byte
		switch counter {
		case 1:
			actualBytes, err = ioutil.ReadFile(path.Join(workingDir, "deploy/testDeployed1.txt"))
			require.Nil(t, err)
			expectedBytes, err = ioutil.ReadFile(path.Join(artifactoryDir, "test1.txt"))
			require.Nil(t, err)
		case 2:
			// 2nd artifact version, using deployer logic from 1st version
			actualBytes, err = ioutil.ReadFile(path.Join(workingDir, "deploy/testDeployed1.txt"))
			require.Nil(t, err)
			expectedBytes, err = ioutil.ReadFile(path.Join(artifactoryDir, "test2.txt"))
			require.Nil(t, err)
		case 3, 4:
			// 2nd artifact version, new deployer logic
			// nothing changed on iter 4
			actualBytes, err = ioutil.ReadFile(path.Join(workingDir, "deploy/testDeployed2.txt"))
			require.Nil(t, err)
			expectedBytes, err = ioutil.ReadFile(path.Join(artifactoryDir, "test2.txt"))
			require.Nil(t, err)
		}
		require.Equal(t, expectedBytes, actualBytes)
		require.DirExists(t, path.Join(tempDir, "init"))
		if counter > 3 {
			cancel()
		}
		counter++
	}

	require.Nil(t, execute())
}

func TestBuildScript(t *testing.T) {
	require.NotNil(t, new(gochips.PipedExec).
		Command("env", "VER=5-SNAPSHOT", "./build.sh").
		Run(os.Stdout, os.Stderr))
}

func setUp() {
	onError(nil)
	afterIteration()
	onError = func(r interface{}) {
		panic(r)
	}
	var err error
	tempDir, err = ioutil.TempDir(os.TempDir(), "cder_test")
	if err != nil {
		panic(err)
	}
}

func tearDown() {
	cmdRoot.ResetFlags()
	cmdCDGit.ResetFlags()
	cmdCDURL.ResetFlags()
	// os.RemoveAll(tempDir)
}
