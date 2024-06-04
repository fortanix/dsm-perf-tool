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
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

const SYM_EXAMPLE_DATA string = "0123456789ABCDEF"

// TODO: get rid of global variables, tracking issue: #16
var keyID string
var decryptOpt bool
var cipherModeStr string
var cipherMode sdkms.CipherMode
var tagLen = uint(128)

var symmetricCryptoLoadTestCmd = &cobra.Command{
	Use:     "symmetric-crypto",
	Aliases: []string{"symmetric", "sym"},
	Short:   "Perform symmetric encryption/decryption load test.",
	Long:    "Perform symmetric encryption/decryption load test.",
	Run: func(cmd *cobra.Command, args []string) {
		symmetricCryptoLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(symmetricCryptoLoadTestCmd)

	symmetricCryptoLoadTestCmd.PersistentFlags().StringVar(&keyID, "kid", "", "Key ID to use for symmetric crypto")
	symmetricCryptoLoadTestCmd.PersistentFlags().BoolVar(&decryptOpt, "decrypt", false, "Perform decryption instead of encryption")
	symmetricCryptoLoadTestCmd.PersistentFlags().StringVar(&cipherModeStr, "mode", "CBC", "Cipher mode used for encryption/decryption, support: CBC, GCM, FPE")
}

func symmetricCryptoLoadTest() {
	cipherMode = validateCipherMode(cipherModeStr)

	// get basic info of the given sobject
	key := GetSobject(&keyID)

	setup := func(client *sdkms.Client, testConfig *TestConfig) (interface{}, error) {
		if testConfig.Sobject == nil {
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
			_, d, p, err := decrypt(client, *er)
			// return the encrypt response so we can decrypt in the next iteration
			return er, d, p, err
		}
		return encrypt(client)
	}

	// construct test name
	operation := "symmetric encryption"
	if decryptOpt {
		operation = "symmetric decryption"
	}
	session := "without session"
	if createSession {
		session = "with session"
	}
	hiVolume := ""
	if key.KeyOps&sdkms.KeyOperationsHighvolume == sdkms.KeyOperationsHighvolume {
		hiVolume = "High Volume "
	}
	name := fmt.Sprintf("%s%s %d %s %s %s", hiVolume, key.ObjType, *key.KeySize, cipherModeStr, operation, session)

	// start the load test
	loadTest(name, setup, test, cleanup)
}

func encrypt(client *sdkms.Client) (*sdkms.EncryptResponse, time.Duration, profilingMetricStr, error) {
	req := sdkms.EncryptRequest{
		Key:    sdkms.SobjectByID(keyID),
		Alg:    sdkms.AlgorithmAes,
		Plain:  []byte(SYM_EXAMPLE_DATA),
		Mode:   sdkms.CryptModeSymmetric(cipherMode),
		TagLen: tagLenFor(cipherMode),
	}

	ctx := context.WithValue(context.Background(), responseHeaderKey, http.Header{})

	t0 := time.Now()
	res, err := client.Encrypt(ctx, req)
	d := time.Since(t0)

	header := ctx.Value(responseHeaderKey).(http.Header)
	p := profilingMetricStr(header.Get("Profiling-Data"))

	return res, d, p, err
}

func decrypt(client *sdkms.Client, c sdkms.EncryptResponse) (*sdkms.DecryptResponse, time.Duration, profilingMetricStr, error) {
	req := sdkms.DecryptRequest{
		Key:    sdkms.SobjectByID(keyID),
		Alg:    someAlgorithm(sdkms.AlgorithmAes),
		Cipher: c.Cipher,
		Iv:     c.Iv,
		Mode:   sdkms.CryptModeSymmetric(cipherMode),
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

func someAlgorithm(a sdkms.Algorithm) *sdkms.Algorithm { return &a }

func validateCipherMode(modeStr string) (mode sdkms.CipherMode) {
	switch {
	case modeStr == "CBC":
		mode = sdkms.CipherModeCbc
	case modeStr == "GCM":
		mode = sdkms.CipherModeGcm
	case modeStr == "FPE":
		mode = sdkms.CipherModeFf1
	default:
		log.Fatalf("Given cipher mode '%v' is no supported\n", modeStr)
	}
	return mode
}

func tagLenFor(mode sdkms.CipherMode) *uint {
	switch mode {
	case sdkms.CipherModeGcm:
		return &tagLen
	default:
		return nil
	}
}
