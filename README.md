# mac-registration-provider
A small service that generates NAC validation data on a Mac.

## Usage
Put the binary on a Mac and run it (`./mac-registration-provider`), optionally with some parameters:

* Relay (default) - connect to a websocket and return validation data when the server requests it.
  * `-relay-server` Use a different relay server (defaults to `https://registration-relay.beeper.com`).
* Submit - periodically generate validation data and push it to a server.
  * The list of addresses to submit to must be provided as arguments after the flags.
  * `-submit-interval` - The interval to submit data at (required).
  * `-submit-token` - A bearer token to include when submitting data (defaults to no auth).
* `-once` - generate a single validation data, print it to stdout and exit
