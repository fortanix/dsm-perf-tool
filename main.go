/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"os"

	"github.com/fortanix/dsm-perf-tool/cmd"
)

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
