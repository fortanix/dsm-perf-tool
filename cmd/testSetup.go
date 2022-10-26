/* Copyright (c) Fortanix, Inc.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	DEFAULT_USER          = ""
	DEFAULT_USER_PASSWORD = ""
)

var createTestUser bool

var (
	testUserEmail    = DEFAULT_USER
	testUserPassword = DEFAULT_USER_PASSWORD
)

var testSetupCmd = &cobra.Command{
	Use:     "test-setup",
	Aliases: []string{"setup"},
	Short:   "Setup a test account useful for testing.",
	Long:    "Setup a test account useful for testing.",
	Run: func(cmd *cobra.Command, args []string) {
		testSetup()
	},
}

func init() {
	rootCmd.AddCommand(testSetupCmd)

	testSetupCmd.PersistentFlags().BoolVar(&createTestUser, "create-test-user", false, fmt.Sprintf("Create test user `%v`", testUserEmail))
	testSetupCmd.PersistentFlags().StringVar(&testUserEmail, "test-user", DEFAULT_USER, "User name for creating account")
	testSetupCmd.PersistentFlags().StringVar(&testUserPassword, "test-user-pwd", DEFAULT_USER_PASSWORD, "User password for creating account")
}

func testSetup() {
	client := sdkmsClient()
	ctx := context.Background()
	checkErr := func(action string, err error) {
		if err != nil {
			log.Fatalf("Failed to %v: %v\n", action, err)
		}
	}

	// create test user if requested
	if createTestUser {
		_, err := client.SignupUser(ctx, sdkms.SignupRequest{
			UserEmail:    testUserEmail,
			UserPassword: testUserPassword,
		})
		checkErr("create test user", err)
	}

	// authenticate as test user
	_, err := client.AuthenticateWithUserPass(ctx, testUserEmail, testUserPassword)
	checkErr("authenticate as test user", err)

	// create a test account
	acct, err := client.CreateAccount(ctx, sdkms.AccountRequest{
		Name: someString("perf-test-" + uuid.NewString()),
	})
	checkErr("create test account", err)

	// select the account
	_, err = client.SelectAccount(ctx, sdkms.SelectAccountRequest{
		AcctID: acct.AcctID,
	})
	checkErr("select account", err)

	// create a group
	group, err := client.CreateGroup(ctx, sdkms.GroupRequest{
		Name: someString("Test Group"),
	})
	checkErr("create group", err)

	// create an app and get its credential
	app_permissions := sdkms.AppPermissionsSign | sdkms.AppPermissionsVerify |
		sdkms.AppPermissionsEncrypt | sdkms.AppPermissionsDecrypt |
		sdkms.AppPermissionsWrapkey | sdkms.AppPermissionsUnwrapkey |
		sdkms.AppPermissionsDerivekey | sdkms.AppPermissionsAgreekey |
		sdkms.AppPermissionsMacgenerate | sdkms.AppPermissionsMacverify |
		sdkms.AppPermissionsExport | sdkms.AppPermissionsManage

	app, err := client.CreateApp(ctx, &sdkms.GetAppParams{}, sdkms.AppRequest{
		DefaultGroup: someString(group.GroupID),
		AddGroups:    &sdkms.AppGroups{group.GroupID: &app_permissions},
		Name:         someString("Test App"),
	})
	checkErr("create app", err)

	appCred, err := client.GetAppCredential(ctx, app.AppID)
	checkErr("get app's credential", err)

	// create an EC-NistP256 key
	ecNistP256Curve := sdkms.EllipticCurveNistP256
	ecNistP256Key, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:          someString("Test EC-NistP256 Key"),
		GroupID:       someString(group.GroupID),
		ObjType:       convertObjectType(objectTypeEC),
		EllipticCurve: &ecNistP256Curve,
	})
	checkErr("create EC-NistP256 key", err)

	// create a RSA key
	rsaKey, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test RSA Key"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeRSA),
		KeySize: someUint32(2048),
	})
	checkErr("create RSA key", err)

	// create a RSA key (4096 bits)
	rsa4096Key, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test RSA Key (4096 bits)"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeRSA),
		KeySize: someUint32(4096),
	})
	checkErr("create RSA key (4096 bits)", err)

	// create an AES key
	aesKey, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test AES Key"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeAES),
		KeySize: someUint32(256),
		Fpe: &sdkms.FpeOptions{
			Radix:     16,
			MinLength: 16,
			MaxLength: 16 * 1024,
		},
	})

	checkErr("create AES key", err)
	// create a high-volume AES key
	highVolumeAesKey, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test AES Key (high volume)"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeAES),
		KeySize: someUint32(256),
		KeyOps:  someKeyOps(sdkms.KeyOperationsEncrypt | sdkms.KeyOperationsDecrypt | sdkms.KeyOperationsWrapkey | sdkms.KeyOperationsUnwrapkey | sdkms.KeyOperationsHighvolume),
		Fpe: &sdkms.FpeOptions{
			Radix:     16,
			MinLength: 16,
			MaxLength: 16 * 1024,
		},
	})
	checkErr("create high-volume AES key", err)

	// create an AES key (192 bits)
	aes192Key, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test AES Key (192 bits)"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeAES),
		KeySize: someUint32(192),
		Fpe: &sdkms.FpeOptions{
			Radix:     16,
			MinLength: 16,
			MaxLength: 16 * 1024,
		},
	})
	checkErr("create AES key (192 bits)", err)
	// create a high-volume AES key (192 bits)
	highVolumeAes192Key, err := client.CreateSobject(ctx, sdkms.SobjectRequest{
		Name:    someString("Test AES Key (high volume) (192 bits)"),
		GroupID: someString(group.GroupID),
		ObjType: convertObjectType(objectTypeAES),
		KeySize: someUint32(192),
		KeyOps:  someKeyOps(sdkms.KeyOperationsEncrypt | sdkms.KeyOperationsDecrypt | sdkms.KeyOperationsWrapkey | sdkms.KeyOperationsUnwrapkey | sdkms.KeyOperationsHighvolume),
		Fpe: &sdkms.FpeOptions{
			Radix:     16,
			MinLength: 16,
			MaxLength: 16 * 1024,
		},
	})
	checkErr("create high-volume AES key (192 bits)", err)

	// create a few plugins
	emptyPlugin, err := createPlugin(&client, ctx, group.GroupID, "Empty", "function run(input) end")
	checkErr("create Empty plugin", err)
	helloPlugin, err := createPlugin(&client, ctx, group.GroupID, "Hello", "function run(input) return 'hello, world!' end")
	checkErr("create Hello plugin", err)
	echoPlugin, err := createPlugin(&client, ctx, group.GroupID, "Echo", "function run(input) return input end")
	checkErr("create Echo plugin", err)
	lookupKeyPlugin, err := createPlugin(&client, ctx, group.GroupID, "LookupKey", "function run(input) return assert(Sobject { name = input.key }) end")
	checkErr("create LookupKey plugin", err)
	// terminate session
	client.TerminateSession(ctx)
	fmt.Printf("export TEST_ACCT_NAME=%v\n", acct.Name)
	fmt.Printf("export TEST_ACCT_ID=%v\n", acct.AcctID)
	fmt.Printf("export TEST_GROUP_ID=%v\n", group.GroupID)
	fmt.Printf("export TEST_APP_ID=%v\n", app.AppID)
	fmt.Printf("export TEST_API_KEY=%v\n", encodeAPIKey(app, appCred))
	fmt.Printf("export TEST_RSA_KEY_ID=%v\n", *rsaKey.Kid)
	fmt.Printf("export TEST_RSA_4096_KEY_ID=%v\n", *rsa4096Key.Kid)
	fmt.Printf("export TEST_EC_NIST_P256_KEY_ID=%v\n", *ecNistP256Key.Kid)
	fmt.Printf("export TEST_AES_KEY_ID=%v\n", *aesKey.Kid)
	fmt.Printf("export TEST_HIVOL_AES_KEY_ID=%v\n", *highVolumeAesKey.Kid)
	fmt.Printf("export TEST_AES_192_KEY_ID=%v\n", *aes192Key.Kid)
	fmt.Printf("export TEST_HIVOL_AES_192_KEY_ID=%v\n", *highVolumeAes192Key.Kid)
	fmt.Printf("export TEST_EMPTY_PLUGIN_ID=%v\n", emptyPlugin.PluginID)
	fmt.Printf("export TEST_HELLO_PLUGIN_ID=%v\n", helloPlugin.PluginID)
	fmt.Printf("export TEST_ECHO_PLUGIN_ID=%v\n", echoPlugin.PluginID)
	fmt.Printf("export TEST_LOOKUP_KEY_PLUGIN_ID=%v\n", lookupKeyPlugin.PluginID)
}

func someString(s string) *string                           { return &s }
func someKeyOps(s sdkms.KeyOperations) *sdkms.KeyOperations { return &s }

func encodeAPIKey(app *sdkms.App, appCred *sdkms.AppCredentialResponse) string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", app.AppID, *appCred.Credential.Secret)))
}

func createPlugin(client *sdkms.Client, ctx context.Context, groupID string, name string, code string) (*sdkms.Plugin, error) {
	return client.CreatePlugin(ctx, sdkms.PluginRequest{
		DefaultGroup: someString(groupID),
		AddGroups:    &[]sdkms.UUID{groupID},
		Name:         someString(name),
		SourceReq: &sdkms.PluginSourceRequest{
			Inline: &sdkms.PluginSourceRequestInline{
				Language: sdkms.LanguageLua,
				Code:     code,
			},
		},
	})
}
