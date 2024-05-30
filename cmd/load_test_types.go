/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/montanaflynn/stats"
)

type TestSummary struct {
	TestTime string      `json:"test_time" yaml:"test_time"` // ISO 8601 timestamp string
	Config   *TestConfig `json:"config" yaml:"config"`
	Result   *TestResult `json:"result" yaml:"result"`
}

type TestConfig struct {
	TestName       string           `json:"test_name" yaml:"test_name"`
	ServerName     string           `json:"server_name" yaml:"server_name"`
	ServerPort     uint16           `json:"server_port" yaml:"server_port"`
	VerifyTls      bool             `json:"verify_tls" yaml:"verify_tls"`
	Connections    uint             `json:"connections" yaml:"connections"`
	CreateSession  bool             `json:"create_session" yaml:"create_session"`
	WarmupDuration time.Duration    `json:"warmup_duration" yaml:"warmup_duration"`
	TestDuration   time.Duration    `json:"test_duration" yaml:"test_duration"`
	TargetQPS      uint             `json:"target_qps" yaml:"target_qps"`
	Sobject        *sdkms.Sobject   `json:"sobject" yaml:"sobject"`
	Plugin         *sdkms.Plugin    `json:"plugin" yaml:"plugin"`
	PluginInput    *json.RawMessage `json:"plugin_input" yaml:"plugin_input"`
}

func (tc *TestConfig) Print(w io.Writer) {
	fmt.Fprintf(w, "TestName:       %s\n", tc.TestName)
	fmt.Fprintf(w, "ServerName:     %s\n", tc.ServerName)
	fmt.Fprintf(w, "ServerPort:     %d\n", tc.ServerPort)
	fmt.Fprintf(w, "VerifyTls:      %t\n", tc.VerifyTls)
	fmt.Fprintf(w, "Connections:    %d\n", tc.Connections)
	fmt.Fprintf(w, "CreateSession:  %t\n", tc.CreateSession)
	fmt.Fprintf(w, "WarmupDuration: %s\n", tc.WarmupDuration)
	fmt.Fprintf(w, "TestDuration:   %s\n", tc.TestDuration)
	fmt.Fprintf(w, "TargetQPS:      %d\n", tc.TargetQPS)
	fmt.Fprintf(w, "Sobject:        %s\n", toJsonStr(tc.Sobject))
	fmt.Fprintf(w, "Plugin:         %s\n", toJsonStr(tc.Plugin))
	fmt.Fprintf(w, "PluginInput:    %s\n", toJsonStr(tc.PluginInput))
}

type TestResult struct {
	Warmup             *Statistic           `json:"warmup" yaml:"warmup"`
	Test               *Statistic           `json:"test" yaml:"test"`
	ActualTestDuration time.Duration        `json:"actual_test_duration" yaml:"actual_test_duration"`
	SendDuration       time.Duration        `json:"send_duration" yaml:"send_duration"`
	ProfilingResults   *ProfilingStatistics `json:"profiling_results" yaml:"profiling_results"`
}

func (tr *TestResult) Print(w io.Writer) {
	fmt.Fprintf(w, "Warmup:             %s\n", tr.Warmup.String())
	fmt.Fprintf(w, "Test:               %s\n", tr.Test.String())
	fmt.Fprintf(w, "ActualTestDuration: %s\n", tr.ActualTestDuration)
	fmt.Fprintf(w, "SendDuration:       %s\n", tr.ActualTestDuration)
}

// Statistic represents the performance metrics of a load test.
type Statistic struct {
	QueryNumber uint    `json:"query_number" yaml:"query_number"` // Number of queries executed
	QPS         float64 `json:"qps" yaml:"qps"`                   // Queries Per Second
	Avg         float64 `json:"avg" yaml:"avg"`                   // Average response time in nanoseconds
	Min         float64 `json:"min" yaml:"min"`                   // Minimum response time in nanoseconds
	Max         float64 `json:"max" yaml:"max"`                   // Maximum response time in nanoseconds
	P50         float64 `json:"p50" yaml:"p50"`                   // 50th percentile (median) response time in nanoseconds
	P75         float64 `json:"p75" yaml:"p75"`                   // 75th percentile response time in nanoseconds
	P90         float64 `json:"p90" yaml:"p90"`                   // 90th percentile response time in nanoseconds
	P95         float64 `json:"p95" yaml:"p95"`                   // 95th percentile response time in nanoseconds
	P99         float64 `json:"p99" yaml:"p99"`                   // 99th percentile response time in nanoseconds
}

