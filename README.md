# TWC Gen 3

Simple monitoring server for the Tesla Wall Connector Gen 3

# Features

- Auto discover Wall Connector on network during server startup
- Record Wall Connector stats to sqlite database every second
- Quickly view Wall Connector information as line chart
- Single binary to reduce infrastructure setup

# Deployment

To run the entire application as a single binary, run:

```
% make build
% ./twc-gen3
```

This will build and start the server, which is running at `http://127.0.0.1:8080`.

# Development

Install the following dependencies

- [Go](https://go.dev/doc/install)
- [NodeJS](https://nodejs.org/en/download)

Start the server with `go build && ./twc-gen3`. The web frontend is stored in `web` and can be started by
running `npm start` from this folder.
