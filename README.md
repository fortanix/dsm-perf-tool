# Build

First, make sure you have a recent version of the Go compiler, this tool requires go version >= `1.18`.

Then issue the following commands to build and install this tool:

```
$ make
$ make install
```

Note that there might be a bug in Go 1.14 that can cause authentication
in test-setup subcommand to fail.

# Running

Get how usage message by:

```shell
./dsm-perf-tool --help
```

## Example steps to run a performance test

1. Test setup
    You need to run `./dsm-perf-tool test-setup` to create groups, keys and plugins before run a performance test:
    - Create a account: `./dsm-perf-tool --server sdkms.test.fortanix.com test-setup --create-test-user --test-user dsm-perf-1@fortanix.com --test-user-pwd testuse1_password | tee test.env`
    - Use an existing account: `./dsm-perf-tool --server sdkms.test.fortanix.com test-setup --test-user dsm-perf-1@fortanix.com --test-user-pwd testuse1_password | tee test.env`
    
    Note:
    - This command may takes ~10 seconds.
    - `| tee test.env` is used for print and save output (multiple lines of `export XXX=abc`) to a file which could be sourced later.
    - Please update `sdkms.test.fortanix.com` to the hostname or ip address to the server you want to target.
    - You could add `--insecure` option to ignore self-signed certificate on remote host.
    - You could add `--port 123456` option to change port (default value is `443`).

2. Run a performance test
    Once you run the test setup and got a env file, such as `test.env`.
    You could start to use environment variables in the env file to run some performance test.
    
    Here is one example of running AES CBC decryption test:
    ```shell
    source test.env && \
    ./dsm-perf-tool --server sdkms.test.fortanix.com load-test --api-key $TEST_API_KEY --connections 5 --create-session --duration 10s --qps 2000  --warmup 5s symmetric-crypto --decrypt --mode CBC --kid $TEST_AES_KEY_ID
    ```
    Explanation:
    ```shell
    source test.env;  # load environment variables in env file
    ./dsm-perf-tool --server sdkms.test.fortanix.com \  # Specify remote server host name
      load-test \
      --api-key $TEST_API_KEY \  # App API key
      --connections 5 \          # Concurrent connections will be created
      --create-session \         # Will create session for each connection
      --duration 10s \           # Test duration value is a golang time string
      --qps 2000 \               # Target QPS, you could set this to a very big value if you want to test max QPS
      --warmup 5s \              # Warmup happened before test
      symmetric-crypto \
      --decrypt \                # Add this option will test decryption instead of encryption by default
      --mode CBC \
      --kid $TEST_AES_KEY_ID     # AES Key UUID, you could use $TEST_HIVOL_AES_KEY_ID to test against high volume key
    ```

    You could add `--output-format json` to print test result in JSON format.
    
    Since test result is printed in stdout and logs are printed to stderr. You could redirect the test result to a file.

    ```shell
    source test.env && \
    ./dsm-perf-tool --server sdkms.test.fortanix.com load-test --api-key $TEST_API_KEY --connections 5 --create-session --duration 10s --qps 2000  --warmup 5s symmetric-crypto --decrypt --mode CBC --kid $TEST_AES_KEY_ID | tee res.json
    ```

## Note

- All logs will are printed to stderr.
- Test summary will be printed to stdout.

# Contributing

We gratefully accept bug reports and contributions from the community.
By participating in this community, you agree to abide by [Code of Conduct](./CODE_OF_CONDUCT.md).
All contributions are covered under the Developer's Certificate of Origin (DCO).

## Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
have the right to submit it under the open source license
indicated in the file; or

(b) The contribution is based upon previous work that, to the best
of my knowledge, is covered under an appropriate open source
license and I have the right under that license to submit that
work with modifications, whether created in whole or in part
by me, under the same open source license (unless I am
permitted to submit under a different license), as indicated
in the file; or

(c) The contribution was provided directly to me by some other
person who certified (a), (b) or (c) and I have not modified
it.

(d) I understand and agree that this project and the contribution
are public and that a record of the contribution (including all
personal information I submit with it, including my sign-off) is
maintained indefinitely and may be redistributed consistent with
this project or the open source license(s) involved.

# License

This project is primarily distributed under the terms of the Mozilla Public License (MPL) 2.0, see [LICENSE](./LICENSE) for details.
