/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

var queriesPerSecond uint
var connections uint
var warmupDuration time.Duration
var testDuration time.Duration
var apiKey string
var createSession bool
var storeProfilingData bool

var loadTestCmd = &cobra.Command{
	Use:     "load-test",
	Aliases: []string{"load"},
	Short:   "A collection of load tests for various types of operations.",
	Long:    "A collection of load tests for various types of operations.",
}

func init() {
	rootCmd.AddCommand(loadTestCmd)

	loadTestCmd.PersistentFlags().UintVar(&queriesPerSecond, "qps", 10, "Queries per second (QPS)")
	loadTestCmd.PersistentFlags().UintVarP(&connections, "connections", "c", 10, "Number of concurrent connections")
	loadTestCmd.PersistentFlags().DurationVarP(&testDuration, "duration", "d", 30*time.Second, "Test duration")
	loadTestCmd.PersistentFlags().DurationVarP(&warmupDuration, "warmup", "w", 10*time.Second, "Warmup duration")
	loadTestCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "API key to use in some load tests")
	loadTestCmd.PersistentFlags().BoolVar(&createSession, "create-session", false, "Create a session for load tests (default is to use API Key as Basic auth header)")
	loadTestCmd.PersistentFlags().BoolVar(&storeProfilingData, "store-profiling-data", false, "Store profiling data in a csv file")
}

type loadTestStage int
type profilingDataStr string

// profilingData is delta metrics of each step in back-end of one API Call
// The unit of all fields are ns
type profilingData struct {
	InQueue                 uint64                    `json:"in_queue"`
	ParseRequest            uint64                    `json:"parse_request"`
	SessionLookup           uint64                    `json:"session_lookup"`
	ValidateInput           uint64                    `json:"validate_input"`
	CheckAccess             uint64                    `json:"check_access"`
	Operate                 uint64                    `json:"operate"`
	DbFlush                 uint64                    `json:"db_flush"`
	Total                   uint64                    `json:"total"`
	AdditionalProfilingData []additionalProfilingData `json:"additional_profiling,omitempty"`
}

type additionalProfilingData struct {
	Action     string                    `json:"action"`
	TookNs     uint64                    `json:"took_ns"`
	SubActions []additionalProfilingData `json:"sub_actions,omitempty"`
}

type profilingDataArr []profilingData

const (
	warmupStage loadTestStage = iota + 1
	testStage
)

type setupFunc func(client *sdkms.Client) (interface{}, error)
type testFunc func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingDataStr, error)
type cleanupFunc func(client *sdkms.Client)

func loadTest(name string, setup setupFunc, test testFunc, cleanup cleanupFunc) {
	fmt.Printf("      Load test: %v\n", name)
	fmt.Printf("         Server: %v:%v\n", serverName, serverPort)
	fmt.Printf("     Target QPS: %v\n", queriesPerSecond)
	fmt.Printf("    Connections: %v\n", connections)
	fmt.Printf("  Test Duration: %v\n", testDuration)
	fmt.Printf("Warmup Duration: %v\n", warmupDuration)
	fmt.Println()
	type testResult struct {
		t time.Time
		d time.Duration
		p profilingDataStr
		s loadTestStage
	}
	ticker := time.NewTicker(time.Duration(warmupDuration.Nanoseconds() / int64(connections)))
	start := make(chan struct{})
	end := make(chan struct{})
	result := make(chan testResult, 1000) // buffered channel just in case
	var ready, finished sync.WaitGroup
	var wg1 sync.WaitGroup

	launchWorker := func() {
		callTestFunc := func(t time.Time, client *sdkms.Client, stage loadTestStage, arg interface{}) interface{} {
			arg, d, p, err := test(client, stage, arg)
			if err != nil {
				if stage == warmupStage {
					log.Fatalf("Fatal error: %v\n", err)
				} else {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				result <- testResult{t, d, p, stage}
			}
			return arg
		}
		ready.Add(1)
		finished.Add(1)
		wg1.Add(1)
		go func() {
			defer wg1.Done()

			client := sdkmsClient()
			arg, err := setup(&client)
			if err != nil {
				log.Fatalf("Fatal error: %v\n", err)
			}
			// ensure TLS is established
			arg = callTestFunc(time.Time{}, &client, warmupStage, arg)
			ready.Done()
			<-start
		testLoop:
			for {
				select {
				case t := <-ticker.C:
					arg = callTestFunc(t, &client, testStage, arg)
				case <-end:
					break testLoop
				}
			}
			finished.Done()
			cleanup(&client)
		}()
	}

	var wg2 sync.WaitGroup
	wg2.Add(2)
	var warmups, tests []time.Duration
	var lastTick time.Time
	var profilingDataStrArr []profilingDataStr
	go func() {
		defer wg2.Done()

		for r := range result {
			if r.s == warmupStage {
				warmups = append(warmups, r.d)
			} else {
				tests = append(tests, r.d)
				if r.p != "" {
					profilingDataStrArr = append(profilingDataStrArr, r.p)
				}
			}

			lastTick = r.t
		}
	}()

	for i := uint(0); i < connections; i++ {
		<-ticker.C
		launchWorker()
	}
	ticker.Reset(time.Duration(time.Second.Nanoseconds() / int64(queriesPerSecond)))

	ready.Wait()
	t0 := time.Now()
	close(start)
	var t1 time.Time

	go func() {
		defer wg2.Done()

		time.Sleep(testDuration)
		close(end)
		finished.Wait()
		t1 = time.Now()
	}()
	wg1.Wait()
	close(result)
	wg2.Wait()
	ticker.Stop()

	sendDuration := lastTick.Sub(t0)
	testDuration := t1.Sub(t0)

	fmt.Printf("        Warmup: %v queries, %v\n", len(warmups), summarizeTimings(warmups))
	fmt.Printf("          Test: %v queries, %v\n", len(tests), summarizeTimings(tests))
	fmt.Printf(" Test duration: %v (%0.2f QPS)\n", testDuration, float64(len(tests))/testDuration.Seconds())
	fmt.Printf(" Send duration: %v (%0.2f QPS)\n", sendDuration, float64(len(tests))/sendDuration.Seconds())
	fmt.Printf("Profiling data: %v examples\n", len(profilingDataStrArr))

	if len(profilingDataStrArr) != 0 {
		dataArr := parseProfilingDataStrArr(profilingDataStrArr)
		summarizeProfilingData(dataArr)
		if storeProfilingData {
			saveProfilingDataToCSV(dataArr)
		}
	}
	fmt.Print("\n\n")
}
