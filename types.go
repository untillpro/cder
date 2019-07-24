/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

// IDeployer s.e.
type IDeployer interface {
	// Start and Stio must NOT panic
	Start() (err error)
	Stop()

	// Deploy functions must panic if something goes wrong
	Deploy(changedRepo string)
	DeployAll(repoPaths []string)
}
