package server

// various states of server lifecycle
const STATE_STARTING int32 = 1
const STATE_READY int32 = 1 << 2
const STATE_BUSY int32 = 1 << 4
const STATE_SHUTTING_DOWN int32 = 1 << 8
const STATE_STOPPED int = 1 << 16
