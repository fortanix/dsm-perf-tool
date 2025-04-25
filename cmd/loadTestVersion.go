/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

var versionLoadTestCmd = &cobra.Command{
	Use:   "version",
	Short: "Load test using version API",
	Long:  "Load test using version API",
	Run: func(cmd *cobra.Command, args []string) {
		versionLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(versionLoadTestCmd)
}

func versionLoadTest() {
	setup := func(client *sdkms.Client, testConfig *TestConfig) (interface{}, error) {
		return nil, nil
	}
	cleanup := func(client *sdkms.Client) {}
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingMetricStr, error) {
		ctx := context.WithValue(context.Background(), responseHeaderKey, http.Header{})

		t0 := time.Now()
		_, err := client.Version(ctx, nil)
		d := time.Since(t0)

		header := ctx.Value(responseHeaderKey).(http.Header)
		p := profilingMetricStr(header.Get("Profiling-Data"))

		return nil, d, p, err
	}
	loadTest("version", setup, test, cleanup)
}
