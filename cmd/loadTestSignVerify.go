/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

const SIGN_EXAMPLE_DATA string = "0123456789abcdef"

var signKeyID string
var verifyOpt bool

var signVerifyLoadTestCmd = &cobra.Command{
	Use:     "sign-verify",
	Aliases: []string{"sign", "verify"},
	Short:   "Perform sign/verify load test.",
	Long:    "Perform sign/verify load test.",
	Run: func(cmd *cobra.Command, args []string) {
		signVerifyLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(signVerifyLoadTestCmd)

	signVerifyLoadTestCmd.PersistentFlags().StringVar(&signKeyID, "kid", "", "Key ID to use for sign and verify")
	signVerifyLoadTestCmd.PersistentFlags().BoolVar(&verifyOpt, "verify", false, "Perform verification instead of sign")
}

func signVerifyLoadTest() {
	setup := func(client *sdkms.Client) (interface{}, error) {
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
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingDataStr, error) {
		if signResp, ok := arg.(*sdkms.SignResponse); verifyOpt && ok {
			_, d, p, err := verify(client, *signResp)
			// return the sign response so we can verify in the next iteration
			return signResp, d, p, err
		}
		return sign(client)
	}
	name := "sign"
	if verifyOpt {
		name = "verify"
	}
	if createSession {
		name += " with session"
	}
	loadTest(name, setup, test, cleanup)
}

func sign(client *sdkms.Client) (*sdkms.SignResponse, time.Duration, profilingDataStr, error) {
	req := sdkms.SignRequest{
		Data:    someBlob([]byte(SIGN_EXAMPLE_DATA)),
		HashAlg: sdkms.DigestAlgorithmSha256,
		Key:     sdkms.SobjectByID(signKeyID),
	}

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	res, err := client.Sign(ctx, req)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return res, d, p, err
}

func verify(client *sdkms.Client, sr sdkms.SignResponse) (*sdkms.VerifyResponse, time.Duration, profilingDataStr, error) {
	req := sdkms.VerifyRequest{
		Signature: sr.Signature,
		Key:       sdkms.SobjectByID(signKeyID),
		HashAlg:   sdkms.DigestAlgorithmSha256,
		Data:      someBlob([]byte(SIGN_EXAMPLE_DATA)),
	}

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	res, err := client.Verify(ctx, req)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return res, d, p, err
}

func someBlob(blob sdkms.Blob) *sdkms.Blob { return &blob }
