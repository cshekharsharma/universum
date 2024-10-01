
# UniversumDB  🌎

**UniversumDB** is a high-performance key-value pair database designed for handling large number of concurrent client connections. It supports an in-memory key-value database with functionalities like auto-snapshot, record replay on start etc.


[![Go Report Card](https://goreportcard.com/badge/github.com/cshekharsharma/universum)](https://goreportcard.com/badge/github.com/cshekharsharma/universum)

## Features

- Key-Value pair storage (ttl supported)
- Manual and auto data snapshot
- Auto replay of data on startup
- Info & Statistics API
- Client connection throttling for high performance

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

Sample configuration file that can be put in /etc/universum/config.ini

```toml
[Server]
ServerPort = 11191
MaxConnections = 100
ConnectionWriteTimeout = 10
RequestExecutionTimeout = 10

[Storage]
StorageEngine = "MEMORY"
MaxRecordSizeInBytes = 1048576

[Storage.Memory]
AllowedMemoryStorageLimit = 1073741824
AutoSnapshotFrequency = 100
SnapshotFileDirectory = "/opt/universum/snapshot"
SnapshotCompressionAlgo = "LZ4"
RestoreSnapshotOnStart = true

[Storage.LSM]
MemtableStorageType = "LB"
WriteBlockSize = 65536
DataStorageDirectory = "/opt/universum/data"
WriteAheadLogDirectory = "/opt/universum/wal"
WriteAheadLogFrequency = 5
WriteAheadLogBufferSize = 1048576

[Logging]
LogFileDirectory = "/var/log/universum"
MinimumLogLevel = "INFO"

[Eviction]
AutoRecordExpiryFrequency = 2
AutoEvictionPolicy = "LRU"

[Auth]
AuthenticationEnabled = true
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

## Improvements & Feedback

Areas of potential improvement:
1. **Dynamic Scaling**: Consider dynamically scaling workers based on connection load.
2. **Authentication**: Add client authentication for secure connections.
3. **Auto-Eviction**: LRU based auto eviction in case of memory overflow


## Client Libraries

- **Go Client**: [https://github.com/cshekharsharma/universum-client-go](https://github.com/cshekharsharma/universum-client-go)

----

For any support, contact [shekharsharma705@gmail.com](mailto:shekharsharma705@gmail.com).
