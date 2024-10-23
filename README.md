
# UniversumDB  🌎

![Alt Text](./docs/files/universumlogo.png)


**UniversumDB** is a high-performance key-value pair database designed for handling large number of concurrent client connections. It supports both purely in-memory storage as well as persistent LSM [(Log-structured merge tree)](https://en.wikipedia.org/wiki/Log-structured_merge-tree) based storage engine. Currently the database is expected to run on a single node where it can take advantage of multi-core CPUs, although the distribured/clustered setup is still part of the roadmap.


[![Go Report Card](https://goreportcard.com/badge/github.com/cshekharsharma/universum)](https://goreportcard.com/badge/github.com/cshekharsharma/universum)

## Features

- In-memory KV pair storage (ttl supported)
- Persistent KV pair storage (ttl supported)
- All standard commands supported (i.e, exists, get, set, delete, append, incr/decr, etc)
- Bulk Get/Set/Delete operations supported for high performance
- Manual and auto data snapshot with auto-replay (memory engine)
- Multithreaded io engine with pooling & long running connections
- Info & Statistics API for monitoring

## Supported Commands

| Command       | Description                                               |
|---------------|-----------------------------------------------------------|
| `PING`        | Check if the server is responsive.                        |
| `EXISTS`      | Determine if a key exists in the database.                |
| `GET`         | Retrieve the value associated with a key.                 |
| `SET`         | Set a value for a key with optional expiration time.      |
| `DELETE`      | Remove a key and its associated value from the database.  |
| `INCR`        | Increment the integer value of a key.                     |
| `DECR`        | Decrement the integer value of a key.                     |
| `APPEND`      | Append a value to a string key.                           |
| `MGET`        | Retrieve values for multiple keys at once.                |
| `MSET`        | Set multiple key-value pairs at once.                     |
| `MDELETE`     | Delete multiple keys at once.                             |
| `TTL`         | Get the remaining time-to-live (TTL) of a key.            |
| `EXPIRE`      | Set a timeout on a key, after which it will be deleted.   |
| `SNAPSHOT`    | Starts snapshot of all database records into a file       |
| `INFO`        | Retrieve server and database information.                 |
| `HELP`        | Prints usage and syntax details of all other commands     |

Details command syntax and request/response summary can be found at [command summary document](./docs/command-summary.md).


## Prerequisites

- Go 1.22 or higher
- *Nix systems (linux/darwin/simiar)

## Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/cshekharsharma/universum.git
   cd universum
   ```

2. **Running Tests**
   ```bash
   make test
   ```

3. **Build the project**
   ```bash
   make build-dev [or build-prod]
   ```

4. **Build and run**
   ```bash
   make all-dev [or all-prod]


## Configuration

Universum uses a toml configuration file for configuring the server behaviour, and that file can be put anywhere on the disk with readable path that will be provided to the server at startup. The recommended path for the config file is at `/etc/universum/config.toml`.
Detailed config params and their significance are explained at [config summary document](./docs/config-summary.md).


## Usage

The server listens for TCP connections on the specified port and processes client requests. Each request is handled by a worker, which reads from the connection, processes commands, and sends the response back to the client.

### Example

To test the server using `nc`:
```bash
nc localhost 11191
```

You can then send RESP3-style commands, such as:
```
SET key value
GET key
```


## Contributing

To contribute:
1. Fork the repository
2. Submit a pull request
3. Open issues for bugs or feature requests

## License

This project is licensed under the Apache 2.0 License.

---

## Roadmap

1. **Cluster Mode**: Support of running the database in distributed manner.
2. **Authentication**: Add client authentication for secure connections.
3. **RANGE Queries**: Add support of range based and aggregation queries.
4. **More Datastructures**: Add support of complex DS like lists, map, bitfields.
5. **Monitoring**: More robust monitoring support and integration with telemetry tools.


## Client Libraries

- **Go Client**: [https://github.com/cshekharsharma/universum-client-go](https://github.com/cshekharsharma/universum-client-go)

----

For any support, contact [shekharsharma705@gmail.com](mailto:shekharsharma705@gmail.com).
