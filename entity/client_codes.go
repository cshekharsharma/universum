package entity

const (
	CRC_PING_SUCCESS     uint32 = 200
	CRC_SNAPSHOT_STARTED uint32 = 201

	CRC_SERVER_SHUTTING_DOWN uint32 = 501
	CRC_SERVER_BUSY          uint32 = 502
	CRC_SNAPSHOT_FAILED      uint32 = 503

	CRC_RECORD_FOUND    uint32 = 1000
	CRC_RECORD_UPDATED  uint32 = 1001
	CRC_RECORD_DELETED  uint32 = 1002
	CRC_HELP_CONTENT_OK uint32 = 1010
	CRC_INFO_CONTENT_OK uint32 = 1011

	CRC_MGET_COMPLETED uint32 = 1100
	CRC_MSET_COMPLETED uint32 = 1101
	CRC_MDEL_COMPLETED uint32 = 1102

	CRC_INVALID_CMD_INPUT  uint32 = 5000
	CRC_RECORD_NOT_FOUND   uint32 = 5001
	CRC_RECORD_EXPIRED     uint32 = 5002
	CRC_RECORD_NOT_DELETED uint32 = 5003
	CRC_INCR_INVALID_TYPE  uint32 = 5004
	CRC_RECORD_TOO_BIG     uint32 = 5005
	CRC_INVALID_DATATYPE   uint32 = 5006

	CRC_DATA_READ_ERROR  uint32 = 5010
	CRC_WAL_WRITE_FAILED uint32 = 5011
)
