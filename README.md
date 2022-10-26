# Build

First, make sure you have a recent version of the Go compiler, i.e. 1.16
or newer.

Then issue the following commands to build and install this tool:

```
$ make
$ make install
```

Note that there might be a bug in Go 1.14 that can cause authentication
in test-setup subcommand to fail.

# Running

Usage:
  dsm-perf-tool [command]

Available Commands:
  completion  Generates bash completion scripts
  get-version Call version API in a loop
  help        Help about any command
  load-test   A collection of load tests for various types of operations.
  test-setup  Setup a test account useful for testing.

Flags:
  -h, --help                               help for dsm-perf-tool
      --idle-connection-timeout duration   Idle connection timeout, 0 means no timeout (default behavior)
      --insecure                           Do not validate server's TLS certificate
  -p, --port uint16                        DSM server port (default 443)
      --request-timeout duration           HTTP request timeout, 0 means no timeout (default 1m0s)
  -s, --server string                      DSM server host name (default "sdkms.test.fortanix.com")

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
