/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const defaultServerName = "sdkms.test.fortanix.com"
const defaultServerPort = uint16(443)
const defaultTlsNotVerify = false
const defaultRequestTimeout = 60 * time.Second
const defaultIdleConnectionTimeout = 0 * time.Second

const (
	Plain string = "plain"
	JSON  string = "json"
	YAML  string = "yaml"
)

// TODO: get rid of global variables, tracking issue: #16
var serverName string
var serverPort uint16
var insecureTLS bool
var requestTimeout time.Duration
var idleConnectionTimeout time.Duration
var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "dsm-perf-tool",
	Short: "DSM performance tool",
	Long:  `DSM performance tool`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// validate arguments
		switch outputFormat {
		case Plain:
			// Plain is accepted
		case JSON:
			// JSON is accepted
		case YAML:
			return fmt.Errorf("yaml is not yet supported")
		default:
			return fmt.Errorf("unacceptable output format option: %v", outputFormat)
		}
		return nil
	},
}

func ExecuteCmd() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "sdkms.test.fortanix.com", "DSM server host name")
	rootCmd.PersistentFlags().Uint16VarP(&serverPort, "port", "p", 443, "DSM server port")
	rootCmd.PersistentFlags().BoolVar(&insecureTLS, "insecure", false, "Do not validate server's TLS certificate")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", Plain, "Output format, accepted options are: 'plain', 'json'")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "request-timeout", 60*time.Second, "HTTP request timeout, 0 means no timeout")
	rootCmd.PersistentFlags().DurationVar(&idleConnectionTimeout, "idle-connection-timeout", 0, "Idle connection timeout, 0 means no timeout (default behavior)")
}
