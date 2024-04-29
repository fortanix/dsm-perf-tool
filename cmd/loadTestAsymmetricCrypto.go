/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

const ASYM_EXAMPLE_DATA string = "0123456789abcdef"

var asymmetricCryptoLoadTestCmd = &cobra.Command{
	Use:     "asymmetric-crypto",
	Aliases: []string{"asymmetric", "asym"},
	Short:   "Perform asymmetric encryption/decryption load test.",
	Long:    "Perform asymmetric encryption/decryption load test.",
	Run: func(cmd *cobra.Command, args []string) {
		asymmetricCryptoLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(asymmetricCryptoLoadTestCmd)

	asymmetricCryptoLoadTestCmd.PersistentFlags().StringVar(&keyID, "kid", "", "Key ID to use for asymmetric crypto")
	asymmetricCryptoLoadTestCmd.PersistentFlags().BoolVar(&decryptOpt, "decrypt", false, "Perform decryption instead of encryption")
}

func asymmetricCryptoLoadTest() {
	// get basic info of the given sobject
	key := GetSobject(&keyID)

	setup := func(client *sdkms.Client, testConfig *TestConfig) (interface{}, error) {
		if testConfig.Sobject != nil {
			testConfig.Sobject = key
		}
		if createSession {
			_, err := client.AuthenticateWithAPIKey(context.Background(), apiKey)
			return nil, err
		}
		client.Auth = sdkms.APIKey(apiKey)
		return nil, nil
	}
	cleanup := func(client *sdkms.Client) {
		if createSession {
			client.TerminateSession(context.Background())
		}
	}
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingMetricStr, error) {
		if er, ok := arg.(*sdkms.EncryptResponse); decryptOpt && ok {
			_, d, p, err := asymmetricDecrypt(client, *er)
			// return the encrypt response so we can decrypt in the next iteration
			return er, d, p, err
		}
		return asymmetricEncrypt(client)
	}

	// construct test name
	name := "asymmetric encryption"
	if decryptOpt {
		name = "asymmetric decryption"
	}
	if createSession {
		name += " with session"
	}
	name = fmt.Sprintf("%s %d %s", key.ObjType, *key.KeySize, name)

	// start the load test
	loadTest(name, setup, test, cleanup)
}

func asymmetricEncrypt(client *sdkms.Client) (*sdkms.EncryptResponse, time.Duration, profilingMetricStr, error) {
	req := sdkms.EncryptRequest{
		Key:   sdkms.SobjectByID(keyID),
		Alg:   sdkms.AlgorithmRsa,
		Plain: []byte(ASYM_EXAMPLE_DATA),
	}

	ctx := context.WithValue(context.Background(), responseHeaderKey, http.Header{})

	t0 := time.Now()
	res, err := client.Encrypt(ctx, req)
	d := time.Since(t0)

	header := ctx.Value(responseHeaderKey).(http.Header)
	p := profilingMetricStr(header.Get("Profiling-Data"))

	return res, d, p, err
}

func asymmetricDecrypt(client *sdkms.Client, c sdkms.EncryptResponse) (*sdkms.DecryptResponse, time.Duration, profilingMetricStr, error) {
	req := sdkms.DecryptRequest{
		Key:    sdkms.SobjectByID(keyID),
		Alg:    someAlgorithm(sdkms.AlgorithmRsa),
		Cipher: c.Cipher,
		Iv:     c.Iv,
		Tag:    c.Tag,
	}

	ctx := context.WithValue(context.Background(), responseHeaderKey, http.Header{})

	t0 := time.Now()
	res, err := client.Decrypt(ctx, req)
	d := time.Since(t0)

	header := ctx.Value(responseHeaderKey).(http.Header)
	p := profilingMetricStr(header.Get("Profiling-Data"))

	return res, d, p, err
}