func StatisticFromDurations(times []time.Duration, duration time.Duration) *Statistic {
	if len(times) == 0 {
		return nil
	}
	data := stats.LoadRawData(times)
	return StatisticFromFloat64Data(data, &duration)
}

func StatisticFromFloat64Data(data stats.Float64Data, totalDuration *time.Duration) *Statistic {
	queryNumber := uint(data.Len())
	min, _ := data.Min()
	max, _ := data.Max()
	avg, _ := data.Mean()
	p50, _ := data.Percentile(50)
	p75, _ := data.Percentile(75)
	p90, _ := data.Percentile(90)
	p95, _ := data.Percentile(95)
	p99, _ := data.Percentile(99)
	qps := math.NaN()
	if totalDuration != nil {
		qps = float64(queryNumber) / totalDuration.Seconds()
	}
	return &Statistic{
		QueryNumber: queryNumber,
		QPS:         qps,
		Avg:         avg,
		Min:         min,
		Max:         max,
		P50:         p50,
		P75:         p75,
		P90:         p90,
		P95:         p95,
		P99:         p99,
	}
}

func (st *Statistic) Print(w io.Writer) {
	fmt.Fprintf(w, "QueryNumber: %d, ", st.QueryNumber)
	fmt.Fprintf(w, "QPS: %.3f, ", st.QPS)
	fmt.Fprintf(w, "Avg: %.3fms, ", st.Avg/1e6)
	fmt.Fprintf(w, "Min: %.3fms, ", st.Min/1e6)
	fmt.Fprintf(w, "Max: %.3fms, ", st.Max/1e6)
	fmt.Fprintf(w, "P50: %.3fms, ", st.P50/1e6)
	fmt.Fprintf(w, "P75: %.3fms, ", st.P75/1e6)
	fmt.Fprintf(w, "P90: %.3fms, ", st.P90/1e6)
	fmt.Fprintf(w, "P95: %.3fms, ", st.P95/1e6)
	fmt.Fprintf(w, "P99: %.3fms", st.P99/1e6)
}

func (st *Statistic) String() string {
	buf := new(bytes.Buffer)
	st.Print(buf)
	return buf.String()
}

type ProfilingStatistics struct {
	InQueue       Statistic `json:"in_queue" yaml:"in_queue"`
	ParseRequest  Statistic `json:"parse_request" yaml:"parse_request"`
	SessionLookup Statistic `json:"session_lookup" yaml:"session_lookup"`
	ValidateInput Statistic `json:"validate_input" yaml:"validate_input"`
	CheckAccess   Statistic `json:"check_access" yaml:"check_access"`
	Operate       Statistic `json:"operate" yaml:"operate"`
	DbFlush       Statistic `json:"db_flush" yaml:"db_flush"`
	Total         Statistic `json:"total" yaml:"total"`
}

func (ps *ProfilingStatistics) Print(w io.Writer) {
	fmt.Fprintf(w, "InQueue:       %s\n", ps.InQueue.String())
	fmt.Fprintf(w, "ParseRequest:  %s\n", ps.ParseRequest.String())
	fmt.Fprintf(w, "SessionLookup: %s\n", ps.SessionLookup.String())
	fmt.Fprintf(w, "ValidateInput: %s\n", ps.ValidateInput.String())
	fmt.Fprintf(w, "CheckAccess:   %s\n", ps.CheckAccess.String())
	fmt.Fprintf(w, "Operate:       %s\n", ps.Operate.String())
	fmt.Fprintf(w, "DbFlush:       %s\n", ps.DbFlush.String())
	fmt.Fprintf(w, "Total:         %s\n", ps.Total.String())
}

type TestSummaryJsonWriter interface {
	WriteJson(w io.Writer) error
}

type TestSummaryPlainWriter interface {
	WritePlain(w io.Writer) error
}

func (ts *TestSummary) WritePlain(w io.Writer) error {
	fmt.Fprintf(w, "-----BEGIN TEST SUMMARY-----\n")
	fmt.Fprintf(w, "TestTime:       %v\n", ts.TestTime)
	ts.Config.Print(w)
	fmt.Fprintf(w, "\n")
	ts.Result.Print(w)
	fmt.Fprintf(w, "-----END TEST SUMMARY-----\n")
	return nil
}

func (ts *TestSummary) WriteJson(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ts)
}
