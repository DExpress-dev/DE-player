## Frequently Asked Questions

#### I'm getting an error while compiling Delve / unsupported architectures and OSs

The most likely cause of this is that you are running an unsupported Operating System or architecture.
Currently Delve supports (GOOS / GOARCH):
* linux / amd64 (86x64)
* linux / arm64 (AARCH64)
* linux / 386
* windows / amd64
* darwin (macOS) / amd64

There is no planned ETA for support of other architectures or operating systems. Bugs tracking requested support are:

- [32bit ARM support](https://github.com/go-delve/delve/issues/328)
- [PowerPC support](https://github.com/go-delve/delve/issues/1564)
- [OpenBSD](https://github.com/go-delve/delve/issues/1477)

See also: [backend test health](backend_test_health.md).

#### How do I use Delve with Docker?

When running the container you should pass the `--security-opt=seccomp:unconfined` option to Docker. You can start a headless instance of Delve inside the container like this:

```
dlv exec --headless --listen :4040 /path/to/executable
```

And then connect to it from outside the container:

```
dlv connect :4040
```

The program will not start executing until you connect to Delve and send the `continue` command.  If you want the program to start immediately you can do that by passing the `--continue` and `--accept-multiclient` options to Delve:

```
dlv exec --headless --continue --listen :4040 --accept-multiclient /path/to/executable
```

Note that the connection to Delve is unauthenticated and will allow arbitrary remote code execution: *do not do this in production*.

#### How can I use Delve to debug a CLI application?

There are three good ways to go about this

1. Run your CLI application in a separate terminal and then attach to it via `dlv attach`. 

1. Run Delve in headless mode via `dlv debug --headless` and then connect to it from
another terminal. This will place the process in the foreground and allow it to access
the terminal TTY.

1. Assign the process its own TTY. This can be done on UNIX systems via the `--tty` flag for the 
`dlv debug` and `dlv exec` commands. For the best experience, you should create your own PTY and 
assign it as the TTY. This can be done via [ptyme](https://github.com/derekparker/ptyme).

#### How can I use Delve for remote debugging?

It is best not to use remote debugging on a public network. If you have to do this, we recommend using ssh tunnels or a vpn connection.  

##### ```Example ``` 

Remote server:
```
dlv exec --headless --listen localhost:4040 /path/to/executable
```

Local client:
1. connect to the server and start a local port forward

```
ssh -NL 4040:localhost:4040 user@remote.ip
```

2. connect local port
```
dlv connect :4040
```
