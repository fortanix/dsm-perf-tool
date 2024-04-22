package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/spf13/cobra"
)

var versionLoadTestCmd = &cobra.Command{
	Use:   "version",
	Short: "Load test using version API",
	Long:  "Load test using version API",
	Run: func(cmd *cobra.Command, args []string) {
		versionLoadTest()
	},
}

func init() {
	loadTestCmd.AddCommand(versionLoadTestCmd)
}

func versionLoadTest() {
	setup := func(client *sdkms.Client) (interface{}, error) {
		return nil, nil
	}
	cleanup := func(client *sdkms.Client) {}
	test := func(client *sdkms.Client, stage loadTestStage, arg interface{}) (interface{}, time.Duration, profilingDataStr, error) {
		ctx := context.WithValue(context.Background(), "ResponseHeader", http.Header{})

		t0 := time.Now()
		_, err := client.Version(ctx)
		d := time.Since(t0)

		header := ctx.Value("ResponseHeader").(http.Header)
		p := profilingDataStr(header.Get("Profiling-Data"))

		return nil, d, p, err
	}
	loadTest("version", setup, test, cleanup)
}
