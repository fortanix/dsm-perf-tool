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

var keyType = objectTypeAES
var keySize uint32

var loadTestGenerateKeyCmd = &cobra.Command{
	Use:     "generate-key",
	Aliases: []string{"generate", "gen"},
	Short:   "Perform key generation load test.",
	Long:    "Perform key generation load test.",
	Run: func(cmd *cobra.Command, args []string) {
		loadTestGenerateKey()
	},
}

func init() {
	loadTestCmd.AddCommand(loadTestGenerateKeyCmd)

	loadTestGenerateKeyCmd.PersistentFlags().VarP(&keyType, "type", "t", "Type of key to generate, support: AES, RSA, EC (EC-NistP256)")
	loadTestGenerateKeyCmd.PersistentFlags().Uint32Var(&keySize, "size", 0, "Key size (defaults to 256 for AES and 2048 for RSA)")
}

func loadTestGenerateKey() {
	// Default key size
	if keySize == 0 {
		switch keyType {
		case objectTypeAES:
			keySize = 256
		case objectTypeRSA:
			keySize = 2048
		}
	}
	setup := func(client *sdkms.Client) (interface{}, error) {
		_, err := client.AuthenticateWithAPIKey(context.Background(), apiKey)
		return nil, err
	}
	cleanup := func(client *sdkms.Client) {
		client.TerminateSession(context.Background())
	}
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingDataStr, error) {
		// Don't want to generate a key in warmup, this is OK because we ensure TLS is established in setup() by authenticating
		if stage == warmupStage {
			return nil, 0, "", nil
		}
		d, p, err := generateKey(client)
		return nil, d, p, err
	}
	name := fmt.Sprintf("generate %v keys", keyType)
	if keyType == objectTypeAES || keyType == objectTypeRSA {
		name += fmt.Sprintf(" (%v bits)", keySize)
	}
	loadTest(name, setup, test, cleanup)
}

func generateKey(client *sdkms.Client) (time.Duration, profilingDataStr, error) {
	req := sdkms.SobjectRequest{
		Transient:     someBool(true),
		ObjType:       convertObjectType(keyType),
		KeySize:       keySizeFor(keyType),
		EllipticCurve: ellipticCurveFor(keyType),
	}

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	_, err := client.CreateSobject(ctx, req)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return d, p, err
}

func someBool(x bool) *bool       { return &x }
func someUint32(x uint32) *uint32 { return &x }
func convertObjectType(t objectType) *sdkms.ObjectType {
	c := sdkms.ObjectType(string(t))
	return &c
}
func keySizeFor(t objectType) *uint32 {
	switch t {
	case objectTypeAES, objectTypeRSA:
		return someUint32(keySize)
	default:
		return nil
	}
}
func ellipticCurveFor(t objectType) *sdkms.EllipticCurve {
	var x sdkms.EllipticCurve
	switch t {
	case objectTypeEC:
		x = sdkms.EllipticCurveNistP256
	default:
		return nil
	}
	return &x
}
