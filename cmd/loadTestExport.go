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

var exportLoadTestCmd = &cobra.Command{
	Use:   "export",
	Short: "Load test using export API",
	Long:  "Load test using export API",
	Run: func(cmd *cobra.Command, args []string) {
		exportLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(exportLoadTestCmd)

	exportLoadTestCmd.PersistentFlags().StringVar(&keyID, "kid", "", "ID of the key to export")
}

func exportLoadTest() {
	setup := func(client *sdkms.Client, testConfig *TestConfig) (interface{}, error) {

		if createSession {
			_, err := client.AuthenticateWithAPIKey(context.Background(), apiKey)
			return nil, err
		}
		client.Auth = sdkms.APIKey(apiKey)
		return nil, nil
	}
	cleanup := func(client *sdkms.Client) {
		if createSession {
			client.TerminateSession(context.Background())
		}
	}
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingMetricStr, error) {
		ctx := context.WithValue(context.Background(), responseHeaderKey, http.Header{})

		t0 := time.Now()
		_, err := client.ExportSobject(ctx, *sdkms.SobjectByID(keyID))
		d := time.Since(t0)

		header := ctx.Value(responseHeaderKey).(http.Header)
		p := profilingMetricStr(header.Get("Profiling-Data"))

		return nil, d, p, err
	}

	loadTest("export", setup, test, cleanup)
}
