# Proxy

Basic HTTP proxy in Go. It has a simple filter that if the HTTP message contains a certain word it will unauthorize the request.

## Features

- Docker
- Syslog logging
- Checksum on build
- Concurrency with go routines

## Usage

To run the logging server:

```bash
$ docker compose up -d
```

To run the proxy server:

```bash
$ scripts/run.sh
```

To regenarate the hashes for all files:

```bash
$ scripts/checksum.sh > out.sha256sum
```

### Why not build the entire project inside the docker compose file?

I needed to log to the server in case of a fail on the build process, so the logging server would have to be up before the image building started. I tried a simple `depends_on` but it didn't work, so I landed on the current solution.
