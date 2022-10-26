/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "sdkms.test.fortanix.com", "DSM server host name")
	rootCmd.PersistentFlags().Uint16VarP(&serverPort, "port", "p", 443, "DSM server port")
	rootCmd.PersistentFlags().BoolVar(&insecureTLS, "insecure", false, "Do not validate server's TLS certificate")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "request-timeout", 60*time.Second, "HTTP request timeout, 0 means no timeout")
	rootCmd.PersistentFlags().DurationVar(&idleConnectionTimeout, "idle-connection-timeout", 0, "Idle connection timeout, 0 means no timeout (default behavior)")

}
