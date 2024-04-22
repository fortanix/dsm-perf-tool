package cmd

import (
	"context"
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
		if er, ok := arg.(*sdkms.EncryptResponse); decryptOpt && ok {
			_, d, p, err := asymmetricDecrypt(client, *er)
			// return the encrypt response so we can decrypt in the next iteration
			return er, d, p, err
		}
		return asymmetricEncrypt(client)
	}
	name := "asymmetric encryption"
	if decryptOpt {
		name = "asymmetric decryption"
	}
	if createSession {
		name += " with session"
	}
	loadTest(name, setup, test, cleanup)
}

func asymmetricEncrypt(client *sdkms.Client) (*sdkms.EncryptResponse, time.Duration, profilingDataStr, error) {
	req := sdkms.EncryptRequest{
		Key:   sdkms.SobjectByID(keyID),
		Alg:   sdkms.AlgorithmRsa,
		Plain: []byte(ASYM_EXAMPLE_DATA),
	}

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	res, err := client.Encrypt(ctx, req)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return res, d, p, err
}

func asymmetricDecrypt(client *sdkms.Client, c sdkms.EncryptResponse) (*sdkms.DecryptResponse, time.Duration, profilingDataStr, error) {
	req := sdkms.DecryptRequest{
		Key:    sdkms.SobjectByID(keyID),
		Alg:    someAlgorithm(sdkms.AlgorithmRsa),
		Cipher: c.Cipher,
		Iv:     c.Iv,
		Tag:    c.Tag,
	}

	ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

	t0 := time.Now()
	res, err := client.Decrypt(ctx, req)
	d := time.Since(t0)

	header := ctx.Value("ResponseHeader").(http.Header)
	p := profilingDataStr(header.Get("Profiling-Data"))

	return res, d, p, err
}
