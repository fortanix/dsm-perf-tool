/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

var pluginID string
var pluginInput string

var invokePluginLoadTestCmd = &cobra.Command{
	Use:     "invoke-plugin",
	Aliases: []string{"invoke", "plugin"},
	Short:   "Perform plugin invocation load test.",
	Long:    "Perform plugin invocation load test.",
	Run: func(cmd *cobra.Command, args []string) {
		invokePluginLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(invokePluginLoadTestCmd)

	invokePluginLoadTestCmd.PersistentFlags().StringVar(&pluginID, "plugin-id", "", "ID of the plugin to invoke")
	invokePluginLoadTestCmd.PersistentFlags().StringVar(&pluginInput, "plugin-input", "null", "Input to pass to the plugin")
}

func invokePluginLoadTest() {
	// Get the given plugin from the server
	client := sdkmsClient()
	client.Auth = sdkms.APIKey(apiKey)
	plugin, err := client.GetPlugin(context.Background(), pluginID)
	if err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}

	input := json.RawMessage(pluginInput)
	_, err = json.Marshal(&input)
	if err != nil {
		log.Fatalf("Plugin input must be valid JSON: %v\n", err)
	}

	setup := func(client *sdkms.Client) (interface{}, error) {
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
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingDataStr, error) {
		return invokePlugin(client)
	}

	// construct test name
	name := fmt.Sprintf("invoke plugin '%s'", plugin.Name)
	if createSession {
		name += " with session"
	}

	// start the load test
	loadTest(name, setup, test, cleanup)
}

func invokePlugin(client *sdkms.Client) (*sdkms.PluginOutput, time.Duration, profilingDataStr, error) {
	input := json.RawMessage(pluginInput)

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	res, err := client.InvokePlugin(ctx, pluginID, &input)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return res, d, p, err
}
