# Universum Database Configuration Guide

This document provides detailed explanations of the configuration settings available for the Universum database. Each setting is described with its purpose, default value, and any relevant notes.

By default the configuration file is expected to be placed at `/etc/universum/config.toml`. But you can place it anywhere and provide the path in the server startup command.

---

## Table of Contents

- [Server](#server)
- [Cluster](#cluster)
- [Storage](#storage)
  - [Storage.Memory](#storagememory)
  - [Storage.LSM](#storagelsm)
- [Logging](#logging)
- [Eviction](#eviction)
- [Auth](#auth)

---

## [Server]

Settings related to the server's network configuration and connection handling.

##### `ServerPort`

- **Description:** The port number on which the server listens for incoming connections.
- **Default Value:** `11191`
- **Example:** `ServerPort = 11191`

##### `MaxConnections`

- **Description:** The maximum number of concurrent client connections allowed.
- **Default Value:** `10000`
- **Example:** `MaxConnections = 10000`

##### `ConnectionWriteTimeout`

- **Description:** The maximum duration (in seconds) the server waits for a client to send data before timing out.
- **Default Value:** `10`
- **Example:** `ConnectionWriteTimeout = 10`

##### `RequestExecutionTimeout`

- **Description:** The maximum duration (in seconds) the server allows for a request to be processed.
- **Default Value:** `10`
- **Example:** `RequestExecutionTimeout = 10`

##### `EnableTLS`

- **Description:** Enables or disables TLS encryption for client connections.
- **Default Value:** `false`
- **Example:** `EnableTLS = false`

##### `TLSCertFilePath`

- **Description:** The file path to the TLS certificate file.
- **Default Value:** `"/etc/universum/cert.pem"`
- **Example:** `TLSCertFilePath = "/etc/universum/cert.pem"`

##### `TLSKeyFilePath`

- **Description:** The file path to the TLS private key file.
- **Default Value:** `"/etc/universum/key.pem"`
- **Example:** `TLSKeyFilePath = "/etc/universum/key.pem"`

---

## [Cluster]

Settings for configuring clustering and replication among multiple database instances.

##### `EnableCluster`

- **Description:** Enables or disables clustering features for distributed database capabilties.
- **Default Value:** `false`
- **Example:** `EnableCluster = false`

##### `HeartbeatPort`

- **Description:** The port used for cluster heartbeat messages.
- **Default Value:** `11192`
- **Example:** `HeartbeatPort = 11192`

##### `Hosts`

- **Description:** A list of IP addresses or hostnames of cluster nodes.
- **Default Value:** `["127.0.0.1"]`
- **Example:** `Hosts = ["127.0.0.1"]`

##### `GossipIntervalMs`

- **Description:** The interval in milliseconds between gossip messages.
- **Default Value:** `1000`
- **Example:** `GossipIntervalMs = 1000`

##### `ReplicationFactor`

- **Description:** The number of replicas for each piece of data in the cluster.
- **Default Value:** `2`
- **Example:** `ReplicationFactor = 2`

---

## [Storage]

General storage settings.

##### `StorageEngine`

- **Description:** The type of storage engine to use. Options include `"LSM"` (Log-Structured Merge Tree) or `"MEMORY"`.
- **Default Value:** `"MEMORY"`
- **Example:** `StorageEngine = "MEMORY"`

##### `MaxRecordSizeInBytes`

- **Description:** The maximum allowed size for a single record, in bytes.
- **Default Value:** `1048576` (1 MB)
- **Example:** `MaxRecordSizeInBytes = 1048576`

---

### [Storage.Memory]

Settings specific to the in-memory storage engine.

###### `AllowedMemoryStorageLimit`

- **Description:** The maximum amount of memory (in bytes) allocated for in-memory storage.
- **Default Value:** `1073741824` (1 GB)
- **Example:** `AllowedMemoryStorageLimit = 1073741824`

###### `AutoSnapshotFrequency`

- **Description:** The frequency (in seconds) at which the database automatically takes snapshots.
- **Default Value:** `100`
- **Example:** `AutoSnapshotFrequency = 100`

###### `SnapshotFileDirectory`

- **Description:** The directory where snapshot files are stored.
- **Default Value:** `"/opt/universum/snapshot"`
- **Example:** `SnapshotFileDirectory = "/opt/universum/snapshot"`

###### `SnapshotCompressionAlgo`

- **Description:** The compression algorithm used for snapshots. Options might include `"LZ4"`, `"NONE"`, etc. `NONE` value turns off the compression.
- **Default Value:** `"LZ4"`
- **Example:** `SnapshotCompressionAlgo = "LZ4"`

###### `RestoreSnapshotOnStart`

- **Description:** Determines whether the database restores from a snapshot upon startup.
- **Default Value:** `true`
- **Example:** `RestoreSnapshotOnStart = true`

---

### [Storage.LSM]

Settings specific to the LSM (Log-Structured Merge Tree) storage engine.

###### `MemtableStorageType`

- **Description:** The type of storage for the memtable. Options might include `"LB"` (SkipList + Bloom filter), `"TB"` (RedBlack tree + Bloom filter), etc.
- **Default Value:** `"LB"`
- **Example:** `MemtableStorageType = "LB"`

###### `BloomFilterMaxRecords`

- **Description:** The maximum number of records for which the bloom filter is optimized.
- **Default Value:** `100000`
- **Example:** `BloomFilterMaxRecords = 100000`

###### `WriteBufferSize`

- **Description:** The size (in bytes) of the write buffer before flushing to disk.
- **Default Value:** `67108864` (64 MB)
- **Example:** `WriteBufferSize = 67108864`

###### `BloomFalsePositiveRate`

- **Description:** The acceptable false positive rate for the bloom filter.
- **Default Value:** `0.01`
- **Example:** `BloomFalsePositiveRate = 0.01`

###### `WriteBlockSize`

- **Description:** The size (in bytes) of blocks when writing to disk.
- **Default Value:** `65536` (64 KB)
- **Example:** `WriteBlockSize = 65536`

###### `BlockCompressionAlgo`

- **Description:** The compression algorithm used for blocks. Options might include `"LZ4"`, `"NONW"`, etc. `NONE` value turns off the compression.
- **Default Value:** `"LZ4"`
- **Example:** `BlockCompressionAlgo = "LZ4"`

###### `DataStorageDirectory`

- **Description:** The directory where compressed/uncompressed data files (sstables) are stored.
- **Default Value:** `"/opt/universum/data"`
- **Example:** `DataStorageDirectory = "/opt/universum/data"`

###### `WriteAheadLogDirectory`

- **Description:** The directory where Write-Ahead Log (WAL) files are stored.
- **Default Value:** `"/opt/universum/wal"`
- **Example:** `WriteAheadLogDirectory = "/opt/universum/wal"`

###### `WriteAheadLogAsyncFlush`

- **Description:** Enables or disables asynchronous flushing of the WAL.
- **Default Value:** `false`
- **Example:** `WriteAheadLogAsyncFlush = false`

###### `WriteAheadLogFrequency`

- **Description:** The frequency (in seconds) at which the WAL is flushed to disk.
- **Default Value:** `5`
- **Example:** `WriteAheadLogFrequency = 5`

###### `WriteAheadLogBufferSize`

- **Description:** The size (in bytes) of the WAL buffer before flushing.
- **Default Value:** `1048576` (1 MB)
- **Example:** `WriteAheadLogBufferSize = 1048576`

###### `BlockCacheMemoryLimit`

- **Description:** The maximum amount of memory (in bytes) allocated for the block cache.
- **Default Value:** `1073741824` (1 GB)
- **Example:** `BlockCacheMemoryLimit = 1073741824`

---

## [Logging]

Settings related to logging behavior.

##### `LogFileDirectory`

- **Description:** The directory where log files are stored.
- **Default Value:** `"/var/log/universum"`
- **Example:** `LogFileDirectory = "/var/log/universum"`

##### `MinimumLogLevel`

- **Description:** The minimum level of logs to record. Options include `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"`, `"FATAL"`.
- **Default Value:** `"INFO"`
- **Example:** `MinimumLogLevel = "INFO"`

---

## [Eviction]

Settings that control record eviction policies and frequencies.

##### `AutoRecordExpiryFrequency`

- **Description:** The frequency (in seconds) at which expired records are automatically removed.
- **Default Value:** `5`
- **Example:** `AutoRecordExpiryFrequency = 5`

##### `AutoEvictionPolicy`

- **Description:** The policy used for evicting records when memory limits are reached. Options might include `"LRU"` (Least Recently Used), `"LFU"` (Least Frequently Used), `"NONE"` (No eviction) etc.
- **Default Value:** `"LRU"`
- **Example:** `AutoEvictionPolicy = "LRU"`

---

## [Auth]

Settings for authentication and access control.

##### `AuthenticationEnabled`

- **Description:** Enables or disables authentication for client connections.
- **Default Value:** `false`
- **Example:** `AuthenticationEnabled = false`

##### `DbUserName`

- **Description:** The username required for authentication.
- **Default Value:** `"admin"`
- **Example:** `DbUserName = "admin"`

##### `DbUserPassword`

- **Description:** The password required for authentication.
- **Default Value:** `"admin"`
- **Example:** `DbUserPassword = "admin"`

---

## Notes

- Ensure that all file paths specified in the configuration exist and have the appropriate permissions.
- When `EnableTLS` is set to `true`, make sure that the `TLSCertFilePath` and `TLSKeyFilePath` are correctly configured and point to valid TLS certificate and key files.
- Adjust `MaxConnections` and memory limits (`AllowedMemoryStorageLimit`, `BlockCacheMemoryLimit`) based on the resources available on your server to prevent performance issues.
- The `ReplicationFactor` in a cluster should not exceed the number of nodes in the cluster.

---

## Example Configuration File

Below is an example of a complete configuration file with default settings:

```toml
[Server]
ServerPort = 11191
MaxConnections = 100
ConnectionWriteTimeout = 10
RequestExecutionTimeout = 10
EnableTLS = false
TLSCertFilePath = "/etc/universum/cert.pem"
TLSKeyFilePath = "/etc/universum/key.pem"

[Cluster]
EnableCluster = false
HeartbeatPort = 11192
Hosts = ["127.0.0.1"]
GossipIntervalMs = 1000
ReplicationFactor = 2

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
WriteAheadLogAsyncFlush = false
WriteAheadLogFrequency = 5
WriteAheadLogBufferSize = 1048576
BlockCacheMemoryLimit = 1048576

[Logging]
LogFileDirectory = "/var/log/universum"
MinimumLogLevel = "INFO"

[Eviction]
AutoRecordExpiryFrequency = 2
AutoEvictionPolicy = "LRU"

[Auth]
AuthenticationEnabled = true
DbUserName = "admin"
DbUserPassword = "admin"
```