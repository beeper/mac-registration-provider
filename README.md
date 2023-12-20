# mac-registration-provider
A small service that generates iMessage registration data on a Mac.

## Supported OS versions
The tool is currently quite hacky, so it only works on specific versions of macOS.

* Intel: 11.5 - 11.7, 14.1 - 14.2
* Apple Silicon: 13.5 - 13.6, 14.0 - 14.2

On unsupported versions, it will tell you that it's unsupported and exit.
A future version may work in less hacky ways to support more OS versions.

## Usage
Put the binary on a Mac and run it (`./mac-registration-provider`), optionally with some parameters:

* Relay (default) - connect to a websocket and return validation data when the server requests it.
  * `-relay-server` Use a different relay server (defaults to `https://registration-relay.beeper.com`).
* Submit - periodically generate validation data and push it to a server.
  * The list of addresses to submit to must be provided as arguments after the flags.
  * `-submit-interval` - The interval to submit data at (required).
  * `-submit-token` - A bearer token to include when submitting data (defaults to no auth).
* `-once` - generate a single validation data, print it to stdout and exit
