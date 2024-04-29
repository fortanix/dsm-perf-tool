/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

const defaultServerName = "sdkms.test.fortanix.com"
const defaultServerPort = uint16(443)
const defaultTlsNotVerify = false
const defaultRequestTimeout = 60 * time.Second
const defaultIdleConnectionTimeout = 0 * time.Second

var serverName string
var serverPort uint16
var insecureTLS bool
var requestTimeout time.Duration
var idleConnectionTimeout time.Duration

var rootCmd = &cobra.Command{
	Use:   "dsm-perf-tool",
	Short: "DSM performance tool",
	Long:  `DSM performance tool`,
}

func ExecuteCmd() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", defaultServerName, "DSM server host name")
	rootCmd.PersistentFlags().Uint16VarP(&serverPort, "port", "p", defaultServerPort, "DSM server port")
	rootCmd.PersistentFlags().BoolVar(&insecureTLS, "insecure", defaultTlsNotVerify, "Do not validate server's TLS certificate")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "request-timeout", defaultRequestTimeout, "HTTP request timeout, 0 means no timeout")
	rootCmd.PersistentFlags().DurationVar(&idleConnectionTimeout, "idle-connection-timeout", defaultIdleConnectionTimeout, "Idle connection timeout, 0 means no timeout (default behavior)")

}
