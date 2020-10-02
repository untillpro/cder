/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	repoURLs  []string
	// used in tests
	afterIteration func()              = func() {}
	onError        func(r interface{}) = func(r interface{}) {}
)

func runCmdRoot(cmd *cobra.Command, args []string) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// *************************************************
	gc.Doing("Starting seeding")
	if ctx == nil {
		ctx, cancel = context.WithCancel(context.Background())
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go seed(ctx, &wg)

	go func() {
		signal := <-signals
		fmt.Println("Got signal:", signal)
		cancel()
	}()

	// *************************************************
	wg.Wait()

	gc.Info("Finished")
	return nil
}

func seed(ctx context.Context, wg *sync.WaitGroup) {
	gc.Info("Seeding started")
	defer func() {
		if r := recover(); r != nil {
			gc.Error("Recovered: ", r)
			onError(r)
		}
		defer wg.Done()
		deployer.Stop()
		gc.Info("Seeding ended")
	}()
	timeoutDur := time.Duration(timeoutSec) * time.Second

	// *************************************************
	gc.Info("timeout", timeoutDur)
	gc.Info("replacements", replacements)
	gc.Info("repos", repoURLs)

	if len(workingDir) > 0 {
		gc.Info("Creating working dir...")
		gc.PanicIfError(os.MkdirAll(workingDir, 0755))
	}

	// *************************************************
	for _, initCmdCombined := range initCmds {
		initCmdSplitted := strings.Split(initCmdCombined, ";")
		for _, initCmd := range initCmdSplitted {
			gc.Info("Executing init command:", initCmd)
			initArgs := strings.Split(initCmd, " ")
			gc.ExitIfError(new(gc.PipedExec).
				Command(initArgs[0], initArgs[1:]...).
				WorkingDir(workingDir).
				Run(os.Stdout, os.Stderr))
		}
	}

	for {
		iteration()
		afterIteration()
		// TODO: clean WD after iteration
		select {
		case <-time.After(timeoutDur):
		case <-ctx.Done():
			gc.Verbose("seeder", "Done")
			return
		}
	}
}

func iteration() {
	defer func() {
		if r := recover(); r != nil {
			gc.Error("iteration: Recovered: ", r)
			onError(r)
		}
	}()

	gc.Verbose("iteration", "Checking if repos changed")
	changedRepos := watcher.Watch(repoURLs)
	if len(changedRepos) > 0 {
		for _, changedRepo := range changedRepos {
			deployer.Deploy(changedRepo)
		}
		deployer.DeployAll(changedRepos)
		watcher.Clean(changedRepos)
	} else {
		gc.Verbose("*** Nothing changed")
	}
}
