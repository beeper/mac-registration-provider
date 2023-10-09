# nacserv-native
A small service that generates NAC validation data on a real Mac and pushes it to a server.

## Usage
Put the binary on a Mac and run it, e.g. `./nacserv-native -url http://localhost:4000`

Flags:

* `-url` - The URL to submit data to. Required.
* `-token` - A bearer token to include when submitting data (defaults to no auth).
* `-interval` - The interval to submit data at (defaults to `5m`).

There's also a debug flag `-once`, which can be used to generate a single
validation data, print it to stdout and exit. `-url` is not required when
using `-once`.
