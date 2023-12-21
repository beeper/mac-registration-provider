# mac-registration-provider
A small service that generates iMessage registration data on a Mac. If you do not have access to Beeper Cloud, you can use this to generate a iMessage Registration Code and use it in Beeper Mini. 

## Supported MacOS versions
The tool is currently quite hacky, so it only works on specific versions of macOS.

* Intel: 11.5 - 11.7, 14.0 - 14.3
* Apple Silicon: 13.5 - 13.6, 14.0 - 14.3

On unsupported versions, it will tell you that it's unsupported and exit.
A future version may work in less hacky ways to support more OS versions.

## Usage
1. On your Mac, download the latest `mac-registration-provider` file from the latest [release](https://github.com/beeper/mac-registration-provider/releases)
![CleanShot 2023-12-21 at 14 32 42@2x](https://github.com/beeper/mac-registration-provider/assets/1048265/4a419ae1-8996-4af4-876e-5723db088816)
2. Open Terminal app (⌘+space -> Terminal), type `cd Downloads`, hit enter
3. type `chmod +x mac-registration-provider`, hit enter
4. Type `./mac-registration-provider`, hit enter


## Future improvements
If anyone wants to package this into an app that lives in your dock and runs at startup, we'd appreciate it!

## Optional parameters:

* Relay (default) - connect to a websocket and return registration data when the server requests it.
  * `-relay-server` Use a different relay server (defaults to `https://registration-relay.beeper.com`).
* Submit - periodically generate registration data and push it to a server.
  * The list of addresses to submit to must be provided as arguments after the flags.
  * `-submit-interval` - The interval to submit data at (required).
  * `-submit-token` - A bearer token to include when submitting data (defaults to no auth).
* `-once` - generate a single registration data, print it to stdout and exit
