/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

// TODO: get rid of global variables, tracking issue: #16
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

const PRINT_QPS_INTERVAL = 5 * time.Second

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

const (
	warmupStage loadTestStage = iota + 1
	testStage
)

type setupFunc func(client *sdkms.Client, testConfig *TestConfig) (interface{}, error)
type testFunc func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingMetricStr, error)
type cleanupFunc func(client *sdkms.Client)

func loadTest(name string, setup setupFunc, test testFunc, cleanup cleanupFunc) {
	testTime := time.Now()

	log.Printf("Load test:       %v\n", name)
	log.Printf("Server:          %v:%v\n", serverName, serverPort)
	log.Printf("Target QPS:      %v\n", queriesPerSecond)
	log.Printf("Connections:     %v\n", connections)
	log.Printf("Test Duration:   %v\n", testDuration)
	log.Printf("Warmup Duration: %v\n\n", warmupDuration)

	testConfig := TestConfig{
		TestName:       name,
		ServerName:     serverName,
		ServerPort:     serverPort,
		VerifyTls:      !insecureTLS,
		Connections:    connections,
		CreateSession:  createSession,
		WarmupDuration: warmupDuration,
		TestDuration:   testDuration,
		TargetQPS:      queriesPerSecond,
	}

	type testMetric struct {
		t time.Time
		d time.Duration
		p profilingMetricStr
		s loadTestStage
	}
	ticker := time.NewTicker(time.Duration(warmupDuration.Nanoseconds() / int64(connections)))
	start := make(chan struct{})
	end := make(chan struct{})
	result := make(chan testMetric, 1000) // buffered channel just in case
	var ready, finished sync.WaitGroup
	var wg1 sync.WaitGroup

	launchWorker := func() {
		callTestFunc := func(t time.Time, client *sdkms.Client, stage loadTestStage, arg interface{}) interface{} {
			arg, d, p, err := test(client, stage, arg)
			if err != nil {
				if stage == warmupStage {
					log.Fatalf("Fatal error: %v\n", err)
				} else {
					log.Printf("Error: %v\n", err)
				}
			} else {
				result <- testMetric{t, d, p, stage}
			}
			return arg
		}
		ready.Add(1)
		finished.Add(1)
		wg1.Add(1)
		go func() {
			defer wg1.Done()

			client := sdkmsClient()
			arg, err := setup(&client, &testConfig)
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
	var profilingMetricStrArr []profilingMetricStr

	var printQpsWg sync.WaitGroup
	go func() {
		defer wg2.Done()
		var lastPrintQpsTick time.Time
		lastQueryNum := len(tests)
		for r := range result {
			if r.s == warmupStage {
				warmups = append(warmups, r.d)
				// use last warmup ticket as start point
				lastPrintQpsTick = r.t
			} else {
				tests = append(tests, r.d)
				if r.t.After(lastPrintQpsTick.Add(PRINT_QPS_INTERVAL)) {
					lastPrintQpsTick = r.t
					currentQueryNum := len(tests)
					printQpsWg.Add(1)
					go func() {
						defer printQpsWg.Done()
						log.Printf("Last %s QPS: %v\n", PRINT_QPS_INTERVAL, float64(currentQueryNum-lastQueryNum)/PRINT_QPS_INTERVAL.Seconds())
						lastQueryNum = currentQueryNum
					}()
				}
				if r.p != "" {
					profilingMetricStrArr = append(profilingMetricStrArr, r.p)
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
	printQpsWg.Wait()

	sendDuration := lastTick.Sub(t0)
	testDuration := t1.Sub(t0)

	testResult := TestResult{
		Warmup:             StatisticFromDurations(warmups, warmupDuration),
		Test:               StatisticFromDurations(tests, testDuration),
		ActualTestDuration: testDuration,
		SendDuration:       sendDuration,
		ProfilingResults:   nil,
	}

	if len(profilingMetricStrArr) != 0 {
		dataArr := parseProfilingMetricStrArr(profilingMetricStrArr)
		testResult.ProfilingResults = getProfilingMetrics(dataArr)
		if storeProfilingData {
			saveProfilingMetricsToCSV(dataArr)
		}
	}

	testSummary := &TestSummary{
		TestTime: testTime.Format(time.RFC3339),
		Config:   &testConfig,
		Result:   &testResult,
	}

	switch outputFormat {
	case Plain:
		err := testSummary.WritePlain(os.Stdout)
		if err != nil {
			log.Fatalf("failed to write test summary in plain: %v\n", err)
		}
	case JSON:
		err := testSummary.WriteJson(os.Stdout)
		if err != nil {
			log.Fatalf("failed to write test summary in json: %v\n", err)
		}
	case YAML:
		log.Fatalf("write test summary in yaml is not yet supported\n")
	default:
		log.Fatalf("unreachable: unacceptable output format option: %v\n", outputFormat)
	}
}
