# Universum Database Commands

The Universum database supports the following commands, along with their input and output formats. All commands return responses encoded in RESP3 format. Below are the details of the commands, including their decoded and raw RESP3 input/output.

## Command Summary

1. [`PING`](#1-ping)
2. [`EXISTS`](#2-exists)
3. [`GET`](#3-get)
4. [`SET`](#4-set)
5. [`DELETE`](#5-delete)
6. [`INCR`](#6-incr)
7. [`DECR`](#7-decr)
8. [`APPEND`](#8-append)
9. [`MGET`](#9-mget)
10. [`MSET`](#10-mset)
11. [`MDELETE`](#11-mdelete)
12. [`TTL`](#12-ttl)
13. [`EXPIRE`](#13-expire)
14. [`SNAPSHOT`](#14-snapshot)
15. [`INFO`](#15-info)
16. [`HELP`](#16-help)

---

### 1. `PING`

- **Description**: Checks if the server is responsive.
- **Input**:
    - Simplified: `PING`
    - Raw (RESP3): `"*1\r\n$4\r\nPING\r\n"`
- **Output**:
    - Simplified: `["OK", 200, ""]`
    - Raw (RESP3): `"*3\r\n$2\r\nOK\r\n:200\r\n$0\r\n"`

---

### 2. `EXISTS`

- **Description**: Checks if a key exists in the database.
- **Input**:
    - Simplified: `EXISTS key`
    - Raw (RESP3): `"*2\r\n$6\r\nEXISTS\r\n$<length>\r\n<key>\r\n"`
- **Output**:
    - Simplified: `[true/false, <code>, ""]`
    - Raw (RESP3): `"*3\r\n#t/#f\r\n:<code>\r\n$0\r\n"`

---

### 3. `GET`

- **Description**: Retrieves the value of a given key.
- **Input**:
    - Simplified: `GET key`
    - Raw (RESP3): `"*2\r\n$3\r\nGET\r\n$<length>\r\n<key>\r\n"`
- **Output**:
    - Simplified: `[value/null, <code>, ""]`
    - Raw (RESP3): `"*3\r\n$<length>/_\r\n<value>\r\n:<code>\r\n$0\r\n"`

---

### 4. `SET`

- **Description**: Sets the value of a key with an optional time-to-live (TTL).
- **Input**:
    - Simplified: `SET key value ttl`
    - Raw (RESP3): `"*4\r\n$3\r\nSET\r\n$<length>\r\n<key>\r\n$<length>\r\n<value>\r\n:<ttl>\r\n"`
- **Output**:
    - Simplified: `[true/false, <code>, ""]`
    - Raw (RESP3): `"*3\r\n#t/#f\r\n:<code>\r\n$0\r\n"`

---

### 5. `DELETE`

- **Description**: Deletes a key from the database.
- **Input**:
    - Simplified: `DELETE key`
    - Raw (RESP3): `"*2\r\n$6\r\nDELETE\r\n$<length>\r\n<key>\r\n"`
- **Output**:
    - Simplified: `[true/false, <code>, ""]`
    - Raw (RESP3): `"*3\r\n#t/#f\r\n:<code>\r\n$0\r\n"`

---

### 6. `INCR`

- **Description**: Increments the integer value of a key by a given offset.
- **Input**:
    - Simplified: `INCR key offset`
    - Raw (RESP3): `"*3\r\n$4\r\nINCR\r\n$<length>\r\n<key>\r\n:<offset>\r\n"`
- **Output**:
    - Simplified: `[new_value, <code>, ""]`
    - Raw (RESP3): `"*3\r\n:<new_value>\r\n:<code>\r\n$0\r\n"`

---

### 7. `DECR`

- **Description**: Decrements the integer value of a key by a given offset.
- **Input**:
    - Simplified: `DECR key offset`
    - Raw (RESP3): `"*3\r\n$4\r\nDECR\r\n$<length>\r\n<key>\r\n:<offset>\r\n"`
- **Output**:
    - Simplified: `[new_value, <code>, ""]`
    - Raw (RESP3): `"*3\r\n:<new_value>\r\n:<code>\r\n$0\r\n"`

---

### 8. `APPEND`

- **Description**: Appends a value to an existing string key.
- **Input**:
    - Simplified: `APPEND key value`
    - Raw (RESP3): `"*3\r\n$6\r\nAPPEND\r\n$<length>\r\n<key>\r\n$<length>\r\n<value>\r\n"`
- **Output**:
    - Simplified: `[new_length, <code>, ""]`
    - Raw (RESP3): `"*3\r\n:<new_length>\r\n:<code>\r\n$0\r\n"`

---

### 9. `MGET`

- **Description**: Retrieves the values of multiple keys.
- **Input**:
    - Simplified: `MGET [key1, key2, ...]`
    - Raw (RESP3): `"*2\r\n$4\r\nMGET\r\n*<number_of_keys>\r\n$<length>\r\n<key1>\r\n...\r\n$<length>\r\n<keyN>\r\n"`
- **Output**:
    - Simplified: `[result_map, <code>, ""]`
    - Raw (RESP3): `"*3\r\n%<number_of_keys>\r\n$<length>\r\n<key>\r\n%2\r\n$5\r\nValue\r\n$<length>/_\r\n<value>\r\n$4\r\nCode\r\n:<code>\r\n...\r\n:<code>\r\n$0\r\n"`

---

### 10. `MSET`

- **Description**: Sets multiple key-value pairs.
- **Input**:
    - Simplified: `MSET {key1: value1, key2: value2, ...}`
    - Raw (RESP3): `"*2\r\n$4\r\nMSET\r\n%<number_of_pairs>\r\n$<length>\r\n<key1>\r\n$<length>\r\n<value1>\r\n...\r\n$<length>\r\n<keyN>\r\n$<length>\r\n<valueN>\r\n"`
- **Output**:
    - Simplified: `[result_map, <code>, ""]`
    - Raw (RESP3): `"*3\r\n%<number_of_pairs>\r\n$<length>\r\n<key>\r\n#t/#f\r\n...\r\n:<code>\r\n$0\r\n"`

---

### 11. `MDELETE`

- **Description**: Deletes multiple keys from the database.
- **Input**:
    - Simplified: `MDELETE [key1, key2, ...]`
    - Raw (RESP3): `"*2\r\n$7\r\nMDELETE\r\n*<number_of_keys>\r\n$<length>\r\n<key1>\r\n...\r\n$<length>\r\n<keyN>\r\n"`
- **Output**:
    - Simplified: `[result_map, <code>, ""]`
    - Raw (RESP3): `"*3\r\n%<number_of_keys>\r\n$<length>\r\n<key>\r\n#t/#f\r\n...\r\n:<code>\r\n$0\r\n"`

---

### 12. `TTL`

- **Description**: Retrieves the time-to-live (TTL) of a key.
- **Input**:
    - Simplified: `TTL key`
    - Raw (RESP3): `"*2\r\n$3\r\nTTL\r\n$<length>\r\n<key>\r\n"`
- **Output**:
    - Simplified: `[ttl_in_seconds, <code>, ""]`
    - Raw (RESP3): `"*3\r\n:<ttl>\r\n:<code>\r\n$0\r\n"`

---

### 13. `EXPIRE`

- **Description**: Sets the TTL of a key.
- **Input**:
    - Simplified: `EXPIRE key ttl`
    - Raw (RESP3): `"*3\r\n$6\r\nEXPIRE\r\n$<length>\r\n<key>\r\n:<ttl>\r\n"`
- **Output**:
    - Simplified: `[true/false, <code>, ""]`
    - Raw (RESP3): `"*3\r\n#t/#f\r\n:<code>\r\n$0\r\n"`

---

### 14. `SNAPSHOT`

- **Description**: Initiates a snapshot of the database.
- **Input**:
    - Simplified: `SNAPSHOT`
    - Raw (RESP3): `"*1\r\n$8\r\nSNAPSHOT\r\n"`
- **Output**:
    - Simplified: `[true/false, <code>, ""]`
    - Raw (RESP3): `"*3\r\n#t/#f\r\n:<code>\r\n$0\r\n"`

---

### 15. `INFO`

- **Description**: Retrieves database information and statistics.
- **Input**:
    - Simplified: `INFO`
    - Raw (RESP3): `"*1\r\n$4\r\nINFO\r\n"`
- **Output**:
    - Simplified: `[info_string, <code>, ""]`
    - Raw (RESP3): `"*3\r\n$<length>\r\n<info_string>\r\n:<code>\r\n$0\r\n"`

---

### 16. `HELP`

- **Description**: Provides help information for commands.
- **Input**:
    - Simplified: `HELP [command_name]`
    - Raw (RESP3): 
        - Without argument: `"*1\r\n$4\r\nHELP\r\n"`
        - With argument: `"*2\r\n$4\r\nHELP\r\n$<length>\r\n<command_name>\r\n"`
- **Output**:
    - Simplified: `[help_content, <code>, ""]`
    - Raw (RESP3): `"*3\r\n$<length>\r\n<help_content>\r\n:<code>\r\n$0\r\n"`

---

## Response Code Summary

| Code  | Name                      | Description                                         |
|-------|---------------------------|-----------------------------------------------------|
| 200   | CRC_PING_SUCCESS          | Ping successful.                                    |
| 201   | CRC_SNAPSHOT_STARTED      | Snapshot initiated successfully.                    |
| 501   | CRC_SERVER_SHUTTING_DOWN  | Server is shutting down.                            |
| 502   | CRC_SERVER_BUSY           | Server is busy.                                     |
| 503   | CRC_SNAPSHOT_FAILED       | Snapshot failed.                                    |
| 1000  | CRC_RECORD_FOUND          | Record found.                                       |
| 1001  | CRC_RECORD_UPDATED        | Record updated successfully.                        |
| 1002  | CRC_RECORD_DELETED        | Record deleted successfully.                        |
| 1010  | CRC_HELP_CONTENT_OK       | Help content retrieved successfully.                |
| 1011  | CRC_INFO_CONTENT_OK       | Info content retrieved successfully.                |
| 1100  | CRC_MGET_COMPLETED        | MGET command completed successfully.                |
| 1101  | CRC_MSET_COMPLETED        | MSET command completed successfully.                |
| 1102  | CRC_MDEL_COMPLETED        | MDELETE command completed successfully.             |
| 5000  | CRC_INVALID_CMD_INPUT     | Invalid command input.                              |
| 5001  | CRC_RECORD_NOT_FOUND      | Record not found.                                   |
| 5002  | CRC_RECORD_EXPIRED        | Record has expired.                                 |
| 5003  | CRC_RECORD_NOT_DELETED    | Record not deleted.                                 |
| 5004  | CRC_INCR_INVALID_TYPE     | Invalid data type for INCR/DECR operation.          |
| 5005  | CRC_RECORD_TOO_BIG        | Record size exceeds maximum allowed size.           |
| 5006  | CRC_INVALID_DATATYPE      | Invalid data type.                                  |
| 5007  | CRC_RECORD_TOMBSTONED     | Record is tombstoned (deleted but not purged).      |
| 5010  | CRC_DATA_READ_ERROR       | Error reading data.                                 |
| 5011  | CRC_WAL_WRITE_FAILED      | Write-Ahead Log write failed.                       |

---

**Note:** Clients should check the `code` in the response to determine if the operation was successful and handle errors appropriately.