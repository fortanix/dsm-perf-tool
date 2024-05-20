/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/stretchr/testify/assert"
)

func TestTestSetupCmd(t *testing.T) {
	resetRootCmdStatus()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sys/v1/version" {
			t.Errorf("Expected to request '/sys/v1/version', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		resp := sdkms.VersionResponse{
			Version:    "1234",
			APIVersion: "1234",
			ServerMode: "httptest",
		}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			t.Fatal("Error encoding fake version response:", err)
		}
		w.Write(respBytes)
	}))
	defer server.Close()
	// Parse the URL
	parsedServerURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal("Error parsing test server URL:", err)
	}
	testServerHost := parsedServerURL.Hostname()
	testServerPort := parsedServerURL.Port()
	if testServerPort == "" {
		testServerPort = "443"
	}

	args := []string{
		"--server", testServerHost,
		"--port", testServerPort,
		"--insecure",
		"version",
		"--count", "1",
	}

	rootCmd.SetArgs(args)
	err = ExecuteCmd()
	assert.NoError(t, err)
}
