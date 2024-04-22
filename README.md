# Build

First, make sure you have a recent version of the Go compiler, i.e. 1.12
or newer. You can find 1.12.10 in our toolchain repo (to use it:
`source /path/to/toolchain/shell/malbork_env`).

Then issue the following commands to build and install this tool:

```
$ make
$ make install
```

Note that there might be a bug in Go 1.14 that can cause authentication
in test-setup subcommand to fail. But the tool should work with 1.12.10
from our toolchain.
