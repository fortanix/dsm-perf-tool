package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	input := json.RawMessage(pluginInput)
	_, err := json.Marshal(&input)
	if err != nil {
		fmt.Printf("plugin input must be valid JSON: %v\n", err)
		os.Exit(1)
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
	name := "plugin invocation"
	if createSession {
		name += " with session"
	}
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
