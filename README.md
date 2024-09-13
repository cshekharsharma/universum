
# UniversumDB  🌎

**UniversumDB** is a high-performance key-value pair database designed for handling large number of concurrent client connections. It supports an in-memory key-value database with functionalities like auto-snapshot, record replay on start etc. 



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

```ini
[server]
ServerPort=11191
MaxConnections=100
ConnectionWriteTimeout=10
RequestExecutionTimeout=10

[storage]
StorageEngine="MEMORY"
MaxRecordSizeInBytes=1048576

[storage.memory]
AllowedMemoryStorageLimit=1073741824

[snapshot]
AutoSnapshotFrequency=100
TransactionLogFilePath=/opt/universum/translog.aof

[logging]
ServerLogFilePath=/var/log/universum/server.log
MinimumLogLevel=INFO

[eviction]
AutoRecordExpiryFrequency=2
RecordAutoEvictionPolicy=LRU

[auth]
AuthenticationEnabled=1
DbUserName=admin
DbUserPassword=admin

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
