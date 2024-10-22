
# UniversumDB  🌎

![Alt Text](./docs/universumlogo.png)


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

## Prerequisites

- Go 1.22 or higher
- Properly configured environment variables for port and connection limits

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
   make build-dev [or build-dev]
   ```

4. **Build and run**
   ```bash
   make all-dev [or all-prod]


## Configuration

Sample configuration file that can be put in /etc/universum/config.toml

```toml
[Server]
ServerPort = 11191
MaxConnections = 100
ConnectionWriteTimeout = 10
RequestExecutionTimeout = 10

[Storage]
StorageEngine = "LSM"
MaxRecordSizeInBytes = 1048576

[Storage.Memory]
AllowedMemoryStorageLimit = 1073741824
AutoSnapshotFrequency = 100
SnapshotFileDirectory = "/opt/universum/snapshot"
SnapshotCompressionAlgo = "LZ4"
RestoreSnapshotOnStart = true

[Storage.LSM]
MemtableStorageType = "LB"
BloomFilterMaxRecords = 100000
WriteBufferSize = 4194304
BloomFalsePositiveRate = 0.01
WriteBlockSize = 65536
BlockCompressionAlgo = "LZ4"
DataStorageDirectory = "/opt/universum/data"
WriteAheadLogDirectory = "/opt/universum/wal"
WriteAheadLogAsyncFlush = true
WriteAheadLogFrequency = 5
WriteAheadLogBufferSize = 1048576

[Logging]
LogFileDirectory = "/var/log/universum"
MinimumLogLevel = "INFO"

[Eviction]
AutoRecordExpiryFrequency = 2
AutoEvictionPolicy = "LRU"

[Auth]
AuthenticationEnabled = false
DbUserName = ""
DbUserPassword = ""


```

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
