/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newRandomStatistic() *Statistic {
	queryNumber := rand.Uint32()
	qps := rand.Float64() * 1000
	min := rand.Float64() * 10
	max := min + rand.Float64()*100
	avg := min + rand.Float64()*(max-min)
	p50 := min + rand.Float64()*(max-min)
	p75 := p50 + rand.Float64()*(max-p50)
	p90 := p75 + rand.Float64()*(max-p75)
	p95 := p90 + rand.Float64()*(max-p90)
	p99 := p95 + rand.Float64()*(max-p95)

	return &Statistic{
		QueryNumber: uint(queryNumber),
		QPS:         qps,
		Min:         min,
		Max:         max,
		Avg:         avg,
		P50:         p50,
		P75:         p75,
		P90:         p90,
		P95:         p95,
		P99:         p99,
	}
}

func TestWriteTestSummaryToJson(t *testing.T) {
	loadTest := newTestSummary()

	var buf bytes.Buffer
	err := loadTest.WriteJson(&buf)
	assert.NoError(t, err)

	var writtenLoadTest TestSummary
	err = json.Unmarshal(buf.Bytes(), &writtenLoadTest)
	assert.NoError(t, err)

	assert.Equal(t, loadTest, &writtenLoadTest)
}

func newTestSummary() *TestSummary {
	loadTest := &TestSummary{
		TestTime: time.Now().Format(time.RFC3339),
		Config: &TestConfig{
			TestName:       "Test1",
			ServerName:     "localhost",
			ServerPort:     8080,
			VerifyTls:      true,
			Connections:    100,
			CreateSession:  false,
			WarmupDuration: 10 * time.Second,
			TestDuration:   30 * time.Second,
			TargetQPS:      1000,
			Sobject:        nil,
		},
		Result: &TestResult{
			Warmup: newRandomStatistic(),
			Test:   newRandomStatistic(),
			ProfilingResults: &ProfilingStatistics{
				InQueue:       *newRandomStatistic(),
				ParseRequest:  *newRandomStatistic(),
				SessionLookup: *newRandomStatistic(),
				ValidateInput: *newRandomStatistic(),
				CheckAccess:   *newRandomStatistic(),
				Operate:       *newRandomStatistic(),
				DbFlush:       *newRandomStatistic(),
				Total:         *newRandomStatistic(),
			},
		},
	}
	return loadTest
}
