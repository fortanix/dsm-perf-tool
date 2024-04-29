/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/montanaflynn/stats"
)

func sdkmsClient() sdkms.Client {
	url := fmt.Sprintf("https://%v:%v", serverName, serverPort)
	// same values as http.DefaultTransport unless noted explicitly
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       idleConnectionTimeout, // different from http.DefaultTransport (90 sec)
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if insecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}
	client := sdkms.Client{
		HTTPClient: httpClient,
		Endpoint:   url,
	}
	return client
}

func setupCloseHandler(onClose func()) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Ctrl+C detected")
		onClose()
		os.Exit(0)
	}()
}

func summarizeTimings(times []time.Duration) string {
	if len(times) == 0 {
		return "--"
	}
	data := stats.LoadRawData(times)
	return summarizeData(data)
}

func summarizeData(data stats.Float64Data) string {
	var sb strings.Builder
	min, _ := data.Min()
	max, _ := data.Max()
	avg, _ := data.Mean()
	sb.WriteString(fmt.Sprintf("min = %0.3f, max = %0.3f, avg = %0.3f", min/1e6, max/1e6, avg/1e6))
	for _, p := range []float64{50, 75, 90, 95, 99} {
		num, _ := data.Percentile(p)
		sb.WriteString(fmt.Sprintf(", p%2.0f = %0.3f", p, num/1e6))
	}
	return sb.String()
}

type objectType string

const (
	objectTypeAES objectType = "AES"
	objectTypeRSA objectType = "RSA"
	objectTypeEC  objectType = "EC"
)

// impl pflag.Value interface for ObjectType

func (o *objectType) String() string {
	return string(*o)
}

func (o *objectType) Set(v string) error {
	switch v {
	case "aes", "AES":
		*o = objectTypeAES
	case "rsa", "RSA":
		*o = objectTypeRSA
	case "ec", "EC":
		*o = objectTypeEC
	default:
		return fmt.Errorf("invalid object type: %v", v)
	}
	return nil
}

func (o *objectType) Type() string {
	return "ObjectType"
}

// StrPad returns the input string padded on the left, right or both sides using padType to the specified padding length padLength.
//
// This helper function is from internet: https://gist.github.com/asessa/3aaec43d93044fc42b7c6d5f728cb039
//
// Example:
//
// input := "Codes";
//
// StrPad(input, 10, " ", "RIGHT")        // produces "Codes     "
//
// StrPad(input, 10, "-=", "LEFT")        // produces "=-=-=Codes"
//
// StrPad(input, 10, "_", "BOTH")         // produces "__Codes___"
//
// StrPad(input, 6, "___", "RIGHT")       // produces "Codes_"
//
// StrPad(input, 3, "*", "RIGHT")         // produces "Codes"
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetSobject retrieves a sobject through SDKMS client.
// It takes a keyID as a parameter and returns a pointer to a Sobject.
// If the keyID is empty, it will return an error.
// If an error occurs during the retrieval process, the function will log a fatal error and exit.
func GetSobject(kid *string) *sdkms.Sobject {
	client := sdkmsClient()
	client.Auth = sdkms.APIKey(apiKey)
	key, err := client.GetSobject(context.Background(), nil, sdkms.SobjectDescriptor{
		Kid: kid,
	})
	if err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}
	return key
}

type httpHeaderKey string

const (
	responseHeaderKey httpHeaderKey = "ResponseHeader"
)

func toJsonStr(v any) string {
	val, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}
	return string(val)
}
