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

Clients
-------------------

[rc-client](https://github.com/sazze/node-rc-client) is the official client library for interacting with this service.

Installation
-------------------
**RPM:**

`yum install remote-control`

*Note: creation of an RPM package is still on the TODO list

**Other:**

* Download the `remote-control` binary file appropriate for your system (linux/MacOS(darwin)) from the [releases](https://github.com/cthayer/go-remote-control/releases)
* Make the downloaded file executable

**Example Init Script (RedHat/Centos):**

```bash
#!/bin/bash
#
#	/etc/rc.d/init.d/rc
#
#	This is a service that allows command line commands to be run on the host it is running on.
#
# chkconfig: 2345 20 80
# description: This is a service that allows command line commands to be run on the host it is running on.
# processname: rc
# config: /etc/rc/config.json
# pidfile: /var/run/rc.pid
#

# Source function library.
. /etc/init.d/functions

PROGNAME=rc
PROG=/usr/local/sbin/$PROGNAME
PIDFILE=/var/run/${PROGNAME}.pid
CONFIGFILE=/etc/$PROGNAME/config.json
LOCKFILE=/var/lock/subsys/$PROGNAME

start() {
    if [ -d /.nar ]; then
        # ensure clean start (nar agressively caches code)
        rm -rf /.nar
    fi
    
	echo -n "Starting $PROGNAME: "
	daemon $PROG --configFile $CONFIGFILE
	ret=$?
	echo ""
	touch $LOCKFILE
	return $ret
}

stop() {
	echo -n "Shutting down $PROGNAME: "
	killproc -p $PIDFILE $PROG
	ret=$?
	echo ""
	rm -f $LOCKFILE
	return $ret
}

case "$1" in
    start)
	start
	;;
    stop)
	stop
	;;
    status)
	status -p $PIDFILE -l $LOCKFILE
	;;
    restart)
    stop
	start
	;;
    *)
	echo "Usage: <servicename> {start|stop|status|reload|restart[|probe]"
	exit 1
	;;
esac
exit $?
```

**Example Systemd Unit:**

```bash
TODO
```

Configuration
-------------------
Configuration is specified in a JSON formated file and passed to the service using the `-configFile` command line flag.

```
remote-control [-configFile FILE] [-version]
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

* `<name>`: the server will verfiy the signature using a certificate stored in `<certDir>/<name>.key` on the server
* `<iso_8601_timestamp>`: an `ISO-8601` formatted timestamp.  This is the data that has been signed
* `<sig>`: a `RSA-SHA256` signature in `base64` format

### Environment Variables

* `DEBUG`: sets the level(s) for logging (default: `error,warn,info`).  `main` and `server` levels are also available for verbose logging.

Testing
-------------------

You can test client actions with `test/scripts/client.js`

**NOTES:** 

* set the `RC_CERT_NAME` and `RC_CERT_DIR` environment variables before running the client test script.
* place your public key in `/tmp/rcPubKeys`
* run the server: `node app.js --configFile ./test/config/config.json`

### Environment Variables

* `RC_CERT_NAME`: the name of the cert for the client to use for authorization (see `name` in the config options)
* `RC_CERT_DIR`: the path to the directory that contains the cert
