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
