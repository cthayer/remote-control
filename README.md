remote-control
-------------------
[![Go Report Card](https://goreportcard.com/badge/github.com/cthayer/go-remote-control?style=flat-square)](https://goreportcard.com/report/github.com/cthayer/go-remote-control)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/cthayer/go-remote-control)
[![Release](https://img.shields.io/github/release/cthayer/go-remote-control.svg?style=flat-square)](https://github.com/cthayer/go-remote-control/releases/latest)

This is a service that allows command line commands to be run on the host it is running on.

It implements the [rc-protocol](https://github.com/cthayer/go-rc-protocol) which requires key based authentication before running commands.

#### Create a key pair:

To create a key pair for a client named `foo` perform the following steps:

```bash
openssl genrsa -out foo.key 4096
openssl rsa -pubout -in foo.key -out foo
```

#### Platform/Arch Support

The following platforms are supported:

```text
darwin/386
darwin/amd64
freebsd/386
freebsd/amd64
freebsd/arm
linux/386
linux/amd64
linux/arm
linux/arm64
linux/mips64
linux/mips64le
linux/mips
linux/mipsle
linux/s390x
netbsd/386
netbsd/amd64
netbsd/arm
openbsd/386
openbsd/amd64
windows/386
windows/amd64
```

Server
-------------------

`remote-control` is the server.  It is meant to be run on the host(s) where you want commands to be executed.

It is recommended to use `systemd` or your favorite init service to run the server.

Client
-------------------

`rc` is the client.  It is meant to be run from a command line and send a command to hosts running the `remote-control` service.

You can use this command line tool to run a command against many servers by piping a newline delimited list of servers to the command via STDIN.

For example:
```bash
cat hosts.txt | rc "uname -a" -c /path/to/config.json
```

You can also import the `pkg/client` and `pkg/client_config` modules into your golang project if you'd like to integrate a client directly into another project.

Installation
-------------------

### Server

* Download the `remote-control` zip archive appropriate for your system from the [releases](https://github.com/cthayer/go-remote-control/releases)
* Unzip the archive.  This will produce a binary file named `remote-control`
* (optional) Verify the file's SHA256 checksum
* (optional) Move the `remote-control` file to `/usr/local/bin` or a well-known path for executables that's platform appropriate.

**Example Systemd Unit File:**

```bash
[Unit]
Description=remote-control service
Requires=network-online.target
Wants=network-online.target
After=network-online.target

[Service]
User=root
Group=root
ExecStart=/usr/local/bin/remote-control
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
RestartSec=5s
Environment=RC_CONFIGFILE=/etc/remote-control/config.json
[Install]
WantedBy=multi-user.target
```

#### Configuration

Configuration is specified in a JSON formatted file and passed to the service using the `--config-file` command line flag or the `RC_CONFIGFILE` environment variable.

Run the following command to see full usage information:
```
remote-control --help
```

The configuration file understands the following options:

* `port`: the port to listen on (default: `4515`)
* `host`: the interface to bind to (default: `::`)
* `certDir`: the directory where authorized users' public keys are stored (default: `/etc/rc/certs`)
* `ciphers`: the list of cyphers to use
* `engineOptions`:
    * `pingTimeout`: ping timeout in ms (default: `5000`)
    * `pingInterval`: pint interval in ms (default: `1000`)
* `pidFile`: the pid file to write (default: `null`)
* `cmdOptions`: the options used by the child processes running the commands (default: `{}`).  Can be overriden by options set on the message.  See [`child_process_exec_command_options`](https://nodejs.org/api/child_process.html#child_process_child_process_exec_command_options_callback) for valid options

The server authenticates clients by requiring that they provide a signature in the `Authorization` header on the initial upgrade request.

The signature header should be in the following format:

`Authorization: RC <name>;<iso_8601_timestamp>;<signature>`

* `<name>`: the server will verify the signature using a certificate stored in `<certDir>/<name>.key` on the server
* `<iso_8601_timestamp>`: an `ISO-8601` formatted timestamp.  This is the data that has been signed
* `<sig>`: a `RSA-SHA256` signature in `base64` format

##### Environment Variables

All options from the configuration file can be passed as environment variables by prefixing the configuration file key name with `RC_` and converting to all capital letters.

### Client

* Download the `rc` zip archive appropriate for your system from the [releases](https://github.com/cthayer/go-remote-control/releases)
* Unzip the archive.  This will produce a binary file named `rc`
* (optional) Verify the file's SHA256 checksum
* (optional) Move the `rc` file to `/usr/local/bin` or a well-known path for executables that's platform appropriate.

#### Configuration

Configuration is specified in a JSON formatted file and passed to the service using the `--config-file` command line flag or the `RC_CONFIGFILE` environment variable.

Run the following command to see full usage information:
```
rc --help
```

The configuration file understands the following options:

* `port`: the port to listen on (default: `4515`)
* `host`: the interface to bind to (default: `::`)
* `keyDir`: the directory where your private key is stored
* `keyName`: the name of the file of the private key to use (without the `.key` extension)
* `logLevel`: the level of logging to display.  Can be one of: error, warn, info, debug (default: info)
* `batchSize`: the max number of servers to run the command on in parallel (default: 5)
* `delay`: the number of milliseconds to wait between batches (default: 0)
* `verbose`: set to `1` to show the raw rc-protocol response from the server(s)
* `retry`: the number of times to retry connecting to a server if the first attempt fails (default: 0)

The server authenticates clients by requiring that they provide a signature in the `Authorization` header on the initial upgrade request.

The signature header should be in the following format:

`Authorization: RC <name>;<iso_8601_timestamp>;<signature>`

* `<name>`: the server will verify the signature using a certificate stored in `<certDir>/<name>.key` on the server
* `<iso_8601_timestamp>`: an `ISO-8601` formatted timestamp.  This is the data that has been signed
* `<sig>`: a `RSA-SHA256` signature in `base64` format

##### Environment Variables

All options from the configuration file can be passed as environment variables by prefixing the configuration file key name with `RC_` and converting to all capital letters.

Testing
-------------------

[Mage](https://magefile.org/) is used for running tests.  It must be installed on the system running tests.

To run tests:
```bash
mage test
```

Building
-------------------

[Mage](https://magefile.org/) and [gox](https://github.com/mitchellh/gox) are used for building binaries.  They must be installed on the system running the build.

There are two binaries that can be built: 1) `rc`, the client binary and 2) `remote-control` the server binary.

To build `rc` (the client binary):
```bash
build/build.sh rc
```

To build `remote-control` (the server binary):
```bash
build/build.sh remote-control
```

The binaries, zip archives, and SHA256 checksum files will be written to the `build/bin` directory.

### Cleaning up

To clean up the results of a build and start fresh, run the following command:

```bash
mage clean
```

### Version

Set the version for a binary by editing the `build/versions.json` file.
