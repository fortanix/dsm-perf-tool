/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/montanaflynn/stats"
)

type profilingMetricStr string

type profilingDataArr []profilingData

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

func getProfilingMetrics(dataArr profilingDataArr) *ProfilingStatistics {
	var inQueueData stats.Float64Data
	var parseRequestData stats.Float64Data
	var sessionLookupData stats.Float64Data
	var validateInputData stats.Float64Data
	var checkAccessData stats.Float64Data
	var operateData stats.Float64Data
	var dbFlushData stats.Float64Data
	var totalData stats.Float64Data
	for _, data := range dataArr {
		inQueueData = append(inQueueData, float64(data.InQueue))
		parseRequestData = append(parseRequestData, float64(data.ParseRequest))
		sessionLookupData = append(sessionLookupData, float64(data.SessionLookup))
		validateInputData = append(validateInputData, float64(data.ValidateInput))
		checkAccessData = append(checkAccessData, float64(data.CheckAccess))
		operateData = append(operateData, float64(data.Operate))
		dbFlushData = append(dbFlushData, float64(data.DbFlush))
		totalData = append(totalData, float64(data.Total))
	}
	return &ProfilingStatistics{
		InQueue:       *StatisticFromFloat64Data(inQueueData, nil),
		ParseRequest:  *StatisticFromFloat64Data(parseRequestData, nil),
		SessionLookup: *StatisticFromFloat64Data(sessionLookupData, nil),
		ValidateInput: *StatisticFromFloat64Data(validateInputData, nil),
		CheckAccess:   *StatisticFromFloat64Data(checkAccessData, nil),
		Operate:       *StatisticFromFloat64Data(operateData, nil),
		DbFlush:       *StatisticFromFloat64Data(dbFlushData, nil),
		Total:         *StatisticFromFloat64Data(totalData, nil),
	}
}

func (dataArr profilingDataArr) getCSVHeaders() (header []string) {
	e := reflect.ValueOf(&dataArr[0]).Elem()
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		header = append(header, varName)
	}
	return header
}

func (dataArr profilingDataArr) getCSVValues() (values [][]string) {
	for _, data := range dataArr {
		values = append(values, []string{
			strconv.FormatUint(data.InQueue, 10),
			strconv.FormatUint(data.ParseRequest, 10),
			strconv.FormatUint(data.SessionLookup, 10),
			strconv.FormatUint(data.ValidateInput, 10),
			strconv.FormatUint(data.CheckAccess, 10),
			strconv.FormatUint(data.Operate, 10),
			strconv.FormatUint(data.DbFlush, 10),
			strconv.FormatUint(data.Total, 10),
		})
	}
	return values
}

func saveProfilingMetricsToCSV(dataArr profilingDataArr) {
	csvFile, err := os.CreateTemp(".", "profilingData.*.csv")
	if err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}
	w := csv.NewWriter(csvFile)
	headers := dataArr.getCSVHeaders()
	values := dataArr.getCSVValues()
	if err := w.Write(headers); err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}
	if err := w.WriteAll(values); err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}

	log.Println("Saved profiling data to:", csvFile.Name())
}

func parseProfilingMetricStrArr(profilingDataStrArr []profilingMetricStr) profilingDataArr {
	var dataArr profilingDataArr
	for _, profilingDataStr := range profilingDataStrArr {
		var profilingData profilingData
		err := json.Unmarshal([]byte(string(profilingDataStr)), &profilingData)
		if err != nil {
			log.Fatalf("Fatal error: %v\n", err)
		}
		dataArr = append(dataArr, profilingData)
	}
	return dataArr
}
