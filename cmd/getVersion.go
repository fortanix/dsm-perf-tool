/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

var getVersionCount uint
var newConnectionPerRequest bool
var getVersionDelay time.Duration

var getVersionCmd = &cobra.Command{
	Use:     "get-version",
	Aliases: []string{"version", "ping"},
	Short:   "Call version API in a loop",
	Long:    "Call version API in a loop",
	Run: func(cmd *cobra.Command, args []string) {
		getVersion()
	},
}

func init() {
	rootCmd.AddCommand(getVersionCmd)

	getVersionCmd.Flags().UintVar(&getVersionCount, "count", 0, "Number of API calls after TLS setup, 0 means infinite (defaults to 0)")
	getVersionCmd.Flags().BoolVar(&newConnectionPerRequest, "new-connection-per-request", false, "Make a new connection for each request")
	getVersionCmd.Flags().DurationVar(&getVersionDelay, "delay", 1*time.Second, "Delay between API calls")
}

func getVersion() {
	client := sdkmsClient()
	v, err := client.Version(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v version: %v (%v)", serverName, v.Version, v.ServerMode)
	var durations []time.Duration
	printSummary := func() { fmt.Printf("%v\n", summarizeTimings(durations)) }
	setupCloseHandler(printSummary)
	var count uint
	for {
		if newConnectionPerRequest {
			client = sdkmsClient()
		}
		if count > 0 {
			time.Sleep(getVersionDelay)
		}
		t0 := time.Now()
		v, err := client.Version(context.Background())
		d := time.Since(t0)
		durations = append(durations, d)
		if err != nil {
			log.Fatal(err)
		}
		count++
		log.Printf("%d -- %v version: %v (%v) -- %v", count, serverName, v.Version, v.ServerMode, d)
		if getVersionCount > 0 && count == getVersionCount {
			break
		}
	}
	printSummary()
}
