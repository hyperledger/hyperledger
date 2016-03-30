## Ethereum Go

Official golang implementation of the Ethereum protocol

          | Linux   | OSX | ARM | Windows | Tests
----------|---------|-----|-----|---------|------
develop   | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=Linux%20Go%20develop%20branch)](https://build.ethdev.com/builders/Linux%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=Linux%20Go%20develop%20branch)](https://build.ethdev.com/builders/OSX%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=ARM%20Go%20develop%20branch)](https://build.ethdev.com/builders/ARM%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=Windows%20Go%20develop%20branch)](https://build.ethdev.com/builders/Windows%20Go%20develop%20branch/builds/-1) | [![Buildr+Status](https://travis-ci.org/ethereum/go-ethereum.svg?branch=develop)](https://travis-ci.org/ethereum/go-ethereum) [![codecov.io](http://codecov.io/github/ethereum/go-ethereum/coverage.svg?branch=develop)](http://codecov.io/github/ethereum/go-ethereum?branch=develop)
master    | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=Linux%20Go%20master%20branch)](https://build.ethdev.com/builders/Linux%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=OSX%20Go%20master%20branch)](https://build.ethdev.com/builders/OSX%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=ARM%20Go%20master%20branch)](https://build.ethdev.com/builders/ARM%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.ethdev.com/buildstatusimage?builder=Windows%20Go%20master%20branch)](https://build.ethdev.com/builders/Windows%20Go%20master%20branch/builds/-1) | [![Buildr+Status](https://travis-ci.org/ethereum/go-ethereum.svg?branch=master)](https://travis-ci.org/ethereum/go-ethereum) [![codecov.io](http://codecov.io/github/ethereum/go-ethereum/coverage.svg?branch=master)](http://codecov.io/github/ethereum/go-ethereum?branch=master)

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/ethereum/go-ethereum) 
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ethereum/go-ethereum?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

## Automated development builds

The following builds are build automatically by our build servers after each push to the [develop](https://github.com/ethereum/go-ethereum/tree/develop) branch.

* [Docker](https://registry.hub.docker.com/u/ethereum/client-go/)
* [OS X](http://build.ethdev.com/builds/OSX%20Go%20develop%20branch/Mist-OSX-latest.dmg)
* Ubuntu
  [trusty](https://build.ethdev.com/builds/Linux%20Go%20develop%20deb%20i386-trusty/latest/) |
  [utopic](https://build.ethdev.com/builds/Linux%20Go%20develop%20deb%20i386-utopic/latest/)
* [Windows 64-bit](https://build.ethdev.com/builds/Windows%20Go%20develop%20branch/Geth-Win64-latest.zip)
* [ARM](https://build.ethdev.com/builds/ARM%20Go%20develop%20branch/geth-ARM-latest.tar.bz2)

## Building the source

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/ethereum/go-ethereum/wiki/Building-Ethereum)
on the wiki.

Building geth requires both a Go and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make geth

## Executables

Go Ethereum comes with several wrappers/executables found in 
[the `cmd` directory](https://github.com/ethereum/go-ethereum/tree/develop/cmd):

 Command  |         |
----------|---------|
`geth` | Ethereum CLI (ethereum command line interface client) |
`bootnode` | runs a bootstrap node for the Discovery Protocol |
`ethtest` | test tool which runs with the [tests](https://github.com/ethereum/tests) suite: `/path/to/test.json > ethtest --test BlockTests --stdin`.
`evm` | is a generic Ethereum Virtual Machine: `evm -code 60ff60ff -gas 10000 -price 0 -dump`. See `-h` for a detailed description. |
`disasm` | disassembles EVM code: `echo "6001" | disasm` |
`rlpdump` | prints RLP structures |

## Command line options

`geth` can be configured via command line options, environment variables and config files.

To get the options available:

    geth help

For further details on options, see the [wiki](https://github.com/ethereum/go-ethereum/wiki/Command-Line-Options)

## Contribution

If you'd like to contribute to go-ethereum please fork, fix, commit and
send a pull request. Commits who do not comply with the coding standards
are ignored (use gofmt!). If you send pull requests make absolute sure that you
commit on the `develop` branch and that you do not merge to master.
Commits that are directly based on master are simply ignored.

See [Developers' Guide](https://github.com/ethereum/go-ethereum/wiki/Developers'-Guide)
for more details on configuring your environment, testing, and
dependency management.
