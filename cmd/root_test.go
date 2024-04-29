/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetRootCmdStatus() {
	rootCmd.SetArgs([]string{})
	serverName = defaultServerName
	serverPort = defaultServerPort
	insecureTLS = defaultTlsNotVerify
	requestTimeout = defaultRequestTimeout
	idleConnectionTimeout = defaultIdleConnectionTimeout
}

func TestRootCmdDefaultFlags(t *testing.T) {
	resetRootCmdStatus()
	err := ExecuteCmd()
	assert.NoError(t, err)
	assert.Equal(t, defaultServerName, serverName)
	assert.Equal(t, defaultServerPort, serverPort)
	assert.Equal(t, defaultTlsNotVerify, insecureTLS)
	assert.Equal(t, defaultRequestTimeout, requestTimeout)
	assert.Equal(t, defaultIdleConnectionTimeout, idleConnectionTimeout)
}

func TestRootCmdWithCustomArgs(t *testing.T) {
	resetRootCmdStatus()
	customServerName := "sdkms.custom.fortanix.com"
	customServerPort := uint16(8080)
	customInsecureTLS := true
	customRequestTimeout := 120 * time.Second
	customIdleConnectionTimeout := 300 * time.Second

	rootCmd.SetArgs([]string{
		"--server", customServerName,
		"--port", fmt.Sprint(customServerPort),
		"--insecure",
		"--request-timeout", fmt.Sprint(customRequestTimeout),
		"--idle-connection-timeout", fmt.Sprint(customIdleConnectionTimeout),
	})

	err := ExecuteCmd()

	assert.NoError(t, err)
	assert.Equal(t, customServerName, serverName)
	assert.Equal(t, customServerPort, serverPort)
	assert.Equal(t, customInsecureTLS, insecureTLS)
	assert.Equal(t, customRequestTimeout, requestTimeout)
	assert.Equal(t, customIdleConnectionTimeout, idleConnectionTimeout)
}

func TestRootCmdWithCustomShortArgs(t *testing.T) {
	resetRootCmdStatus()

	customServerName := "sdkms.custom.fortanix.com"
	customServerPort := uint16(8080)

	rootCmd.SetArgs([]string{
		"-s", customServerName,
		"-p", fmt.Sprint(customServerPort),
	})

	err := ExecuteCmd()

	assert.NoError(t, err)
	assert.Equal(t, customServerName, serverName)
	assert.Equal(t, customServerPort, serverPort)
	assert.Equal(t, defaultTlsNotVerify, insecureTLS)
	assert.Equal(t, defaultRequestTimeout, requestTimeout)
	assert.Equal(t, defaultIdleConnectionTimeout, idleConnectionTimeout)
}

func TestRootCmdWithInvalidArgs(t *testing.T) {
	resetRootCmdStatus()

	invalidServerPort := "invalid-server-port"
	invalidRequestTimeout := "invalid-request-timeout"
	invalidIdleConnectionTimeout := "invalid-idle-connection-timeout"
	var err error

	rootCmd.SetArgs([]string{
		"--port", invalidServerPort,
	})
	err = ExecuteCmd()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")

	rootCmd.SetArgs([]string{
		"--request-timeout", invalidRequestTimeout,
	})
	err = ExecuteCmd()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
	assert.Contains(t, err.Error(), "invalid duration")

	rootCmd.SetArgs([]string{
		"--idle-connection-timeout", invalidIdleConnectionTimeout,
	})
	err = ExecuteCmd()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
	assert.Contains(t, err.Error(), "invalid duration")
}
